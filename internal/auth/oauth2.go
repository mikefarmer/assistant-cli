package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

// OAuth2Provider implements authentication using OAuth2 flow with browser
type OAuth2Provider struct {
	clientID     string
	clientSecret string
	tokenFile    string
	config       *oauth2.Config
	token        *oauth2.Token
	client       *texttospeech.Client
}

// NewOAuth2Provider creates a new OAuth2 authentication provider
func NewOAuth2Provider(clientID, clientSecret, tokenFile string) *OAuth2Provider {
	// If no parameters provided, try to get from environment
	if clientID == "" {
		clientID = os.Getenv("ASSISTANT_CLI_OAUTH2_CLIENT_ID")
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("ASSISTANT_CLI_OAUTH2_CLIENT_SECRET")
	}
	if tokenFile == "" {
		tokenFile = os.Getenv("ASSISTANT_CLI_OAUTH2_TOKEN_FILE")
		if tokenFile == "" {
			// Default token file location
			home, _ := os.UserHomeDir()
			tokenFile = filepath.Join(home, ".assistant-cli-oauth2-token.json")
		}
	}

	provider := &OAuth2Provider{
		clientID:     clientID,
		clientSecret: clientSecret,
		tokenFile:    tokenFile,
	}

	if provider.isOAuth2Configured() {
		provider.config = &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  "http://localhost:8080/callback",
			Scopes: []string{
				"https://www.googleapis.com/auth/cloud-platform",
			},
			Endpoint: google.Endpoint,
		}
	}

	return provider
}

// GetClient returns a Google Cloud TTS client configured with OAuth2 authentication
func (p *OAuth2Provider) GetClient(ctx context.Context) (*texttospeech.Client, error) {
	if p.client != nil {
		return p.client, nil
	}

	if !p.IsConfigured() {
		return nil, fmt.Errorf("OAuth2 is not configured")
	}

	// Get valid token
	token, err := p.getValidToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get valid token: %w", err)
	}

	// Create HTTP client with OAuth2 token
	httpClient := p.config.Client(ctx, token)

	// Create TTS client with OAuth2 HTTP client
	client, err := texttospeech.NewClient(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create TTS client with OAuth2: %w", err)
	}

	p.client = client
	return p.client, nil
}

// IsConfigured returns true if OAuth2 is properly configured
func (p *OAuth2Provider) IsConfigured() bool {
	return p.isOAuth2Configured() && p.hasValidToken()
}

// GetMethod returns the authentication method
func (p *OAuth2Provider) GetMethod() AuthMethod {
	return AuthMethodOAuth2
}

// Authenticate performs OAuth2 authentication flow
func (p *OAuth2Provider) Authenticate(ctx context.Context) error {
	if !p.isOAuth2Configured() {
		return fmt.Errorf("OAuth2 client ID and secret are not configured. " +
			"Set ASSISTANT_CLI_OAUTH2_CLIENT_ID and ASSISTANT_CLI_OAUTH2_CLIENT_SECRET environment variables")
	}

	// Try to load existing token
	if err := p.loadToken(); err == nil && p.token.Valid() {
		return nil // Token is already valid
	}

	// If token exists but is expired, try to refresh it
	if p.token != nil && !p.token.Valid() {
		if refreshed, err := p.refreshToken(ctx); err == nil {
			p.token = refreshed
			return p.saveToken()
		}
	}

	// Perform full OAuth2 flow
	return p.performOAuth2Flow(ctx)
}

// isOAuth2Configured checks if OAuth2 client credentials are available
func (p *OAuth2Provider) isOAuth2Configured() bool {
	return p.clientID != "" && p.clientSecret != ""
}

// hasValidToken checks if we have a valid OAuth2 token
func (p *OAuth2Provider) hasValidToken() bool {
	if err := p.loadToken(); err != nil {
		return false
	}
	return p.token != nil && p.token.Valid()
}

// loadToken loads the OAuth2 token from file
func (p *OAuth2Provider) loadToken() error {
	if p.token != nil {
		return nil // Already loaded
	}

	data, err := os.ReadFile(p.tokenFile)
	if err != nil {
		return err
	}

	token := &oauth2.Token{}
	if err := json.Unmarshal(data, token); err != nil {
		return err
	}

	p.token = token
	return nil
}

// saveToken saves the OAuth2 token to file
func (p *OAuth2Provider) saveToken() error {
	if p.token == nil {
		return fmt.Errorf("no token to save")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(p.tokenFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(p.token, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(p.tokenFile, data, 0600)
}

// getValidToken returns a valid OAuth2 token, refreshing if necessary
func (p *OAuth2Provider) getValidToken(ctx context.Context) (*oauth2.Token, error) {
	if err := p.loadToken(); err != nil {
		return nil, fmt.Errorf("failed to load token: %w", err)
	}

	if p.token.Valid() {
		return p.token, nil
	}

	// Try to refresh the token
	refreshed, err := p.refreshToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	p.token = refreshed
	if err := p.saveToken(); err != nil {
		return nil, fmt.Errorf("failed to save refreshed token: %w", err)
	}

	return p.token, nil
}

// refreshToken attempts to refresh the OAuth2 token
func (p *OAuth2Provider) refreshToken(ctx context.Context) (*oauth2.Token, error) {
	if p.token == nil || p.token.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	tokenSource := p.config.TokenSource(ctx, p.token)
	return tokenSource.Token()
}

// performOAuth2Flow performs the complete OAuth2 flow
func (p *OAuth2Provider) performOAuth2Flow(ctx context.Context) error {
	// Generate authorization URL
	state := fmt.Sprintf("state-%d", time.Now().Unix())
	authURL := p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	// Start local server for callback
	code := make(chan string)
	errCh := make(chan error)

	server := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 15 * time.Second,
	}
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Check state parameter
		if r.FormValue("state") != state {
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			errCh <- fmt.Errorf("invalid state parameter")
			return
		}

		// Get authorization code
		authCode := r.FormValue("code")
		if authCode == "" {
			http.Error(w, "No authorization code received", http.StatusBadRequest)
			errCh <- fmt.Errorf("no authorization code received")
			return
		}

		// Send success response
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<html><body><h1>Authorization successful!</h1><p>You can close this window and return to the terminal.</p></body></html>`))

		code <- authCode
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Open browser (user will need to do this manually for now)
	fmt.Printf("Please open the following URL in your browser to authorize the application:\n\n%s\n\n", authURL)
	fmt.Println("Waiting for authorization...")

	// Wait for authorization code or error
	select {
	case authCode := <-code:
		// Exchange code for token
		token, err := p.config.Exchange(ctx, authCode)
		if err != nil {
			if shutdownErr := server.Shutdown(ctx); shutdownErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: server shutdown error: %v\n", shutdownErr)
			}
			return fmt.Errorf("failed to exchange authorization code: %w", err)
		}

		p.token = token
		_ = server.Shutdown(ctx) // Ignore shutdown errors in cleanup

		// Save token
		if err := p.saveToken(); err != nil {
			return fmt.Errorf("failed to save token: %w", err)
		}

		fmt.Println("OAuth2 authentication successful!")
		return nil

	case err := <-errCh:
		_ = server.Shutdown(ctx) // Ignore shutdown errors in cleanup
		return fmt.Errorf("OAuth2 flow failed: %w", err)

	case <-ctx.Done():
		_ = server.Shutdown(ctx) // Ignore shutdown errors in cleanup
		return fmt.Errorf("OAuth2 flow canceled: %w", ctx.Err())
	}
}

// ValidateOAuth2 validates the OAuth2 configuration by making a test API call
func (p *OAuth2Provider) ValidateOAuth2(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("OAuth2 is not configured")
	}

	// Create a temporary client to test the credentials
	client, err := p.GetClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create client for validation: %w", err)
	}

	// Make a simple API call to validate the credentials
	req := &texttospeechpb.ListVoicesRequest{}
	_, err = client.ListVoices(ctx, req)
	if err != nil {
		return fmt.Errorf("OAuth2 validation failed: %w", err)
	}

	return nil
}

// RevokeToken revokes the current OAuth2 token
func (p *OAuth2Provider) RevokeToken(ctx context.Context) error {
	if p.token == nil {
		return nil // No token to revoke
	}

	// Revoke token with Google
	revokeURL := fmt.Sprintf("https://oauth2.googleapis.com/revoke?token=%s", url.QueryEscape(p.token.AccessToken))
	// #nosec G107 - URL is constructed with Google's official revoke endpoint
	resp, err := http.Post(revokeURL, "application/x-www-form-urlencoded", nil)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	resp.Body.Close()

	// Remove token file
	if err := os.Remove(p.tokenFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove token file: %w", err)
	}

	// Clear in-memory token
	p.token = nil

	return nil
}

// Close closes the underlying client connection
func (p *OAuth2Provider) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

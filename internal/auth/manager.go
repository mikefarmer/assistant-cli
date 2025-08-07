package auth

import (
	"context"
	"fmt"
	"os"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
)

// AuthMethod represents the different authentication methods available
type AuthMethod int

const (
	// AuthMethodAPIKey uses API key authentication (simplest)
	AuthMethodAPIKey AuthMethod = iota
	// AuthMethodServiceAccount uses service account JSON file
	AuthMethodServiceAccount
	// AuthMethodOAuth2 uses OAuth2 flow with browser
	AuthMethodOAuth2
)

// String returns the string representation of the auth method
func (a AuthMethod) String() string {
	switch a {
	case AuthMethodAPIKey:
		return "apikey"
	case AuthMethodServiceAccount:
		return "serviceaccount"
	case AuthMethodOAuth2:
		return "oauth2"
	default:
		return "unknown"
	}
}

// AuthConfig holds the configuration for authentication
type AuthConfig struct {
	Method             AuthMethod
	APIKey             string
	ServiceAccountFile string
	OAuth2ClientID     string
	OAuth2ClientSecret string
	OAuth2TokenFile    string
}

// AuthProvider interface defines the contract for authentication providers
type AuthProvider interface {
	// GetClient returns a configured Google Cloud TTS client
	GetClient(ctx context.Context) (*texttospeech.Client, error)
	// IsConfigured returns true if the provider is properly configured
	IsConfigured() bool
	// GetMethod returns the authentication method
	GetMethod() AuthMethod
	// Authenticate performs any necessary authentication steps
	Authenticate(ctx context.Context) error
}

// AuthManager coordinates between different authentication methods
type AuthManager struct {
	config    AuthConfig
	providers map[AuthMethod]AuthProvider
	active    AuthProvider
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(config AuthConfig) *AuthManager {
	manager := &AuthManager{
		config:    config,
		providers: make(map[AuthMethod]AuthProvider),
	}

	// Initialize providers
	manager.providers[AuthMethodAPIKey] = NewAPIKeyProvider(config.APIKey)
	manager.providers[AuthMethodServiceAccount] = NewServiceAccountProvider(config.ServiceAccountFile)
	manager.providers[AuthMethodOAuth2] = NewOAuth2Provider(config.OAuth2ClientID,
		config.OAuth2ClientSecret, config.OAuth2TokenFile)

	return manager
}

// SelectAuthMethod determines the best authentication method to use
// Priority: explicit config > environment variables > auto-detection
func (am *AuthManager) SelectAuthMethod() (AuthMethod, error) {
	// If method is explicitly set, use it
	if am.config.Method != AuthMethodAPIKey || am.config.APIKey != "" {
		return am.config.Method, nil
	}

	// Check for API key in environment
	if apiKey := os.Getenv("ASSISTANT_CLI_API_KEY"); apiKey != "" {
		return AuthMethodAPIKey, nil
	}

	// Check for service account file
	if serviceAccountFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); serviceAccountFile != "" {
		if _, err := os.Stat(serviceAccountFile); err == nil {
			return AuthMethodServiceAccount, nil
		}
	}

	// Check for OAuth2 configuration
	if am.config.OAuth2ClientID != "" && am.config.OAuth2ClientSecret != "" {
		return AuthMethodOAuth2, nil
	}

	// Default to API key method (user will need to provide key)
	return AuthMethodAPIKey, nil
}

// GetClient returns an authenticated Google Cloud TTS client
func (am *AuthManager) GetClient(ctx context.Context) (*texttospeech.Client, error) {
	if am.active == nil {
		method, err := am.SelectAuthMethod()
		if err != nil {
			return nil, fmt.Errorf("failed to select auth method: %w", err)
		}

		provider, exists := am.providers[method]
		if !exists {
			return nil, fmt.Errorf("no provider for auth method: %s", method)
		}

		if !provider.IsConfigured() {
			return nil, fmt.Errorf("authentication provider %s is not configured", method)
		}

		// Authenticate if necessary
		if err := provider.Authenticate(ctx); err != nil {
			return nil, fmt.Errorf("authentication failed for method %s: %w", method, err)
		}

		am.active = provider
	}

	return am.active.GetClient(ctx)
}

// GetActiveMethod returns the currently active authentication method
func (am *AuthManager) GetActiveMethod() AuthMethod {
	if am.active != nil {
		return am.active.GetMethod()
	}
	return AuthMethodAPIKey // Default
}

// IsConfigured returns true if any authentication method is properly configured
func (am *AuthManager) IsConfigured() bool {
	for _, provider := range am.providers {
		if provider.IsConfigured() {
			return true
		}
	}
	return false
}

// Validate checks if the authentication is properly configured and working
func (am *AuthManager) Validate(ctx context.Context) error {
	if am.active == nil {
		method, err := am.SelectAuthMethod()
		if err != nil {
			return fmt.Errorf("failed to select auth method: %w", err)
		}

		provider, exists := am.providers[method]
		if !exists {
			return fmt.Errorf("no provider for auth method: %s", method)
		}

		if !provider.IsConfigured() {
			return fmt.Errorf("authentication provider %s is not configured", method)
		}

		// Authenticate if necessary
		if err := provider.Authenticate(ctx); err != nil {
			return fmt.Errorf("authentication failed for method %s: %w", method, err)
		}

		am.active = provider
	}

	// Test the connection by creating a client
	_, err := am.GetClient(ctx)
	return err
}

// DefaultAuthConfig returns a default authentication configuration
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		Method:             AuthMethodAPIKey,
		APIKey:             os.Getenv("ASSISTANT_CLI_API_KEY"),
		ServiceAccountFile: os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		OAuth2ClientID:     os.Getenv("ASSISTANT_CLI_OAUTH2_CLIENT_ID"),
		OAuth2ClientSecret: os.Getenv("ASSISTANT_CLI_OAUTH2_CLIENT_SECRET"),
		OAuth2TokenFile:    os.Getenv("ASSISTANT_CLI_OAUTH2_TOKEN_FILE"),
	}
}

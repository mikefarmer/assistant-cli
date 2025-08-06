package auth

import (
	"context"
	"fmt"
	"os"
	"strings"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"google.golang.org/api/option"
)

// APIKeyProvider implements authentication using Google Cloud API keys
type APIKeyProvider struct {
	apiKey string
	client *texttospeech.Client
}

// NewAPIKeyProvider creates a new API key authentication provider
func NewAPIKeyProvider(apiKey string) *APIKeyProvider {
	// If no API key provided, try to get from environment
	if apiKey == "" {
		apiKey = os.Getenv("ASSISTANT_CLI_API_KEY")
	}
	
	return &APIKeyProvider{
		apiKey: apiKey,
	}
}

// GetClient returns a Google Cloud TTS client configured with API key authentication
func (p *APIKeyProvider) GetClient(ctx context.Context) (*texttospeech.Client, error) {
	if p.client != nil {
		return p.client, nil
	}

	if !p.IsConfigured() {
		return nil, fmt.Errorf("API key is not configured")
	}

	// Create client with API key
	client, err := texttospeech.NewClient(ctx, option.WithAPIKey(p.apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create TTS client with API key: %w", err)
	}

	p.client = client
	return p.client, nil
}

// IsConfigured returns true if the API key is available and appears valid
func (p *APIKeyProvider) IsConfigured() bool {
	return p.apiKey != "" && p.isValidAPIKey(p.apiKey)
}

// GetMethod returns the authentication method
func (p *APIKeyProvider) GetMethod() AuthMethod {
	return AuthMethodAPIKey
}

// Authenticate performs authentication (no-op for API keys)
func (p *APIKeyProvider) Authenticate(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("API key is not configured. Set ASSISTANT_CLI_API_KEY environment variable or use --api-key flag")
	}
	return nil
}

// SetAPIKey updates the API key (useful for the login command)
func (p *APIKeyProvider) SetAPIKey(apiKey string) {
	p.apiKey = apiKey
	// Clear cached client to force recreation
	if p.client != nil {
		p.client.Close()
		p.client = nil
	}
}

// GetAPIKey returns the current API key (for testing purposes)
func (p *APIKeyProvider) GetAPIKey() string {
	return p.apiKey
}

// isValidAPIKey performs basic validation of the API key format
func (p *APIKeyProvider) isValidAPIKey(apiKey string) bool {
	// Basic validation - Google Cloud API keys typically start with "AIza" and are 39 characters long
	// However, this can vary, so we'll do minimal validation
	apiKey = strings.TrimSpace(apiKey)
	
	if len(apiKey) < 20 {
		return false
	}
	
	// Check for common prefixes (this is not exhaustive but covers most cases)
	validPrefixes := []string{"AIza", "BIza", "CIza", "DIza"}
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(apiKey, prefix) {
			return len(apiKey) >= 35 && len(apiKey) <= 50
		}
	}
	
	// If it doesn't match common patterns but is reasonable length, allow it
	// This handles edge cases and future API key formats
	return len(apiKey) >= 20 && len(apiKey) <= 100
}

// ValidateAPIKey validates the API key by making a test API call
func (p *APIKeyProvider) ValidateAPIKey(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("API key is not configured")
	}

	// Create a temporary client to test the API key
	client, err := texttospeech.NewClient(ctx, option.WithAPIKey(p.apiKey))
	if err != nil {
		return fmt.Errorf("failed to create client for validation: %w", err)
	}
	defer client.Close()

	// Make a simple API call to validate the key
	// List voices is a good test call as it's lightweight
	req := &texttospeechpb.ListVoicesRequest{}
	_, err = client.ListVoices(ctx, req)
	if err != nil {
		return fmt.Errorf("API key validation failed: %w", err)
	}

	return nil
}

// Close closes the underlying client connection
func (p *APIKeyProvider) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
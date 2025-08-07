package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"google.golang.org/api/option"
)

// ServiceAccountProvider implements authentication using Google Cloud service account JSON files
type ServiceAccountProvider struct {
	serviceAccountFile string
	client             *texttospeech.Client
}

// NewServiceAccountProvider creates a new service account authentication provider
func NewServiceAccountProvider(serviceAccountFile string) *ServiceAccountProvider {
	// If no service account file provided, try to get from environment
	if serviceAccountFile == "" {
		serviceAccountFile = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	}

	return &ServiceAccountProvider{
		serviceAccountFile: serviceAccountFile,
	}
}

// GetClient returns a Google Cloud TTS client configured with service account authentication
func (p *ServiceAccountProvider) GetClient(ctx context.Context) (*texttospeech.Client, error) {
	if p.client != nil {
		return p.client, nil
	}

	if !p.IsConfigured() {
		return nil, fmt.Errorf("service account file is not configured or invalid")
	}

	// Create client with service account credentials
	client, err := texttospeech.NewClient(ctx, option.WithCredentialsFile(p.serviceAccountFile))
	if err != nil {
		return nil, fmt.Errorf("failed to create TTS client with service account: %w", err)
	}

	p.client = client
	return p.client, nil
}

// IsConfigured returns true if the service account file is available and appears valid
func (p *ServiceAccountProvider) IsConfigured() bool {
	if p.serviceAccountFile == "" {
		return false
	}

	// Check if file exists
	if _, err := os.Stat(p.serviceAccountFile); os.IsNotExist(err) {
		return false
	}

	// Validate JSON structure
	return p.isValidServiceAccountFile(p.serviceAccountFile)
}

// GetMethod returns the authentication method
func (p *ServiceAccountProvider) GetMethod() AuthMethod {
	return AuthMethodServiceAccount
}

// Authenticate performs authentication (validation for service account)
func (p *ServiceAccountProvider) Authenticate(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("service account file is not configured. " +
			"Set GOOGLE_APPLICATION_CREDENTIALS environment variable or use --service-account flag")
	}
	return nil
}

// SetServiceAccountFile updates the service account file path
func (p *ServiceAccountProvider) SetServiceAccountFile(file string) {
	p.serviceAccountFile = file
	// Clear cached client to force recreation
	if p.client != nil {
		p.client.Close()
		p.client = nil
	}
}

// GetServiceAccountFile returns the current service account file path
func (p *ServiceAccountProvider) GetServiceAccountFile() string {
	return p.serviceAccountFile
}

// ServiceAccountKey represents the structure of a Google Cloud service account JSON key
type ServiceAccountKey struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

// isValidServiceAccountFile validates the service account JSON file structure
func (p *ServiceAccountProvider) isValidServiceAccountFile(filename string) bool {
	data, err := os.ReadFile(filename)
	if err != nil {
		return false
	}

	var key ServiceAccountKey
	if err := json.Unmarshal(data, &key); err != nil {
		return false
	}

	// Validate required fields
	required := map[string]string{
		"type":         key.Type,
		"project_id":   key.ProjectID,
		"private_key":  key.PrivateKey,
		"client_email": key.ClientEmail,
		"client_id":    key.ClientID,
	}

	for field, value := range required {
		if value == "" {
			fmt.Printf("Service account file missing required field: %s\n", field)
			return false
		}
	}

	// Check that type is service_account
	if key.Type != "service_account" {
		return false
	}

	return true
}

// ValidateServiceAccount validates the service account by making a test API call
func (p *ServiceAccountProvider) ValidateServiceAccount(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("service account file is not configured")
	}

	// Create a temporary client to test the service account
	client, err := texttospeech.NewClient(ctx, option.WithCredentialsFile(p.serviceAccountFile))
	if err != nil {
		return fmt.Errorf("failed to create client for validation: %w", err)
	}
	defer client.Close()

	// Make a simple API call to validate the credentials
	req := &texttospeechpb.ListVoicesRequest{}
	_, err = client.ListVoices(ctx, req)
	if err != nil {
		return fmt.Errorf("service account validation failed: %w", err)
	}

	return nil
}

// GetProjectID returns the project ID from the service account file
func (p *ServiceAccountProvider) GetProjectID() (string, error) {
	if !p.IsConfigured() {
		return "", fmt.Errorf("service account file is not configured")
	}

	data, err := os.ReadFile(p.serviceAccountFile)
	if err != nil {
		return "", fmt.Errorf("failed to read service account file: %w", err)
	}

	var key ServiceAccountKey
	if err := json.Unmarshal(data, &key); err != nil {
		return "", fmt.Errorf("failed to parse service account file: %w", err)
	}

	return key.ProjectID, nil
}

// GetClientEmail returns the client email from the service account file
func (p *ServiceAccountProvider) GetClientEmail() (string, error) {
	if !p.IsConfigured() {
		return "", fmt.Errorf("service account file is not configured")
	}

	data, err := os.ReadFile(p.serviceAccountFile)
	if err != nil {
		return "", fmt.Errorf("failed to read service account file: %w", err)
	}

	var key ServiceAccountKey
	if err := json.Unmarshal(data, &key); err != nil {
		return "", fmt.Errorf("failed to parse service account file: %w", err)
	}

	return key.ClientEmail, nil
}

// Close closes the underlying client connection
func (p *ServiceAccountProvider) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

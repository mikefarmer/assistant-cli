package auth

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuthManager(t *testing.T) {
	config := AuthConfig{
		Method:              AuthMethodAPIKey,
		APIKey:              "test-api-key",
		ServiceAccountFile:  "/path/to/service-account.json",
		OAuth2ClientID:      "test-client-id",
		OAuth2ClientSecret:  "test-client-secret",
		OAuth2TokenFile:     "/path/to/token.json",
	}

	manager := NewAuthManager(config)

	assert.NotNil(t, manager)
	assert.Equal(t, config, manager.config)
	assert.Len(t, manager.providers, 3)
	assert.Contains(t, manager.providers, AuthMethodAPIKey)
	assert.Contains(t, manager.providers, AuthMethodServiceAccount)
	assert.Contains(t, manager.providers, AuthMethodOAuth2)
}

func TestAuthMethod_String(t *testing.T) {
	testCases := []struct {
		method   AuthMethod
		expected string
	}{
		{AuthMethodAPIKey, "apikey"},
		{AuthMethodServiceAccount, "serviceaccount"},
		{AuthMethodOAuth2, "oauth2"},
		{AuthMethod(999), "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.method.String()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestDefaultAuthConfig(t *testing.T) {
	// Save original env vars
	origAPIKey := os.Getenv("ASSISTANT_CLI_API_KEY")
	origServiceAccount := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	origOAuth2ClientID := os.Getenv("ASSISTANT_CLI_OAUTH2_CLIENT_ID")
	origOAuth2ClientSecret := os.Getenv("ASSISTANT_CLI_OAUTH2_CLIENT_SECRET")
	origOAuth2TokenFile := os.Getenv("ASSISTANT_CLI_OAUTH2_TOKEN_FILE")

	// Set test values
	os.Setenv("ASSISTANT_CLI_API_KEY", "test-api-key")
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/path/to/service.json")
	os.Setenv("ASSISTANT_CLI_OAUTH2_CLIENT_ID", "test-client-id")
	os.Setenv("ASSISTANT_CLI_OAUTH2_CLIENT_SECRET", "test-client-secret")
	os.Setenv("ASSISTANT_CLI_OAUTH2_TOKEN_FILE", "/path/to/token.json")

	// Test default config
	config := DefaultAuthConfig()

	assert.Equal(t, AuthMethodAPIKey, config.Method)
	assert.Equal(t, "test-api-key", config.APIKey)
	assert.Equal(t, "/path/to/service.json", config.ServiceAccountFile)
	assert.Equal(t, "test-client-id", config.OAuth2ClientID)
	assert.Equal(t, "test-client-secret", config.OAuth2ClientSecret)
	assert.Equal(t, "/path/to/token.json", config.OAuth2TokenFile)

	// Restore original env vars
	if origAPIKey != "" {
		os.Setenv("ASSISTANT_CLI_API_KEY", origAPIKey)
	} else {
		os.Unsetenv("ASSISTANT_CLI_API_KEY")
	}
	if origServiceAccount != "" {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", origServiceAccount)
	} else {
		os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
	}
	if origOAuth2ClientID != "" {
		os.Setenv("ASSISTANT_CLI_OAUTH2_CLIENT_ID", origOAuth2ClientID)
	} else {
		os.Unsetenv("ASSISTANT_CLI_OAUTH2_CLIENT_ID")
	}
	if origOAuth2ClientSecret != "" {
		os.Setenv("ASSISTANT_CLI_OAUTH2_CLIENT_SECRET", origOAuth2ClientSecret)
	} else {
		os.Unsetenv("ASSISTANT_CLI_OAUTH2_CLIENT_SECRET")
	}
	if origOAuth2TokenFile != "" {
		os.Setenv("ASSISTANT_CLI_OAUTH2_TOKEN_FILE", origOAuth2TokenFile)
	} else {
		os.Unsetenv("ASSISTANT_CLI_OAUTH2_TOKEN_FILE")
	}
}

func TestAuthManager_SelectAuthMethod(t *testing.T) {
	testCases := []struct {
		name     string
		config   AuthConfig
		envVars  map[string]string
		expected AuthMethod
	}{
		{
			name: "explicit API key config",
			config: AuthConfig{
				Method: AuthMethodAPIKey,
				APIKey: "explicit-key",
			},
			expected: AuthMethodAPIKey,
		},
		{
			name: "explicit service account config",
			config: AuthConfig{
				Method: AuthMethodServiceAccount,
			},
			expected: AuthMethodServiceAccount,
		},
		{
			name: "explicit OAuth2 config",
			config: AuthConfig{
				Method: AuthMethodOAuth2,
			},
			expected: AuthMethodOAuth2,
		},
		{
			name: "API key from environment",
			config: AuthConfig{
				Method: AuthMethodAPIKey,
			},
			envVars: map[string]string{
				"ASSISTANT_CLI_API_KEY": "env-api-key",
			},
			expected: AuthMethodAPIKey,
		},
		{
			name:   "OAuth2 with client credentials",
			config: AuthConfig{
				Method:             AuthMethodAPIKey,
				OAuth2ClientID:     "test-client-id",
				OAuth2ClientSecret: "test-client-secret",
			},
			expected: AuthMethodOAuth2,
		},
		{
			name:     "default to API key",
			config:   AuthConfig{},
			expected: AuthMethodAPIKey,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tc.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			manager := NewAuthManager(tc.config)
			method, err := manager.SelectAuthMethod()

			require.NoError(t, err)
			assert.Equal(t, tc.expected, method)
		})
	}
}

func TestAuthManager_IsConfigured(t *testing.T) {
	testCases := []struct {
		name     string
		config   AuthConfig
		expected bool
	}{
		{
			name: "API key configured",
			config: AuthConfig{
				APIKey: "AIza1234567890123456789012345678901234", // Use a valid API key
			},
			expected: true,
		},
		{
			name:     "no configuration",
			config:   AuthConfig{},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager := NewAuthManager(tc.config)
			result := manager.IsConfigured()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestAuthManager_GetActiveMethod(t *testing.T) {
	config := AuthConfig{
		Method: AuthMethodServiceAccount,
		APIKey: "test-api-key",
	}

	manager := NewAuthManager(config)

	// Before activation, should return default
	method := manager.GetActiveMethod()
	assert.Equal(t, AuthMethodAPIKey, method)

	// After setting active provider
	manager.active = manager.providers[AuthMethodServiceAccount]
	method = manager.GetActiveMethod()
	assert.Equal(t, AuthMethodServiceAccount, method)
}

func TestAuthManager_GetClient_NoCredentials(t *testing.T) {
	// Test with empty config (no credentials)
	config := AuthConfig{}
	manager := NewAuthManager(config)

	ctx := context.Background()
	_, err := manager.GetClient(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}
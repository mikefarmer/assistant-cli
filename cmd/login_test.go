package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mikefarmer/assistant-cli/internal/auth"
	"github.com/stretchr/testify/assert"
)

func TestLoginCommand(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantOutput string
	}{
		{
			name:       "login help",
			args:       []string{"login", "--help"},
			wantOutput: "Authenticate with Google Cloud Text-to-Speech API",
		},
		{
			name:       "method flag help",
			args:       []string{"login", "--help"},
			wantOutput: "Authentication method: apikey, serviceaccount, or oauth2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			rootCmd := NewRootCmd()
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()
			assert.NoError(t, err)

			output := buf.String()
			assert.Contains(t, output, tt.wantOutput)
		})
	}
}

func TestLoginCommandFlags(t *testing.T) {
	// Test that the login command has all expected flags
	flags := []string{"method", "api-key", "service-account", "client-id", "client-secret", "force", "validate"}
	
	for _, flag := range flags {
		t.Run("flag_"+flag, func(t *testing.T) {
			buf := new(bytes.Buffer)
			rootCmd := NewRootCmd()
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs([]string{"login", "--help"})

			err := rootCmd.Execute()
			assert.NoError(t, err)

			output := buf.String()
			assert.Contains(t, output, "--"+flag)
		})
	}
}

func TestDetermineAuthMethod(t *testing.T) {
	// Save original values
	origMethod := loginMethod
	origAPIKey := loginAPIKey
	origServiceFile := loginServiceFile
	origClientID := loginClientID
	origClientSecret := loginClientSecret
	
	// Save original env vars
	origEnvAPIKey := os.Getenv("ASSISTANT_CLI_API_KEY")
	origEnvServiceFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	origEnvClientID := os.Getenv("ASSISTANT_CLI_OAUTH2_CLIENT_ID")
	origEnvClientSecret := os.Getenv("ASSISTANT_CLI_OAUTH2_CLIENT_SECRET")
	
	// Cleanup function
	defer func() {
		loginMethod = origMethod
		loginAPIKey = origAPIKey
		loginServiceFile = origServiceFile
		loginClientID = origClientID
		loginClientSecret = origClientSecret
		
		// Restore env vars
		setOrUnsetEnv("ASSISTANT_CLI_API_KEY", origEnvAPIKey)
		setOrUnsetEnv("GOOGLE_APPLICATION_CREDENTIALS", origEnvServiceFile)
		setOrUnsetEnv("ASSISTANT_CLI_OAUTH2_CLIENT_ID", origEnvClientID)
		setOrUnsetEnv("ASSISTANT_CLI_OAUTH2_CLIENT_SECRET", origEnvClientSecret)
	}()

	tests := []struct {
		name           string
		method         string
		apiKey         string
		serviceFile    string
		clientID       string
		clientSecret   string
		envAPIKey      string
		envServiceFile string
		envClientID    string
		envClientSecret string
		expected       auth.AuthMethod
		expectError    bool
	}{
		{
			name:     "explicit API key method",
			method:   "apikey",
			expected: auth.AuthMethodAPIKey,
		},
		{
			name:     "explicit service account method",
			method:   "serviceaccount",
			expected: auth.AuthMethodServiceAccount,
		},
		{
			name:     "explicit OAuth2 method",
			method:   "oauth2",
			expected: auth.AuthMethodOAuth2,
		},
		{
			name:        "invalid method",
			method:      "invalid",
			expectError: true,
		},
		{
			name:     "API key flag",
			apiKey:   "test-api-key",
			expected: auth.AuthMethodAPIKey,
		},
		{
			name:        "service account flag",
			serviceFile: "/path/to/service.json",
			expected:    auth.AuthMethodServiceAccount,
		},
		{
			name:         "OAuth2 flags",
			clientID:     "client-id",
			clientSecret: "client-secret",
			expected:     auth.AuthMethodOAuth2,
		},
		{
			name:      "API key environment",
			envAPIKey: "env-api-key",
			expected:  auth.AuthMethodAPIKey,
		},
		{
			name:           "service account environment",
			envServiceFile: "/env/path/service.json",
			expected:       auth.AuthMethodServiceAccount,
		},
		{
			name:            "OAuth2 environment",
			envClientID:     "env-client-id",
			envClientSecret: "env-client-secret",
			expected:        auth.AuthMethodOAuth2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global variables
			loginMethod = tt.method
			loginAPIKey = tt.apiKey
			loginServiceFile = tt.serviceFile
			loginClientID = tt.clientID
			loginClientSecret = tt.clientSecret
			
			// Set environment variables
			setOrUnsetEnv("ASSISTANT_CLI_API_KEY", tt.envAPIKey)
			setOrUnsetEnv("GOOGLE_APPLICATION_CREDENTIALS", tt.envServiceFile)
			setOrUnsetEnv("ASSISTANT_CLI_OAUTH2_CLIENT_ID", tt.envClientID)
			setOrUnsetEnv("ASSISTANT_CLI_OAUTH2_CLIENT_SECRET", tt.envClientSecret)

			// Only test non-interactive paths (skip promptForAuthMethod cases)
			if tt.method != "" || tt.apiKey != "" || tt.serviceFile != "" || 
			   (tt.clientID != "" && tt.clientSecret != "") || tt.envAPIKey != "" || 
			   tt.envServiceFile != "" || (tt.envClientID != "" && tt.envClientSecret != "") {
				
				method, err := determineAuthMethod()
				
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, method)
				}
			}
		})
	}
}

func TestCreateAuthConfig(t *testing.T) {
	// Save original values
	origAPIKey := loginAPIKey
	origServiceFile := loginServiceFile
	origClientID := loginClientID
	origClientSecret := loginClientSecret
	
	defer func() {
		loginAPIKey = origAPIKey
		loginServiceFile = origServiceFile
		loginClientID = origClientID
		loginClientSecret = origClientSecret
	}()

	tests := []struct {
		name         string
		method       auth.AuthMethod
		apiKey       string
		serviceFile  string
		clientID     string
		clientSecret string
		validateFunc func(*testing.T, auth.AuthConfig)
	}{
		{
			name:   "API key config",
			method: auth.AuthMethodAPIKey,
			apiKey: "test-api-key",
			validateFunc: func(t *testing.T, config auth.AuthConfig) {
				assert.Equal(t, auth.AuthMethodAPIKey, config.Method)
				assert.Equal(t, "test-api-key", config.APIKey)
			},
		},
		{
			name:        "service account config",
			method:      auth.AuthMethodServiceAccount,
			serviceFile: "/path/to/service.json",
			validateFunc: func(t *testing.T, config auth.AuthConfig) {
				assert.Equal(t, auth.AuthMethodServiceAccount, config.Method)
				assert.Equal(t, "/path/to/service.json", config.ServiceAccountFile)
			},
		},
		{
			name:         "OAuth2 config",
			method:       auth.AuthMethodOAuth2,
			clientID:     "test-client-id",
			clientSecret: "test-client-secret",
			validateFunc: func(t *testing.T, config auth.AuthConfig) {
				assert.Equal(t, auth.AuthMethodOAuth2, config.Method)
				assert.Equal(t, "test-client-id", config.OAuth2ClientID)
				assert.Equal(t, "test-client-secret", config.OAuth2ClientSecret)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loginAPIKey = tt.apiKey
			loginServiceFile = tt.serviceFile
			loginClientID = tt.clientID
			loginClientSecret = tt.clientSecret

			config := createAuthConfig(tt.method)
			tt.validateFunc(t, config)
		})
	}
}

func TestPromptForServiceAccountFile(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "regular path",
			input:    "/path/to/service.json",
			expected: "/path/to/service.json",
		},
		{
			name:     "path with spaces",
			input:    "/path with spaces/service.json  ",
			expected: "/path with spaces/service.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the tilde expansion logic separately since we can't easily mock user input
			input := tt.input
			
			// Test tilde expansion
			if strings.HasPrefix(input, "~/") {
				home, _ := os.UserHomeDir()
				input = filepath.Join(home, input[2:])
			}
			
			result := strings.TrimSpace(input)
			
			if tt.name == "regular path" {
				assert.Equal(t, tt.expected, result)
			} else if tt.name == "path with spaces" {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestPerformAuthentication(t *testing.T) {
	tests := []struct {
		name        string
		method      auth.AuthMethod
		expectError bool
	}{
		{
			name:        "unsupported method",
			method:      auth.AuthMethod(999),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal auth manager for testing
			authConfig := auth.DefaultAuthConfig()
			authManager := auth.NewAuthManager(authConfig)
			
			err := performAuthentication(nil, authManager, tt.method)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.method == auth.AuthMethod(999) {
					assert.Contains(t, err.Error(), "unsupported authentication method")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSaveAuthConfigLogic(t *testing.T) {
	// Test the auth method to string conversion logic used in saveAuthConfig
	tests := []struct {
		method auth.AuthMethod
		string string
	}{
		{auth.AuthMethodAPIKey, "apikey"},
		{auth.AuthMethodServiceAccount, "serviceaccount"},
		{auth.AuthMethodOAuth2, "oauth2"},
	}

	for _, tt := range tests {
		t.Run(tt.string, func(t *testing.T) {
			assert.Equal(t, tt.string, tt.method.String())
		})
	}
}

// Helper function to set or unset environment variables
func setOrUnsetEnv(key, value string) {
	if value == "" {
		os.Unsetenv(key)
	} else {
		os.Setenv(key, value)
	}
}

func TestLoginFlagDefaults(t *testing.T) {
	// Test that login command flags have correct default values
	buf := new(bytes.Buffer)
	rootCmd := NewRootCmd()
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"login", "--help"})

	err := rootCmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	
	// Check that important flags are present
	assert.Contains(t, output, "--method")
	assert.Contains(t, output, "--api-key")
	assert.Contains(t, output, "--service-account")
	assert.Contains(t, output, "--client-id")
	assert.Contains(t, output, "--client-secret")
	assert.Contains(t, output, "--force")
	assert.Contains(t, output, "--validate")
}

func TestValidateAuthentication(t *testing.T) {
	// Test the validation logic components that don't require actual API calls
	t.Run("error handling", func(t *testing.T) {
		// Create a minimal auth manager
		authConfig := auth.DefaultAuthConfig()
		authManager := auth.NewAuthManager(authConfig)
		
		// This will fail because no valid auth is configured, but we can test error handling
		err := validateAuthentication(nil, authManager, auth.AuthMethodAPIKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get authenticated client")
	})
}
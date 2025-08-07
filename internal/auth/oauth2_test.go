package auth

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestNewOAuth2Provider(t *testing.T) {
	provider := NewOAuth2Provider("client-id", "client-secret", "/path/to/token.json")
	assert.NotNil(t, provider)
	assert.Equal(t, AuthMethodOAuth2, provider.GetMethod())
}

func TestOAuth2Provider_IsConfigured(t *testing.T) {
	testCases := []struct {
		name         string
		clientID     string
		clientSecret string
		tokenFile    string
		expected     bool
	}{
		{"complete configuration", "client-id", "client-secret", "/path/to/token.json", false}, // false because no valid token exists
		{"missing client ID", "", "client-secret", "/path/to/token.json", false},
		{"missing client secret", "client-id", "", "/path/to/token.json", false},
		{"missing token file", "client-id", "client-secret", "", false},
		{"all empty", "", "", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := NewOAuth2Provider(tc.clientID, tc.clientSecret, tc.tokenFile)
			result := provider.IsConfigured()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestOAuth2Provider_loadToken(t *testing.T) {
	// Create temporary directory
	tempDir, err := ioutil.TempDir("", "oauth2_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create valid token
	validToken := &oauth2.Token{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}

	// Create valid token file
	validFile := filepath.Join(tempDir, "valid_token.json")
	validData, err := json.Marshal(validToken)
	require.NoError(t, err)
	err = ioutil.WriteFile(validFile, validData, 0600)
	require.NoError(t, err)

	// Create invalid JSON file
	invalidFile := filepath.Join(tempDir, "invalid_token.json")
	err = ioutil.WriteFile(invalidFile, []byte("invalid json"), 0600)
	require.NoError(t, err)

	testCases := []struct {
		name      string
		tokenFile string
		expectErr bool
	}{
		{"valid token file", validFile, false},
		{"non-existent file", "/path/to/nonexistent.json", true},
		{"invalid JSON file", invalidFile, true},
		{"empty path", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := NewOAuth2Provider("client-id", "client-secret", tc.tokenFile)
			err := provider.loadToken()
			
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider.token)
				assert.Equal(t, "access-token", provider.token.AccessToken)
				assert.Equal(t, "refresh-token", provider.token.RefreshToken)
			}
		})
	}
}

func TestOAuth2Provider_saveToken(t *testing.T) {
	// Create temporary directory
	tempDir, err := ioutil.TempDir("", "oauth2_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tokenFile := filepath.Join(tempDir, "test_token.json")
	provider := NewOAuth2Provider("client-id", "client-secret", tokenFile)

	// Set token in provider
	provider.token = &oauth2.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}

	// Test saving token
	err = provider.saveToken()
	assert.NoError(t, err)

	// Verify file exists and has correct permissions
	info, err := os.Stat(tokenFile)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode())

	// Verify token can be loaded back
	newProvider := NewOAuth2Provider("client-id", "client-secret", tokenFile)
	err = newProvider.loadToken()
	assert.NoError(t, err)
	assert.Equal(t, provider.token.AccessToken, newProvider.token.AccessToken)
	assert.Equal(t, provider.token.RefreshToken, newProvider.token.RefreshToken)
}

func TestOAuth2Provider_hasValidToken(t *testing.T) {
	testCases := []struct {
		name     string
		token    *oauth2.Token
		expected bool
	}{
		{
			name:     "nil token",
			token:    nil,
			expected: false,
		},
		{
			name: "expired token without refresh",
			token: &oauth2.Token{
				AccessToken: "access-token",
				Expiry:      time.Now().Add(-time.Hour),
			},
			expected: false,
		},
		{
			name: "expired token with refresh",
			token: &oauth2.Token{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
				Expiry:       time.Now().Add(-time.Hour),
			},
			expected: false, // token.Valid() returns false for expired tokens
		},
		{
			name: "valid token",
			token: &oauth2.Token{
				AccessToken: "access-token",
				Expiry:      time.Now().Add(time.Hour),
			},
			expected: true,
		},
		{
			name: "token without expiry",
			token: &oauth2.Token{
				AccessToken: "access-token",
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := NewOAuth2Provider("client-id", "client-secret", "")
			provider.token = tc.token
			result := provider.hasValidToken()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestOAuth2Provider_Authenticate_NoCredentials(t *testing.T) {
	provider := NewOAuth2Provider("", "", "")
	ctx := context.Background()

	err := provider.Authenticate(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OAuth2 client ID and secret are not configured")
}

func TestOAuth2Provider_GetClient_NoCredentials(t *testing.T) {
	provider := NewOAuth2Provider("", "", "")
	ctx := context.Background()

	_, err := provider.GetClient(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OAuth2 is not configured")
}
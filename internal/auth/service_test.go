package auth

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServiceAccountProvider(t *testing.T) {
	testCases := []struct {
		name         string
		keyFile      string
		expectConfig bool
	}{
		{"with key file", "/path/to/key.json", true},
		{"empty key file", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := NewServiceAccountProvider(tc.keyFile)
			assert.NotNil(t, provider)
			assert.Equal(t, AuthMethodServiceAccount, provider.GetMethod())
			// Note: IsConfigured will check file existence, so we can't test it easily here
		})
	}
}

func TestServiceAccountProvider_IsConfigured(t *testing.T) {
	// Create a temporary valid service account file
	validServiceAccount := map[string]interface{}{
		"type":                        "service_account",
		"project_id":                  "test-project",
		"private_key_id":              "key123",
		"private_key":                 "-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQC7VJTUt9Us8cKB\n-----END PRIVATE KEY-----\n",
		"client_email":                "test@test-project.iam.gserviceaccount.com",
		"client_id":                   "123456789",
		"auth_uri":                    "https://accounts.google.com/o/oauth2/auth",
		"token_uri":                   "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_x509_cert_url":        "https://www.googleapis.com/robot/v1/metadata/x509/test%40test-project.iam.gserviceaccount.com",
	}

	// Create temporary directory
	tempDir, err := ioutil.TempDir("", "auth_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create valid service account file
	validFile := filepath.Join(tempDir, "valid.json")
	validData, err := json.Marshal(validServiceAccount)
	require.NoError(t, err)
	err = ioutil.WriteFile(validFile, validData, 0600)
	require.NoError(t, err)

	// Create invalid JSON file
	invalidFile := filepath.Join(tempDir, "invalid.json")
	err = ioutil.WriteFile(invalidFile, []byte("invalid json"), 0600)
	require.NoError(t, err)

	// Create incomplete service account file (missing required fields)
	incompleteFile := filepath.Join(tempDir, "incomplete.json")
	incompleteData, err := json.Marshal(map[string]interface{}{
		"type": "service_account",
		// Missing required fields
	})
	require.NoError(t, err)
	err = ioutil.WriteFile(incompleteFile, incompleteData, 0600)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		keyFile  string
		expected bool
	}{
		{"valid service account file", validFile, true},
		{"non-existent file", "/path/to/nonexistent.json", false},
		{"invalid JSON file", invalidFile, false},
		{"incomplete service account", incompleteFile, false},
		{"empty path", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := NewServiceAccountProvider(tc.keyFile)
			result := provider.IsConfigured()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestServiceAccountProvider_Authenticate(t *testing.T) {
	provider := NewServiceAccountProvider("/path/to/key.json")
	ctx := context.Background()

	// Service account authentication should fail for non-configured provider
	err := provider.Authenticate(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service account file is not configured")
}

func TestServiceAccountProvider_GetClient_InvalidFile(t *testing.T) {
	provider := NewServiceAccountProvider("/path/to/nonexistent.json")
	ctx := context.Background()

	_, err := provider.GetClient(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service account file is not configured or invalid")
}

func TestServiceAccountProvider_GetClient_EmptyFile(t *testing.T) {
	provider := NewServiceAccountProvider("")
	ctx := context.Background()

	_, err := provider.GetClient(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service account file is not configured or invalid")
}
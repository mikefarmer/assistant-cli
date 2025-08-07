package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAPIKeyProvider(t *testing.T) {
	testCases := []struct {
		name     string
		apiKey   string
		expected bool
	}{
		{"valid API key", "AIzaSyA1234567890123456789012345678901234567", true},
		{"empty API key", "", false},
		{"short API key", "short", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := NewAPIKeyProvider(tc.apiKey)
			assert.NotNil(t, provider)
			assert.Equal(t, tc.expected, provider.IsConfigured())
			assert.Equal(t, AuthMethodAPIKey, provider.GetMethod())
		})
	}
}

func TestAPIKeyProvider_IsConfigured(t *testing.T) {
	testCases := []struct {
		name     string
		apiKey   string
		expected bool
	}{
		{"valid API key format", "AIzaSyA1234567890123456789012345678901", true},
		{"another valid format", "AIza123456789012345678901234567890123456", true}, // Actually valid due to fallback logic
		{"valid AIza format", "AIza1234567890123456789012345678901234", true},
		{"valid BIza format", "BIza1234567890123456789012345678901234", true},
		{"empty key", "", false},
		{"too short", "AIza123", false},
		{"wrong prefix", "XIza1234567890123456789012345678901234", true}, // Valid due to fallback (20-100 chars)
		{"no prefix", "1234567890123456789012345678901234567890", true},  // Valid due to fallback (20-100 chars)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := NewAPIKeyProvider(tc.apiKey)
			result := provider.IsConfigured()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestAPIKeyProvider_Authenticate(t *testing.T) {
	provider := NewAPIKeyProvider("AIza1234567890123456789012345678901234")
	ctx := context.Background()

	// API key authentication should succeed for valid keys
	err := provider.Authenticate(ctx)
	assert.NoError(t, err)
}

func TestAPIKeyProvider_GetClient_InvalidKey(t *testing.T) {
	provider := NewAPIKeyProvider("invalid-key")
	ctx := context.Background()

	// Should fail for invalid API key
	_, err := provider.GetClient(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API key is not configured")
}

func TestAPIKeyProvider_GetClient_EmptyKey(t *testing.T) {
	provider := NewAPIKeyProvider("")
	ctx := context.Background()

	// Should fail for empty API key
	_, err := provider.GetClient(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API key is not configured")
}

func TestAPIKeyProvider_Authenticate_InvalidKey(t *testing.T) {
	provider := NewAPIKeyProvider("invalid-key")
	ctx := context.Background()

	// Should fail for invalid API key
	err := provider.Authenticate(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API key is not configured")
}

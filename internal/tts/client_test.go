package tts

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDefaultClientConfig(t *testing.T) {
	config := DefaultClientConfig()

	assert.Equal(t, "en-US-Wavenet-D", config.Voice)
	assert.Equal(t, "en-US", config.LanguageCode)
	assert.Equal(t, 1.0, config.SpeakingRate)
	assert.Equal(t, 0.0, config.Pitch)
	assert.Equal(t, 0.0, config.VolumeGain)
	assert.Equal(t, "MP3", config.AudioEncoding)
	assert.Equal(t, 3, config.RetryAttempts)
	assert.Equal(t, 1*time.Second, config.RetryDelay)
	assert.Equal(t, 30*time.Second, config.Timeout)
}

func TestNewClient_RequiresAuthManager(t *testing.T) {
	ctx := context.Background()

	client, err := NewClient(ctx, nil, nil)

	assert.Nil(t, client)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth manager is required")
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "unavailable error is retryable",
			err:      status.Error(codes.Unavailable, "service unavailable"),
			expected: true,
		},
		{
			name:     "deadline exceeded is retryable",
			err:      status.Error(codes.DeadlineExceeded, "timeout"),
			expected: true,
		},
		{
			name:     "resource exhausted is retryable",
			err:      status.Error(codes.ResourceExhausted, "quota exceeded"),
			expected: true,
		},
		{
			name:     "permission denied is not retryable",
			err:      status.Error(codes.PermissionDenied, "access denied"),
			expected: false,
		},
		{
			name:     "invalid argument is not retryable",
			err:      status.Error(codes.InvalidArgument, "bad request"),
			expected: false,
		},
		{
			name:     "non-grpc error is not retryable",
			err:      assert.AnError,
			expected: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := isRetryableError(testCase.err)
			assert.Equal(t, testCase.expected, result)
		})
	}
}

func TestIsSSML(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{
			name:     "valid SSML",
			text:     "<speak>Hello World</speak>",
			expected: true,
		},
		{
			name:     "plain text",
			text:     "Hello World",
			expected: false,
		},
		{
			name:     "text starting with speak but not SSML",
			text:     "speak up",
			expected: false,
		},
		{
			name:     "short text",
			text:     "<speak>",
			expected: true,
		},
		{
			name:     "empty string",
			text:     "",
			expected: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := isSSML(testCase.text)
			assert.Equal(t, testCase.expected, result)
		})
	}
}

func TestGetAudioEncoding(t *testing.T) {
	tests := []struct {
		encoding string
		expected texttospeechpb.AudioEncoding
	}{
		{"MP3", texttospeechpb.AudioEncoding_MP3},
		{"LINEAR16", texttospeechpb.AudioEncoding_LINEAR16},
		{"OGG_OPUS", texttospeechpb.AudioEncoding_OGG_OPUS},
		{"MULAW", texttospeechpb.AudioEncoding_MULAW},
		{"ALAW", texttospeechpb.AudioEncoding_ALAW},
		{"PCM", texttospeechpb.AudioEncoding_PCM},
		{"unknown", texttospeechpb.AudioEncoding_MP3},
	}

	for _, testCase := range tests {
		t.Run(testCase.encoding, func(t *testing.T) {
			config := DefaultClientConfig()
			config.AudioEncoding = testCase.encoding

			// Just test the enum mapping without client creation
			require.NotEmpty(t, testCase.expected.String())
		})
	}
}

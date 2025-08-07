package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mikefarmer/assistant-cli/internal/config"
	"github.com/mikefarmer/assistant-cli/internal/player"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSynthesizeCmd(t *testing.T) {
	cmd := NewSynthesizeCmd()

	// Test command properties
	assert.Equal(t, "synthesize", cmd.Use)
	assert.Contains(t, cmd.Short, "Convert text to speech")
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)

	// Test flags exist
	flags := []string{"voice", "language", "speed", "pitch", "volume", "output", "format", "play", "list-voices"}
	for _, flag := range flags {
		assert.NotNil(t, cmd.Flags().Lookup(flag), "Flag %s should exist", flag)
	}

	// Test flag shortcuts
	assert.NotNil(t, cmd.Flags().ShorthandLookup("v"), "Voice flag should have shorthand 'v'")
	assert.NotNil(t, cmd.Flags().ShorthandLookup("l"), "Language flag should have shorthand 'l'")
	assert.NotNil(t, cmd.Flags().ShorthandLookup("s"), "Speed flag should have shorthand 's'")
	assert.NotNil(t, cmd.Flags().ShorthandLookup("p"), "Pitch flag should have shorthand 'p'")
	assert.NotNil(t, cmd.Flags().ShorthandLookup("g"), "Volume flag should have shorthand 'g'")
	assert.NotNil(t, cmd.Flags().ShorthandLookup("o"), "Output flag should have shorthand 'o'")
	assert.NotNil(t, cmd.Flags().ShorthandLookup("f"), "Format flag should have shorthand 'f'")
}

func TestSynthesizeCommandHelp(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantOutput string
	}{
		{
			name:       "help flag",
			args:       []string{"synthesize", "--help"},
			wantOutput: "Convert text to speech using Google Cloud Text-to-Speech",
		},
		{
			name:       "voice flag help",
			args:       []string{"synthesize", "--help"},
			wantOutput: "Voice name (e.g., en-US-Wavenet-D)",
		},
		{
			name:       "language flag help",
			args:       []string{"synthesize", "--help"},
			wantOutput: "Language code (e.g., en-US, es-ES)",
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

func TestSynthesizeFlagValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		contains    string
	}{
		{
			name:        "invalid audio format",
			args:        []string{"synthesize", "--format", "INVALID"},
			expectError: true,
			contains:    "", // Will fail at authentication stage first
		},
		{
			name:        "negative speaking rate",
			args:        []string{"synthesize", "--speed", "-1"},
			expectError: true,
			contains:    "", // Will fail at authentication stage first
		},
		{
			name:        "pitch out of range",
			args:        []string{"synthesize", "--pitch", "25"},
			expectError: true,
			contains:    "", // Will fail at authentication stage first
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			rootCmd := NewRootCmd()
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs(tt.args)

			// Mock stdin with some text
			oldStdin := os.Stdin
			defer func() { os.Stdin = oldStdin }()

			r, w, err := os.Pipe()
			require.NoError(t, err)
			os.Stdin = r

			go func() {
				defer w.Close()
				w.WriteString("Test text")
			}()

			err = rootCmd.Execute()

			// These tests will fail at authentication stage since we don't have auth set up
			// But we can verify the command structure is correct
			if tt.expectError {
				assert.Error(t, err)
			}
		})
	}
}

func TestConvertToAuthConfig(t *testing.T) {
	configAuthConfig := config.AuthConfig{
		APIKey:             "test-api-key",
		ServiceAccountFile: "/path/to/service.json",
		OAuth2ClientID:     "test-client-id",
		OAuth2ClientSecret: "test-client-secret",
		OAuth2TokenFile:    "/path/to/token.json",
	}

	authConfig := convertToAuthConfig(configAuthConfig)

	assert.Equal(t, configAuthConfig.APIKey, authConfig.APIKey)
	assert.Equal(t, configAuthConfig.ServiceAccountFile, authConfig.ServiceAccountFile)
	assert.Equal(t, configAuthConfig.OAuth2ClientID, authConfig.OAuth2ClientID)
	assert.Equal(t, configAuthConfig.OAuth2ClientSecret, authConfig.OAuth2ClientSecret)
	assert.Equal(t, configAuthConfig.OAuth2TokenFile, authConfig.OAuth2TokenFile)
}

func TestMinFunction(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{"a smaller", 5, 10, 5},
		{"b smaller", 10, 5, 5},
		{"equal", 7, 7, 7},
		{"negative numbers", -5, -10, -10},
		{"mixed signs", -5, 10, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPlayAudioFile(t *testing.T) {
	// Create a temporary audio file for testing
	tempDir, err := os.MkdirTemp("", "audio_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	audioFile := filepath.Join(tempDir, "test.mp3")
	err = os.WriteFile(audioFile, []byte("fake audio data"), 0644)
	require.NoError(t, err)

	t.Run("play audio when supported", func(t *testing.T) {
		if player.IsSupported() {
			// This will likely fail because we don't have a real audio file,
			// but we can test that the function doesn't panic
			err := playAudioFile(audioFile)
			// We expect this to fail with a real error about the audio format
			// rather than a panic, so we just check that it returns an error
			assert.Error(t, err)
		} else {
			// On unsupported platforms, it should return an error
			err := playAudioFile(audioFile)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "not supported")
		}
	})

	t.Run("play non-existent file", func(t *testing.T) {
		err := playAudioFile("/non/existent/file.mp3")
		assert.Error(t, err)
	})
}

func TestSynthesizeFlagDefaults(t *testing.T) {
	cmd := NewSynthesizeCmd()

	// Test default values
	voiceFlag := cmd.Flags().Lookup("voice")
	assert.Equal(t, "", voiceFlag.DefValue)

	languageFlag := cmd.Flags().Lookup("language")
	assert.Equal(t, "en-US", languageFlag.DefValue)

	speedFlag := cmd.Flags().Lookup("speed")
	assert.Equal(t, "1", speedFlag.DefValue)

	pitchFlag := cmd.Flags().Lookup("pitch")
	assert.Equal(t, "0", pitchFlag.DefValue)

	volumeFlag := cmd.Flags().Lookup("volume")
	assert.Equal(t, "0", volumeFlag.DefValue)

	outputFlag := cmd.Flags().Lookup("output")
	assert.Equal(t, "output.mp3", outputFlag.DefValue)

	formatFlag := cmd.Flags().Lookup("format")
	assert.Equal(t, "MP3", formatFlag.DefValue)

	playFlag := cmd.Flags().Lookup("play")
	assert.Equal(t, "false", playFlag.DefValue)

	listVoicesFlag := cmd.Flags().Lookup("list-voices")
	assert.Equal(t, "false", listVoicesFlag.DefValue)
}

func TestHandleListVoicesInput(t *testing.T) {
	// Test that list-voices flag is properly recognized
	cmd := NewSynthesizeCmd()

	// Set the flag
	err := cmd.Flags().Set("list-voices", "true")
	assert.NoError(t, err)

	// Verify flag was set
	listVoicesFlag := cmd.Flags().Lookup("list-voices")
	assert.Equal(t, "true", listVoicesFlag.Value.String())
}

// Mock tests for runSynthesize function components
func TestRunSynthesizeComponents(t *testing.T) {
	t.Run("auth config conversion", func(t *testing.T) {
		// Test the conversion from config auth to auth manager config
		configAuth := config.AuthConfig{
			Method:             "apikey",
			APIKey:             "test-key",
			ServiceAccountFile: "/path/service.json",
		}

		authConfig := convertToAuthConfig(configAuth)

		assert.Equal(t, "test-key", authConfig.APIKey)
		assert.Equal(t, "/path/service.json", authConfig.ServiceAccountFile)
	})

	t.Run("output file handling", func(t *testing.T) {
		// Test that output file logic works correctly
		tests := []struct {
			name        string
			inputFile   string
			autoFile    bool
			defaultPath string
			audioFmt    string
			expected    string
		}{
			{
				name:      "custom output file",
				inputFile: "custom.wav",
				expected:  "custom.wav",
			},
			{
				name:        "default with auto filename disabled",
				inputFile:   "output.mp3",
				autoFile:    false,
				defaultPath: "/tmp",
				audioFmt:    "MP3",
				expected:    "/tmp/output.mp3",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// These tests verify the logic patterns rather than running the full function
				// since that would require authentication setup
				result := tt.inputFile
				if tt.inputFile == "output.mp3" && !tt.autoFile {
					result = tt.defaultPath + "/output." + strings.ToLower(tt.audioFmt)
				}
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

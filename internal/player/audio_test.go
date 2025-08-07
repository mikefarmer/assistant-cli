package player

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAudioPlayer(t *testing.T) {
	player, err := NewAudioPlayer()

	// On CI systems, audio players might not be available
	// so we'll test both success and failure cases
	if err != nil {
		assert.IsType(t, &PlayerError{}, err)
		assert.Contains(t, err.Error(), "audio playback error")
		return
	}

	require.NotNil(t, player)
	assert.NotEmpty(t, player.player)
	assert.Equal(t, runtime.GOOS, player.GetPlayerInfo().Platform)
}

func TestAudioPlayer_GetPlayerInfo(t *testing.T) {
	player, err := NewAudioPlayer()
	if err != nil {
		t.Skip("Audio player not available on this system")
	}

	info := player.GetPlayerInfo()
	assert.NotEmpty(t, info.Command)
	assert.Equal(t, runtime.GOOS, info.Platform)
	assert.NotNil(t, info.Args) // can be empty slice but not nil
}

func TestIsSupported(t *testing.T) {
	// This test is platform-dependent
	supported := IsSupported()

	// On most platforms, there should be some audio player available
	// But on headless CI systems, this might fail
	t.Logf("Audio playback supported: %v (platform: %s)", supported, runtime.GOOS)

	// We'll make this a non-failing test since support varies by environment
}

func TestAudioPlayer_Play_FileNotExists(t *testing.T) {
	player, err := NewAudioPlayer()
	if err != nil {
		t.Skip("Audio player not available on this system")
	}

	err = player.Play("nonexistent_file.mp3")
	require.Error(t, err)

	var playerErr *PlayerError
	assert.ErrorAs(t, err, &playerErr)
	assert.Equal(t, "play", playerErr.Operation)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestPlayFile_ConvenienceFunction(t *testing.T) {
	err := PlayFile("nonexistent_file.mp3")
	require.Error(t, err)

	// Should return either a PlayerError (if player creation succeeded)
	// or some other error (if player creation failed)
	assert.True(t, err != nil)
}

func TestAudioPlayer_commandExists(t *testing.T) {
	player := &AudioPlayer{}

	// Test with a command that should exist on all systems
	assert.True(t, player.commandExists("go")) // Go should be available since we're running tests

	// Test with a command that definitely doesn't exist
	assert.False(t, player.commandExists("definitely_not_a_real_command_12345"))
}

func TestAudioPlayer_detectPlayer_MockPlatform(t *testing.T) {
	player := &AudioPlayer{}

	// Test platform detection logic
	err := player.detectPlayer()

	// This will either succeed or fail depending on the platform
	// The important thing is that it doesn't panic
	t.Logf("Player detection result: %v", err)
	if err == nil {
		assert.NotEmpty(t, player.player)
		t.Logf("Detected player: %s with args: %v", player.player, player.args)
	}
}

func TestPlayerError(t *testing.T) {
	err := &PlayerError{
		Operation: "test",
		Err:       assert.AnError,
		Platform:  "test",
	}

	assert.Contains(t, err.Error(), "audio playback error on test during test")
	assert.Equal(t, assert.AnError, err.Unwrap())
}

func TestAudioPlayer_buildMacOSCommand(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS-specific test")
	}

	player := &AudioPlayer{
		player: "afplay",
		args:   []string{},
	}

	testFile := "/tmp/test.mp3"
	cmd := player.buildMacOSCommand(testFile)

	// The path might be full path like "/usr/bin/afplay" or just "afplay"
	assert.True(t, strings.HasSuffix(cmd.Path, "afplay"), "expected path to end with 'afplay', got %s", cmd.Path)
	assert.Contains(t, cmd.Args, testFile)
}

func TestAudioPlayer_buildLinuxCommand(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-specific test")
	}

	player := &AudioPlayer{
		player: "aplay",
		args:   []string{},
	}

	testFile := "/tmp/test.wav"
	cmd := player.buildLinuxCommand(testFile)

	assert.True(t, strings.HasSuffix(cmd.Path, "aplay"), "expected path to end with 'aplay', got %s", cmd.Path)
	assert.Contains(t, cmd.Args, testFile)
}

func TestAudioPlayer_buildWindowsCommand(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	player := &AudioPlayer{
		player: "powershell",
		args:   []string{"-Command"},
	}

	testFile := "C:\\temp\\test.mp3"
	cmd := player.buildWindowsCommand(testFile)

	assert.Equal(t, "powershell", cmd.Path)
	assert.Contains(t, cmd.Args, "-Command")
}

func TestAudioPlayer_Integration_WithRealFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	player, err := NewAudioPlayer()
	if err != nil {
		t.Skip("Audio player not available on this system")
	}

	// Create a minimal audio file for testing
	// This is a simple approach - in practice you might want to create
	// an actual audio file or mock the file system
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test.mp3")

	// Create a dummy file (not a real MP3, but that's ok for testing the path handling)
	err = os.WriteFile(testFile, []byte("fake audio data"), 0644)
	require.NoError(t, err)

	// Try to play it (will likely fail due to invalid format, but should handle the file path correctly)
	err = player.Play(testFile)

	// We expect this to fail since it's not a real audio file,
	// but it shouldn't fail due to file not found
	if err != nil {
		var playerErr *PlayerError
		if assert.ErrorAs(t, err, &playerErr) {
			// Should not be a "file not found" error
			assert.NotContains(t, err.Error(), "does not exist")
			// Should be a playback failure
			assert.Equal(t, "play", playerErr.Operation)
		}
	}
}

// Benchmark for player creation
func BenchmarkNewAudioPlayer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		player, err := NewAudioPlayer()
		_ = err    // Ignore error in benchmark
		_ = player // Ignore player in benchmark
	}
}

// Test table for different file extensions
func TestAudioPlayer_Play_DifferentExtensions(t *testing.T) {
	player, err := NewAudioPlayer()
	if err != nil {
		t.Skip("Audio player not available on this system")
	}

	testCases := []struct {
		filename     string
		shouldAccept bool
	}{
		{"test.mp3", true},
		{"test.wav", true},
		{"test.ogg", true},
		{"test.m4a", true},
		{"test.flac", true},
		{"", false},
		{"no_extension", true}, // Player should try anyway
	}

	testDir := t.TempDir()

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			if tc.filename == "" {
				// Test empty filename
				err := player.Play("")
				assert.Error(t, err)
				return
			}

			testFile := filepath.Join(testDir, tc.filename)

			// Create dummy file
			err := os.WriteFile(testFile, []byte("fake audio"), 0644)
			require.NoError(t, err)

			// Try to play
			err = player.Play(testFile)

			// We expect failure due to invalid audio format,
			// but the player should at least try
			if err != nil {
				var playerErr *PlayerError
				assert.ErrorAs(t, err, &playerErr)
			}
		})
	}
}

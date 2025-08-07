package test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests for the assistant-cli binary
// These tests build and run the actual CLI binary

const testTimeout = 30 * time.Second

func TestCLIBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Build the CLI binary
	tempDir := t.TempDir()
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(tempDir, "assistant-cli-test"), "main.go")
	buildCmd.Dir = ".." // Set working directory to project root
	
	err := buildCmd.Run()
	require.NoError(t, err, "Failed to build CLI binary")
	
	// Clean up
	defer func() {
		os.Remove(filepath.Join(tempDir, "assistant-cli-test"))
	}()
	
	// Test that binary was created
	binaryPath := filepath.Join(tempDir, "assistant-cli-test")
	assert.FileExists(t, binaryPath)
}

func TestCLIHelp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	binary := buildTestBinary(t)
	defer os.Remove(binary)

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "root help",
			args:     []string{"--help"},
			expected: "A personal assistant CLI tool",
		},
		{
			name:     "version flag",
			args:     []string{"--version"},
			expected: "assistant-cli version",
		},
		{
			name:     "synthesize help",
			args:     []string{"synthesize", "--help"},
			expected: "Convert text to speech",
		},
		{
			name:     "login help",
			args:     []string{"login", "--help"},
			expected: "Authenticate with Google Cloud",
		},
		{
			name:     "config help",
			args:     []string{"config", "--help"},
			expected: "Manage configuration settings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()

			cmd := exec.CommandContext(ctx, binary, tt.args...)
			output, err := cmd.CombinedOutput()

			// Help commands should exit with code 0
			if strings.Contains(tt.args[len(tt.args)-1], "help") || tt.args[0] == "--help" {
				assert.NoError(t, err, "Help command should not error")
			}

			assert.Contains(t, string(output), tt.expected)
		})
	}
}

func TestCLIConfigCommands(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	binary := buildTestBinary(t)
	defer os.Remove(binary)

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	t.Run("config generate", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, binary, "config", "generate", configPath)
		output, err := cmd.CombinedOutput()

		// Command should succeed
		assert.NoError(t, err, "Config generate should succeed")
		assert.Contains(t, string(output), "Generated configuration file")
		assert.FileExists(t, configPath)

		// Check file content
		content, err := os.ReadFile(configPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "Assistant-CLI Configuration")
	})

	t.Run("config validate", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, binary, "config", "validate", configPath)
		output, _ := cmd.CombinedOutput()

		// Validation may fail due to missing auth, but command should run
		outputStr := string(output)
		assert.True(t, 
			strings.Contains(outputStr, "Configuration validation passed") ||
			strings.Contains(outputStr, "Configuration validation failed"),
			"Should show validation result")
	})

	t.Run("config show", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, binary, "config", "show")
		output, err := cmd.CombinedOutput()

		// Config show should work without auth
		assert.NoError(t, err, "Config show should succeed")
		assert.Contains(t, string(output), "auth:")
	})
}

func TestCLISynthesizeNoAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	binary := buildTestBinary(t)
	defer os.Remove(binary)

	t.Run("synthesize without auth", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, binary, "synthesize", "--output", "test.mp3")
		
		// Provide some test input
		cmd.Stdin = strings.NewReader("Hello, world!")
		
		output, err := cmd.CombinedOutput()

		// Should fail due to missing authentication
		assert.Error(t, err, "Should fail without authentication")
		assert.Contains(t, string(output), "authentication failed")
	})

	t.Run("synthesize list voices without auth", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, binary, "synthesize", "--list-voices")
		output, err := cmd.CombinedOutput()

		// Should fail due to missing authentication
		assert.Error(t, err, "Should fail without authentication")
		assert.Contains(t, string(output), "authentication failed")
	})
}

func TestCLILoginNoCredentials(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	binary := buildTestBinary(t)
	defer os.Remove(binary)

	t.Run("login without method", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, binary, "login")
		
		// This will likely hang waiting for user input, so we'll test with a method
		// that doesn't require interactive input but will fail validation
		cmd = exec.CommandContext(ctx, binary, "login", "--method", "apikey", "--api-key", "invalid-key", "--validate=false")
		output, _ := cmd.CombinedOutput()

		// Should either succeed (if validation is off) or fail with clear message
		outputStr := string(output)
		assert.True(t,
			strings.Contains(outputStr, "Authentication completed") ||
			strings.Contains(outputStr, "authentication failed") ||
			strings.Contains(outputStr, "API key is not properly configured"),
			"Should show clear authentication status")
	})
}

func TestCLIVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	binary := buildTestBinary(t)
	defer os.Remove(binary)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary, "--version")
	output, err := cmd.CombinedOutput()

	assert.NoError(t, err, "Version command should succeed")
	assert.Contains(t, string(output), "assistant-cli version")
	
	// Should contain either "dev" or a version number
	outputStr := string(output)
	assert.True(t,
		strings.Contains(outputStr, "dev") || 
		strings.Contains(outputStr, "v") ||
		strings.Contains(outputStr, "."),
		"Should contain version information")
}

// buildTestBinary builds the CLI binary for testing and returns the path
func buildTestBinary(t *testing.T) string {
	t.Helper()
	
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "assistant-cli-test")
	
	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "main.go")
	buildCmd.Dir = ".." // Set working directory to project root
	
	var stderr bytes.Buffer
	buildCmd.Stderr = &stderr
	
	err := buildCmd.Run()
	require.NoError(t, err, "Failed to build CLI binary: %s", stderr.String())
	
	return binaryPath
}

func TestCLIConfigGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	binary := buildTestBinary(t)
	defer os.Remove(binary)

	t.Run("generate config with force", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "force-config.yaml")

		// Create an existing file
		err := os.WriteFile(configPath, []byte("existing content"), 0644)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, binary, "config", "generate", "--force", configPath)
		output, err := cmd.CombinedOutput()

		assert.NoError(t, err, "Config generate with force should succeed")
		assert.Contains(t, string(output), "Generated configuration file")

		// Check that file was overwritten
		content, err := os.ReadFile(configPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "Assistant-CLI Configuration")
		assert.NotContains(t, string(content), "existing content")
	})
}

// Benchmark tests for performance
func BenchmarkCLIHelp(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	binary := buildTestBinaryForBench(b)
	defer os.Remove(binary)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := exec.Command(binary, "--help")
		_, err := cmd.CombinedOutput()
		if err != nil {
			b.Fatalf("Help command failed: %v", err)
		}
	}
}

func buildTestBinaryForBench(b *testing.B) string {
	b.Helper()
	
	tempDir := b.TempDir()
	binaryPath := filepath.Join(tempDir, "assistant-cli-bench")
	
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "main.go")
	buildCmd.Dir = ".." // Set working directory to project root
	err := buildCmd.Run()
	if err != nil {
		b.Fatalf("Failed to build CLI binary: %v", err)
	}
	
	return binaryPath
}
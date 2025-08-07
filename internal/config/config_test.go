package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()
	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}
	if manager.config == nil {
		t.Error("Manager config is nil")
	}
	if manager.viper == nil {
		t.Error("Manager viper is nil")
	}
}

func TestGetDefaults(t *testing.T) {
	defaults := GetDefaults()
	if defaults == nil {
		t.Fatal("GetDefaults() returned nil")
	}
	
	// Test some default values
	if defaults.Auth.Method != "auto" {
		t.Errorf("Expected default auth method 'auto', got '%s'", defaults.Auth.Method)
	}
	if defaults.TTS.Language != "en-US" {
		t.Errorf("Expected default TTS language 'en-US', got '%s'", defaults.TTS.Language)
	}
	if defaults.TTS.SpeakingRate != 1.0 {
		t.Errorf("Expected default speaking rate 1.0, got %f", defaults.TTS.SpeakingRate)
	}
	if defaults.Output.Format != "MP3" {
		t.Errorf("Expected default output format 'MP3', got '%s'", defaults.Output.Format)
	}
}

func TestManagerLoad_Defaults(t *testing.T) {
	manager := NewManager()
	
	if err := manager.Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	
	config := manager.Get()
	if config.Auth.Method != "auto" {
		t.Errorf("Expected auth method 'auto', got '%s'", config.Auth.Method)
	}
	if config.TTS.Language != "en-US" {
		t.Errorf("Expected TTS language 'en-US', got '%s'", config.TTS.Language)
	}
}

func TestManagerLoadWithConfigFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test-config.yaml")
	
	configContent := `
auth:
  method: "apikey"
  timeout: "45s"
  retry_attempts: 5

tts:
  language: "es-ES"
  speaking_rate: 1.5
  pitch: 2.0
  
output:
  format: "LINEAR16"
  auto_filename: true
`
	
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	
	manager := NewManager()
	manager.SetConfigFile(configFile)
	
	if err := manager.Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	
	config := manager.Get()
	
	// Test loaded values
	if config.Auth.Method != "apikey" {
		t.Errorf("Expected auth method 'apikey', got '%s'", config.Auth.Method)
	}
	if config.Auth.Timeout != 45*time.Second {
		t.Errorf("Expected auth timeout 45s, got %v", config.Auth.Timeout)
	}
	if config.Auth.RetryAttempts != 5 {
		t.Errorf("Expected auth retry attempts 5, got %d", config.Auth.RetryAttempts)
	}
	if config.TTS.Language != "es-ES" {
		t.Errorf("Expected TTS language 'es-ES', got '%s'", config.TTS.Language)
	}
	if config.TTS.SpeakingRate != 1.5 {
		t.Errorf("Expected TTS speaking rate 1.5, got %f", config.TTS.SpeakingRate)
	}
	if config.TTS.Pitch != 2.0 {
		t.Errorf("Expected TTS pitch 2.0, got %f", config.TTS.Pitch)
	}
	if config.Output.Format != "LINEAR16" {
		t.Errorf("Expected output format 'LINEAR16', got '%s'", config.Output.Format)
	}
	if !config.Output.AutoFilename {
		t.Error("Expected auto_filename to be true")
	}
}

func TestManagerLoadWithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("ASSISTANT_CLI_TTS_LANGUAGE", "fr-FR")
	os.Setenv("ASSISTANT_CLI_TTS_SPEAKING_RATE", "0.8")
	os.Setenv("ASSISTANT_CLI_OUTPUT_AUTO_FILENAME", "true")
	defer func() {
		os.Unsetenv("ASSISTANT_CLI_TTS_LANGUAGE")
		os.Unsetenv("ASSISTANT_CLI_TTS_SPEAKING_RATE") 
		os.Unsetenv("ASSISTANT_CLI_OUTPUT_AUTO_FILENAME")
	}()
	
	manager := NewManager()
	if err := manager.Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	
	config := manager.Get()
	
	// Test environment variable values
	if config.TTS.Language != "fr-FR" {
		t.Errorf("Expected TTS language 'fr-FR' from env var, got '%s'", config.TTS.Language)
	}
	if config.TTS.SpeakingRate != 0.8 {
		t.Errorf("Expected TTS speaking rate 0.8 from env var, got %f", config.TTS.SpeakingRate)
	}
	if !config.Output.AutoFilename {
		t.Error("Expected auto_filename to be true from env var")
	}
}

func TestGetConfigFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test-config.yaml")
	
	configContent := `
auth:
  method: "apikey"
`
	
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	
	manager := NewManager()
	manager.SetConfigFile(configFile)
	
	if err := manager.Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	
	if path := manager.GetConfigFilePath(); path != configFile {
		t.Errorf("Expected config file path '%s', got '%s'", configFile, path)
	}
}

func TestGenerateExampleConfig(t *testing.T) {
	content := GenerateExampleConfig()
	if content == "" {
		t.Error("GenerateExampleConfig() returned empty string")
	}
	
	// Check that it contains expected sections
	expectedSections := []string{
		"auth:",
		"tts:",
		"output:",
		"playback:",
		"input:",
		"logging:",
		"app:",
	}
	
	for _, section := range expectedSections {
		if !containsString(content, section) {
			t.Errorf("Generated config missing section: %s", section)
		}
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || containsString(s[1:], substr)))
}

func TestValidation_ValidConfig(t *testing.T) {
	manager := NewManager()
	if err := manager.Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	
	if err := manager.Validate(); err != nil {
		t.Errorf("Validation failed for valid default config: %v", err)
	}
}

func TestValidation_InvalidConfig(t *testing.T) {
	manager := NewManager()
	if err := manager.Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	
	// Modify config to have invalid values
	config := manager.Get()
	config.Auth.RetryAttempts = -1  // Invalid
	config.TTS.SpeakingRate = -1.0  // Invalid
	config.TTS.Pitch = 100.0       // Invalid
	
	if err := manager.Validate(); err == nil {
		t.Error("Expected validation to fail for invalid config, but it passed")
	}
}
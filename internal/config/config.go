package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Authentication method constants
const (
	authMethodAuto           = "auto"
	authMethodAPIKey         = "apikey"
	authMethodServiceAccount = "serviceaccount"
	authMethodOAuth2         = "oauth2"
)

// Language constants
const (
	languageEnUS = "en-US"
)

// Config represents the complete configuration structure for the assistant-cli
type Config struct {
	// Authentication settings
	Auth AuthConfig `mapstructure:"auth" yaml:"auth" json:"auth"`

	// Text-to-Speech settings
	TTS TTSConfig `mapstructure:"tts" yaml:"tts" json:"tts"`

	// Output settings
	Output OutputConfig `mapstructure:"output" yaml:"output" json:"output"`

	// Playback settings
	Playback PlaybackConfig `mapstructure:"playback" yaml:"playback" json:"playback"`

	// Input processing settings
	Input InputConfig `mapstructure:"input" yaml:"input" json:"input"`

	// Logging settings
	Logging LoggingConfig `mapstructure:"logging" yaml:"logging" json:"logging"`

	// General application settings
	App AppConfig `mapstructure:"app" yaml:"app" json:"app"`
}

// AuthConfig contains authentication-related configuration
type AuthConfig struct {
	// Preferred authentication method: "apikey", "serviceaccount", "oauth2", "auto"
	Method string `mapstructure:"method" yaml:"method" json:"method" validate:"oneof=apikey serviceaccount oauth2 auto"`

	// API Key for authentication (prefer environment variable)
	APIKey string `mapstructure:"api_key" yaml:"api_key,omitempty" json:"api_key,omitempty"`

	// Path to service account JSON file
	ServiceAccountFile string `mapstructure:"service_account_file" yaml:"service_account_file,omitempty"`

	// OAuth2 client ID (prefer environment variable)
	OAuth2ClientID string `mapstructure:"oauth2_client_id" yaml:"oauth2_client_id,omitempty"`

	// OAuth2 client secret (prefer environment variable)
	OAuth2ClientSecret string `mapstructure:"oauth2_client_secret" yaml:"oauth2_client_secret,omitempty"`

	// OAuth2 token file path
	OAuth2TokenFile string `mapstructure:"oauth2_token_file" yaml:"oauth2_token_file,omitempty"`

	// Connection timeout for authentication
	Timeout time.Duration `mapstructure:"timeout" yaml:"timeout" json:"timeout"`

	// Number of retry attempts for authentication
	RetryAttempts int `mapstructure:"retry_attempts" yaml:"retry_attempts" json:"retry_attempts" validate:"min=0,max=10"`
}

// TTSConfig contains text-to-speech configuration
type TTSConfig struct {
	// Default voice name (e.g., "en-US-Wavenet-D")
	Voice string `mapstructure:"voice" yaml:"voice" json:"voice"`

	// Default language code (e.g., "en-US")
	Language string `mapstructure:"language" yaml:"language" json:"language" validate:"required"`

	// Speaking rate (0.25 to 4.0)
	SpeakingRate float64 `mapstructure:"speaking_rate" yaml:"speaking_rate" validate:"min=0.25,max=4.0"`

	// Voice pitch (-20.0 to 20.0)
	Pitch float64 `mapstructure:"pitch" yaml:"pitch" json:"pitch" validate:"min=-20,max=20"`

	// Volume gain in dB (-96.0 to 16.0)
	VolumeGain float64 `mapstructure:"volume_gain" yaml:"volume_gain" json:"volume_gain" validate:"min=-96,max=16"`

	// Audio encoding format
	AudioEncoding string `mapstructure:"audio_encoding" yaml:"audio_encoding"`

	// Effects profile ID
	EffectsProfile []string `mapstructure:"effects_profile" yaml:"effects_profile" json:"effects_profile"`

	// Request timeout
	Timeout time.Duration `mapstructure:"timeout" yaml:"timeout" json:"timeout"`

	// Maximum retry attempts
	MaxRetries int `mapstructure:"max_retries" yaml:"max_retries" json:"max_retries" validate:"min=0,max=10"`

	// Enable SSML validation
	EnableSSMLValidation bool `mapstructure:"enable_ssml_validation" yaml:"enable_ssml_validation"`
}

// OutputConfig contains output-related configuration
type OutputConfig struct {
	// Default output directory
	DefaultPath string `mapstructure:"default_path" yaml:"default_path" json:"default_path"`

	// Default audio format
	Format string `mapstructure:"format" yaml:"format" validate:"oneof=MP3 LINEAR16 WAV OGG_OPUS MULAW ALAW PCM"`

	// File overwrite behavior: "never", "always", "prompt", "backup"
	OverwriteMode string `mapstructure:"overwrite_mode" yaml:"overwrite_mode" validate:"oneof=never always prompt backup"`

	// File permissions (octal)
	FilePermissions string `mapstructure:"file_permissions" yaml:"file_permissions" json:"file_permissions"`

	// Directory permissions (octal)
	DirPermissions string `mapstructure:"dir_permissions" yaml:"dir_permissions" json:"dir_permissions"`

	// Enable automatic filename generation
	AutoFilename bool `mapstructure:"auto_filename" yaml:"auto_filename" json:"auto_filename"`

	// Maximum filename length
	MaxFilenameLength int `mapstructure:"max_filename_length" yaml:"max_filename_length" validate:"min=10,max=255"`

	// Create directories automatically
	CreateDirs bool `mapstructure:"create_dirs" yaml:"create_dirs" json:"create_dirs"`
}

// PlaybackConfig contains audio playback configuration
type PlaybackConfig struct {
	// Automatically play audio after synthesis
	AutoPlay bool `mapstructure:"auto_play" yaml:"auto_play" json:"auto_play"`

	// Preferred audio player (auto-detected if empty)
	Player string `mapstructure:"player" yaml:"player" json:"player"`

	// Player arguments
	PlayerArgs []string `mapstructure:"player_args" yaml:"player_args" json:"player_args"`

	// Volume level (0.0 to 1.0)
	Volume float64 `mapstructure:"volume" yaml:"volume" json:"volume" validate:"min=0,max=1"`

	// Enable fallback players
	EnableFallback bool `mapstructure:"enable_fallback" yaml:"enable_fallback" json:"enable_fallback"`
}

// InputConfig contains input processing configuration
type InputConfig struct {
	// Maximum text length for processing
	MaxLength int `mapstructure:"max_length" yaml:"max_length" json:"max_length" validate:"min=1,max=100000"`

	// Buffer size for reading input
	BufferSize int `mapstructure:"buffer_size" yaml:"buffer_size" json:"buffer_size" validate:"min=1024,max=65536"`

	// Enable automatic text cleaning
	AutoClean bool `mapstructure:"auto_clean" yaml:"auto_clean" json:"auto_clean"`

	// Enable input validation
	EnableValidation bool `mapstructure:"enable_validation" yaml:"enable_validation" json:"enable_validation"`

	// Enable SSML security validation
	EnableSSMLSecurity bool `mapstructure:"enable_ssml_security" yaml:"enable_ssml_security" json:"enable_ssml_security"`

	// Show input statistics
	ShowStats bool `mapstructure:"show_stats" yaml:"show_stats" json:"show_stats"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	// Log level: "debug", "info", "warn", "error"
	Level string `mapstructure:"level" yaml:"level" json:"level" validate:"oneof=debug info warn error"`

	// Log format: "text", "json"
	Format string `mapstructure:"format" yaml:"format" json:"format" validate:"oneof=text json"`

	// Log output: "stdout", "stderr", or file path
	Output string `mapstructure:"output" yaml:"output" json:"output"`

	// Enable timestamps in logs
	Timestamps bool `mapstructure:"timestamps" yaml:"timestamps" json:"timestamps"`

	// Enable caller information in logs
	Caller bool `mapstructure:"caller" yaml:"caller" json:"caller"`

	// Enable performance logging
	Performance bool `mapstructure:"performance" yaml:"performance" json:"performance"`
}

// AppConfig contains general application configuration
type AppConfig struct {
	// Application name
	Name string `mapstructure:"name" yaml:"name" json:"name"`

	// Configuration file version (for migration)
	ConfigVersion string `mapstructure:"config_version" yaml:"config_version" json:"config_version"`

	// Enable color output
	ColorOutput bool `mapstructure:"color_output" yaml:"color_output" json:"color_output"`

	// Enable progress indicators
	ShowProgress bool `mapstructure:"show_progress" yaml:"show_progress" json:"show_progress"`

	// Quiet mode (minimal output)
	Quiet bool `mapstructure:"quiet" yaml:"quiet" json:"quiet"`

	// Verbose mode (detailed output)
	Verbose bool `mapstructure:"verbose" yaml:"verbose" json:"verbose"`

	// Check for updates
	CheckUpdates bool `mapstructure:"check_updates" yaml:"check_updates" json:"check_updates"`

	// Update check interval
	UpdateCheckInterval time.Duration `mapstructure:"update_check_interval" yaml:"update_check_interval"`
}

// Manager handles configuration loading, validation, and management
type Manager struct {
	config          *Config
	viper           *viper.Viper
	configFileIsSet bool
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{
		config: &Config{},
		viper:  viper.New(),
	}
}

// GetDefaults returns the default configuration values
func GetDefaults() *Config {
	return &Config{
		Auth: AuthConfig{
			Method:        "auto",
			Timeout:       30 * time.Second,
			RetryAttempts: 3,
		},
		TTS: TTSConfig{
			Language:             "en-US",
			SpeakingRate:         1.0,
			Pitch:                0.0,
			VolumeGain:           0.0,
			AudioEncoding:        "MP3",
			EffectsProfile:       []string{"headphone-class-device"},
			Timeout:              30 * time.Second,
			MaxRetries:           3,
			EnableSSMLValidation: true,
		},
		Output: OutputConfig{
			DefaultPath:       ".",
			Format:            "MP3",
			OverwriteMode:     "backup",
			FilePermissions:   "0644",
			DirPermissions:    "0755",
			AutoFilename:      false,
			MaxFilenameLength: 100,
			CreateDirs:        true,
		},
		Playback: PlaybackConfig{
			AutoPlay:       false,
			Volume:         1.0,
			EnableFallback: true,
		},
		Input: InputConfig{
			MaxLength:          5000,
			BufferSize:         4096,
			AutoClean:          true,
			EnableValidation:   true,
			EnableSSMLSecurity: true,
			ShowStats:          false,
		},
		Logging: LoggingConfig{
			Level:       "info",
			Format:      "text",
			Output:      "stderr",
			Timestamps:  true,
			Caller:      false,
			Performance: false,
		},
		App: AppConfig{
			Name:                "assistant-cli",
			ConfigVersion:       "1.5.0",
			ColorOutput:         true,
			ShowProgress:        true,
			Quiet:               false,
			Verbose:             false,
			CheckUpdates:        true,
			UpdateCheckInterval: 24 * time.Hour,
		},
	}
}

// Load loads configuration from various sources with proper precedence
// SetConfigFile sets a specific config file path to load
func (m *Manager) SetConfigFile(path string) {
	m.viper.SetConfigFile(path)
	m.configFileIsSet = true
}

// GetViper returns the underlying viper instance for backward compatibility
func (m *Manager) GetViper() *viper.Viper {
	return m.viper
}

func (m *Manager) Load() error {
	// Set defaults
	defaults := GetDefaults()
	m.setDefaults(defaults)

	// Configure viper
	m.viper.SetEnvPrefix("ASSISTANT_CLI")
	m.viper.AutomaticEnv()
	m.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Only set config search paths if no specific config file was set
	if !m.configFileIsSet {
		// Set config file search paths
		m.viper.SetConfigName(".assistant-cli")
		m.viper.SetConfigType("yaml")

		// Add search paths
		if home, err := os.UserHomeDir(); err == nil {
			m.viper.AddConfigPath(home)
		}
		m.viper.AddConfigPath(".")
	}

	// Try to read config file
	if err := m.viper.ReadInConfig(); err != nil {
		// Config file not found is not an error
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal into config struct
	if err := m.viper.Unmarshal(m.config); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	if err := m.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	return nil
}

// setDefaults sets default values in viper
func (m *Manager) setDefaults(config *Config) {
	// Use reflection to set defaults from the struct
	m.setDefaultsRecursive("", reflect.ValueOf(config).Elem(), reflect.TypeOf(config).Elem())
}

// setDefaultsRecursive recursively sets default values using reflection
func (m *Manager) setDefaultsRecursive(prefix string, v reflect.Value, t reflect.Type) {
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		m.processField(prefix, field, fieldType)
	}
}

func (m *Manager) processField(prefix string, field reflect.Value, fieldType reflect.StructField) {
	// Get the mapstructure tag
	tag := fieldType.Tag.Get("mapstructure")
	if tag == "" {
		return
	}

	key := m.buildKey(prefix, tag)
	m.setFieldDefault(key, field, fieldType)
}

func (m *Manager) buildKey(prefix, tag string) string {
	if prefix != "" {
		return prefix + "." + tag
	}
	return tag
}

func (m *Manager) setFieldDefault(key string, field reflect.Value, fieldType reflect.StructField) {
	// Handle different field types
	switch field.Kind() {
	case reflect.Struct:
		// Recursively handle nested structs
		m.setDefaultsRecursive(key, field, fieldType.Type)
	case reflect.String:
		m.setStringDefault(key, field)
	case reflect.Int, reflect.Int64:
		m.setIntDefault(key, field, fieldType)
	case reflect.Float64:
		m.setFloatDefault(key, field)
	case reflect.Bool:
		m.viper.SetDefault(key, field.Bool())
	case reflect.Slice:
		m.setSliceDefault(key, field)
	}
}

func (m *Manager) setStringDefault(key string, field reflect.Value) {
	if field.String() != "" {
		m.viper.SetDefault(key, field.String())
	}
}

func (m *Manager) setIntDefault(key string, field reflect.Value, fieldType reflect.StructField) {
	if field.Int() != 0 {
		if fieldType.Type == reflect.TypeOf(time.Duration(0)) {
			// Handle time.Duration specially
			m.viper.SetDefault(key, field.Interface())
		} else {
			m.viper.SetDefault(key, field.Int())
		}
	}
}

func (m *Manager) setFloatDefault(key string, field reflect.Value) {
	if field.Float() != 0.0 {
		m.viper.SetDefault(key, field.Float())
	}
}

func (m *Manager) setSliceDefault(key string, field reflect.Value) {
	if !field.IsNil() && field.Len() > 0 {
		m.viper.SetDefault(key, field.Interface())
	}
}

// Validate validates the configuration
func (m *Manager) Validate() error {
	// Use the comprehensive validation from ValidateComprehensive
	// but return a simple error instead of ValidationErrors
	if err := m.ValidateComprehensive(); err != nil {
		return err
	}
	return nil
}

// Get returns the current configuration
func (m *Manager) Get() *Config {
	return m.config
}

// GetConfigFilePath returns the path of the config file being used
func (m *Manager) GetConfigFilePath() string {
	return m.viper.ConfigFileUsed()
}

// SaveConfig saves the current configuration to a file
func (m *Manager) SaveConfig(path string) error {
	if path == "" {
		path = m.getDefaultConfigPath()
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return m.viper.WriteConfigAs(path)
}

// getDefaultConfigPath returns the default configuration file path
func (m *Manager) getDefaultConfigPath() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".assistant-cli.yaml")
	}
	return ".assistant-cli.yaml"
}

// GenerateExampleConfig generates an example configuration file with comments
func GenerateExampleConfig() string {
	return `# Assistant-CLI Configuration File
# This file contains all available configuration options with their default values

# Authentication settings
auth:
  # Authentication method: "auto", "apikey", "serviceaccount", "oauth2"
  method: "auto"
  
  # Connection timeout for authentication requests
  timeout: "30s"
  
  # Number of retry attempts for authentication
  retry_attempts: 3
  
  # Note: Sensitive credentials should be set via environment variables:
  # ASSISTANT_CLI_API_KEY="your-api-key"
  # GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"
  # ASSISTANT_CLI_OAUTH2_CLIENT_ID="your-client-id"
  # ASSISTANT_CLI_OAUTH2_CLIENT_SECRET="your-client-secret"

# Text-to-Speech settings
tts:
  # Default language code (required)
  language: "en-US"
  
  # Default voice name (optional, will use language default if not set)
  # voice: "en-US-Wavenet-D"
  
  # Speaking rate (0.25 to 4.0)
  speaking_rate: 1.0
  
  # Voice pitch (-20.0 to 20.0)
  pitch: 0.0
  
  # Volume gain in dB (-96.0 to 16.0)
  volume_gain: 0.0
  
  # Audio encoding format
  audio_encoding: "MP3"
  
  # Effects profile for enhanced audio
  effects_profile:
    - "headphone-class-device"
  
  # Request timeout
  timeout: "30s"
  
  # Maximum retry attempts
  max_retries: 3
  
  # Enable SSML validation
  enable_ssml_validation: true

# Output settings
output:
  # Default output directory
  default_path: "."
  
  # Default audio format
  format: "MP3"
  
  # File overwrite behavior: "never", "always", "prompt", "backup"
  overwrite_mode: "backup"
  
  # File permissions (octal notation)
  file_permissions: "0644"
  
  # Directory permissions (octal notation)
  dir_permissions: "0755"
  
  # Enable automatic filename generation from input text
  auto_filename: false
  
  # Maximum filename length
  max_filename_length: 100
  
  # Create directories automatically
  create_dirs: true

# Audio playback settings
playback:
  # Automatically play audio after synthesis
  auto_play: false
  
  # Preferred audio player (auto-detected if empty)
  # player: ""
  
  # Additional player arguments
  # player_args: []
  
  # Volume level (0.0 to 1.0)
  volume: 1.0
  
  # Enable fallback players if primary player fails
  enable_fallback: true

# Input processing settings
input:
  # Maximum text length for processing
  max_length: 5000
  
  # Buffer size for reading input
  buffer_size: 4096
  
  # Enable automatic text cleaning
  auto_clean: true
  
  # Enable input validation
  enable_validation: true
  
  # Enable SSML security validation
  enable_ssml_security: true
  
  # Show input statistics
  show_stats: false

# Logging settings
logging:
  # Log level: "debug", "info", "warn", "error"
  level: "info"
  
  # Log format: "text", "json"
  format: "text"
  
  # Log output: "stdout", "stderr", or file path
  output: "stderr"
  
  # Enable timestamps in logs
  timestamps: true
  
  # Enable caller information in logs
  caller: false
  
  # Enable performance logging
  performance: false

# Application settings
app:
  # Application name
  name: "assistant-cli"
  
  # Configuration file version (for migration)
  config_version: "1.5.0"
  
  # Enable color output
  color_output: true
  
  # Show progress indicators
  show_progress: true
  
  # Quiet mode (minimal output)
  quiet: false
  
  # Verbose mode (detailed output)
  verbose: false
  
  # Check for updates
  check_updates: true
  
  # Update check interval
  update_check_interval: "24h"
`
}

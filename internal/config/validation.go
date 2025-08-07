package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("config validation error for field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []*ValidationError

func (ve ValidationErrors) Error() string {
	messages := make([]string, 0, len(ve))
	for _, err := range ve {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// ValidateComprehensive performs comprehensive validation of the configuration
func (m *Manager) ValidateComprehensive() error {
	var errors ValidationErrors
	
	config := m.config
	
	// Validate Auth configuration
	if authErrors := m.validateAuth(&config.Auth); authErrors != nil {
		errors = append(errors, authErrors...)
	}
	
	// Validate TTS configuration
	if ttsErrors := m.validateTTS(&config.TTS); ttsErrors != nil {
		errors = append(errors, ttsErrors...)
	}
	
	// Validate Output configuration
	if outputErrors := m.validateOutput(&config.Output); outputErrors != nil {
		errors = append(errors, outputErrors...)
	}
	
	// Validate Playback configuration
	if playbackErrors := m.validatePlayback(&config.Playback); playbackErrors != nil {
		errors = append(errors, playbackErrors...)
	}
	
	// Validate Input configuration
	if inputErrors := m.validateInput(&config.Input); inputErrors != nil {
		errors = append(errors, inputErrors...)
	}
	
	// Validate Logging configuration
	if loggingErrors := m.validateLogging(&config.Logging); loggingErrors != nil {
		errors = append(errors, loggingErrors...)
	}
	
	// Validate App configuration
	if appErrors := m.validateApp(&config.App); appErrors != nil {
		errors = append(errors, appErrors...)
	}
	
	if len(errors) > 0 {
		return errors
	}
	
	return nil
}

// validateAuth validates authentication configuration
func (m *Manager) validateAuth(auth *AuthConfig) []*ValidationError {
	var errors []*ValidationError
	
	// Validate method
	validMethods := []string{"auto", "apikey", "serviceaccount", "oauth2"}
	if auth.Method != "" && !contains(validMethods, auth.Method) {
		errors = append(errors, &ValidationError{
			Field:   "auth.method",
			Value:   auth.Method,
			Message: fmt.Sprintf("must be one of: %s", strings.Join(validMethods, ", ")),
		})
	}
	
	// Validate service account file if specified
	if auth.ServiceAccountFile != "" {
		if !filepath.IsAbs(auth.ServiceAccountFile) && !strings.HasPrefix(auth.ServiceAccountFile, "~") {
			errors = append(errors, &ValidationError{
				Field:   "auth.service_account_file",
				Value:   auth.ServiceAccountFile,
				Message: "must be an absolute path",
			})
		}
		
		if _, err := os.Stat(expandPath(auth.ServiceAccountFile)); os.IsNotExist(err) {
			errors = append(errors, &ValidationError{
				Field:   "auth.service_account_file",
				Value:   auth.ServiceAccountFile,
				Message: "file does not exist",
			})
		}
	}
	
	// Validate OAuth2 token file if specified
	if auth.OAuth2TokenFile != "" {
		dir := filepath.Dir(expandPath(auth.OAuth2TokenFile))
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			errors = append(errors, &ValidationError{
				Field:   "auth.oauth2_token_file",
				Value:   auth.OAuth2TokenFile,
				Message: "directory does not exist",
			})
		}
	}
	
	// Validate timeout
	if auth.Timeout < 0 {
		errors = append(errors, &ValidationError{
			Field:   "auth.timeout",
			Value:   auth.Timeout,
			Message: "must be non-negative",
		})
	}
	if auth.Timeout > 5*time.Minute {
		errors = append(errors, &ValidationError{
			Field:   "auth.timeout",
			Value:   auth.Timeout,
			Message: "timeout too long (max 5 minutes)",
		})
	}
	
	// Validate retry attempts
	if auth.RetryAttempts < 0 || auth.RetryAttempts > 10 {
		errors = append(errors, &ValidationError{
			Field:   "auth.retry_attempts",
			Value:   auth.RetryAttempts,
			Message: "must be between 0 and 10",
		})
	}
	
	return errors
}

// validateTTS validates TTS configuration
func (m *Manager) validateTTS(tts *TTSConfig) []*ValidationError {
	var errors []*ValidationError
	
	// Validate language (required)
	if tts.Language == "" {
		errors = append(errors, &ValidationError{
			Field:   "tts.language",
			Value:   tts.Language,
			Message: "is required",
		})
	} else if !isValidLanguageCode(tts.Language) {
		errors = append(errors, &ValidationError{
			Field:   "tts.language",
			Value:   tts.Language,
			Message: "invalid language code format (expected format: en-US)",
		})
	}
	
	// Validate speaking rate
	if tts.SpeakingRate < 0.25 || tts.SpeakingRate > 4.0 {
		errors = append(errors, &ValidationError{
			Field:   "tts.speaking_rate",
			Value:   tts.SpeakingRate,
			Message: "must be between 0.25 and 4.0",
		})
	}
	
	// Validate pitch
	if tts.Pitch < -20.0 || tts.Pitch > 20.0 {
		errors = append(errors, &ValidationError{
			Field:   "tts.pitch",
			Value:   tts.Pitch,
			Message: "must be between -20.0 and 20.0",
		})
	}
	
	// Validate volume gain
	if tts.VolumeGain < -96.0 || tts.VolumeGain > 16.0 {
		errors = append(errors, &ValidationError{
			Field:   "tts.volume_gain",
			Value:   tts.VolumeGain,
			Message: "must be between -96.0 and 16.0",
		})
	}
	
	// Validate audio encoding
	validEncodings := []string{"MP3", "LINEAR16", "OGG_OPUS", "MULAW", "ALAW", "PCM"}
	if tts.AudioEncoding != "" && !contains(validEncodings, tts.AudioEncoding) {
		errors = append(errors, &ValidationError{
			Field:   "tts.audio_encoding",
			Value:   tts.AudioEncoding,
			Message: fmt.Sprintf("must be one of: %s", strings.Join(validEncodings, ", ")),
		})
	}
	
	// Validate timeout
	if tts.Timeout < 0 {
		errors = append(errors, &ValidationError{
			Field:   "tts.timeout",
			Value:   tts.Timeout,
			Message: "must be non-negative",
		})
	}
	if tts.Timeout > 10*time.Minute {
		errors = append(errors, &ValidationError{
			Field:   "tts.timeout",
			Value:   tts.Timeout,
			Message: "timeout too long (max 10 minutes)",
		})
	}
	
	// Validate max retries
	if tts.MaxRetries < 0 || tts.MaxRetries > 10 {
		errors = append(errors, &ValidationError{
			Field:   "tts.max_retries",
			Value:   tts.MaxRetries,
			Message: "must be between 0 and 10",
		})
	}
	
	return errors
}

// validateOutput validates output configuration
func (m *Manager) validateOutput(output *OutputConfig) []*ValidationError {
	var errors []*ValidationError
	
	// Validate default path
	if output.DefaultPath != "" {
		// Check if it's a valid path format
		if strings.ContainsAny(output.DefaultPath, "<>:\"|?*") {
			errors = append(errors, &ValidationError{
				Field:   "output.default_path",
				Value:   output.DefaultPath,
				Message: "contains invalid path characters",
			})
		}
	}
	
	// Validate format
	validFormats := []string{"MP3", "LINEAR16", "WAV", "OGG_OPUS", "MULAW", "ALAW", "PCM"}
	if output.Format != "" && !contains(validFormats, output.Format) {
		errors = append(errors, &ValidationError{
			Field:   "output.format",
			Value:   output.Format,
			Message: fmt.Sprintf("must be one of: %s", strings.Join(validFormats, ", ")),
		})
	}
	
	// Validate overwrite mode
	validModes := []string{"never", "always", "prompt", "backup"}
	if output.OverwriteMode != "" && !contains(validModes, output.OverwriteMode) {
		errors = append(errors, &ValidationError{
			Field:   "output.overwrite_mode",
			Value:   output.OverwriteMode,
			Message: fmt.Sprintf("must be one of: %s", strings.Join(validModes, ", ")),
		})
	}
	
	// Validate file permissions
	if output.FilePermissions != "" {
		if err := validateOctalPermissions(output.FilePermissions); err != nil {
			errors = append(errors, &ValidationError{
				Field:   "output.file_permissions",
				Value:   output.FilePermissions,
				Message: err.Error(),
			})
		}
	}
	
	// Validate directory permissions
	if output.DirPermissions != "" {
		if err := validateOctalPermissions(output.DirPermissions); err != nil {
			errors = append(errors, &ValidationError{
				Field:   "output.dir_permissions",
				Value:   output.DirPermissions,
				Message: err.Error(),
			})
		}
	}
	
	// Validate max filename length
	if output.MaxFilenameLength < 10 || output.MaxFilenameLength > 255 {
		errors = append(errors, &ValidationError{
			Field:   "output.max_filename_length",
			Value:   output.MaxFilenameLength,
			Message: "must be between 10 and 255",
		})
	}
	
	return errors
}

// validatePlayback validates playback configuration
func (m *Manager) validatePlayback(playback *PlaybackConfig) []*ValidationError {
	var errors []*ValidationError
	
	// Validate volume
	if playback.Volume < 0.0 || playback.Volume > 1.0 {
		errors = append(errors, &ValidationError{
			Field:   "playback.volume",
			Value:   playback.Volume,
			Message: "must be between 0.0 and 1.0",
		})
	}
	
	// Validate player if specified
	if playback.Player != "" {
		// Check if it's a valid command (basic validation)
		if strings.ContainsAny(playback.Player, "<>:\"|?*") {
			errors = append(errors, &ValidationError{
				Field:   "playback.player",
				Value:   playback.Player,
				Message: "contains invalid characters for a command",
			})
		}
	}
	
	return errors
}

// validateInput validates input configuration
func (m *Manager) validateInput(input *InputConfig) []*ValidationError {
	var errors []*ValidationError
	
	// Validate max length
	if input.MaxLength <= 0 || input.MaxLength > 100000 {
		errors = append(errors, &ValidationError{
			Field:   "input.max_length",
			Value:   input.MaxLength,
			Message: "must be between 1 and 100000",
		})
	}
	
	// Validate buffer size
	if input.BufferSize < 1024 || input.BufferSize > 65536 {
		errors = append(errors, &ValidationError{
			Field:   "input.buffer_size",
			Value:   input.BufferSize,
			Message: "must be between 1024 and 65536",
		})
	}
	
	return errors
}

// validateLogging validates logging configuration
func (m *Manager) validateLogging(logging *LoggingConfig) []*ValidationError {
	var errors []*ValidationError
	
	// Validate level
	validLevels := []string{"debug", "info", "warn", "error"}
	if logging.Level != "" && !contains(validLevels, logging.Level) {
		errors = append(errors, &ValidationError{
			Field:   "logging.level",
			Value:   logging.Level,
			Message: fmt.Sprintf("must be one of: %s", strings.Join(validLevels, ", ")),
		})
	}
	
	// Validate format
	validFormats := []string{"text", "json"}
	if logging.Format != "" && !contains(validFormats, logging.Format) {
		errors = append(errors, &ValidationError{
			Field:   "logging.format",
			Value:   logging.Format,
			Message: fmt.Sprintf("must be one of: %s", strings.Join(validFormats, ", ")),
		})
	}
	
	// Validate output
	if logging.Output != "" && logging.Output != "stdout" && logging.Output != "stderr" {
		// If it's not stdout/stderr, treat it as a file path
		expandedPath := expandPath(logging.Output)
		dir := filepath.Dir(expandedPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			errors = append(errors, &ValidationError{
				Field:   "logging.output",
				Value:   logging.Output,
				Message: "directory does not exist",
			})
		}
	}
	
	return errors
}

// validateApp validates app configuration
func (m *Manager) validateApp(app *AppConfig) []*ValidationError {
	var errors []*ValidationError
	
	// Validate config version format
	if app.ConfigVersion != "" {
		if !isValidSemanticVersion(app.ConfigVersion) {
			errors = append(errors, &ValidationError{
				Field:   "app.config_version",
				Value:   app.ConfigVersion,
				Message: "invalid semantic version format",
			})
		}
	}
	
	// Validate update check interval
	if app.UpdateCheckInterval < 0 {
		errors = append(errors, &ValidationError{
			Field:   "app.update_check_interval",
			Value:   app.UpdateCheckInterval,
			Message: "must be non-negative",
		})
	}
	if app.UpdateCheckInterval > 0 && app.UpdateCheckInterval < time.Hour {
		errors = append(errors, &ValidationError{
			Field:   "app.update_check_interval",
			Value:   app.UpdateCheckInterval,
			Message: "minimum interval is 1 hour",
		})
	}
	
	return errors
}

// Helper functions

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			return strings.Replace(path, "~", home, 1)
		}
	}
	return path
}

// isValidLanguageCode checks if a string is a valid language code (e.g., en-US)
func isValidLanguageCode(code string) bool {
	// Simple validation for language-COUNTRY format
	matched, _ := regexp.MatchString(`^[a-z]{2}-[A-Z]{2}$`, code)
	return matched
}

// validateOctalPermissions validates octal permission strings
func validateOctalPermissions(perms string) error {
	if len(perms) != 4 || !strings.HasPrefix(perms, "0") {
		return fmt.Errorf("must be 4-digit octal (e.g., 0644)")
	}
	
	if _, err := strconv.ParseUint(perms, 8, 32); err != nil {
		return fmt.Errorf("invalid octal format")
	}
	
	return nil
}

// isValidSemanticVersion checks if a string is a valid semantic version
func isValidSemanticVersion(version string) bool {
	// Simple semantic version validation (major.minor.patch)
	matched, _ := regexp.MatchString(`^\d+\.\d+\.\d+$`, version)
	return matched
}
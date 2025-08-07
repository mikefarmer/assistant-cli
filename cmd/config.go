package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mikefarmer/assistant-cli/internal/config"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration settings",
	Long: `Manage configuration settings for assistant-cli.

This command provides utilities to generate, validate, and view configuration files.
It supports creating example configuration files, viewing current settings, and
validating configuration for errors.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, show help
		_ = cmd.Help()
	},
}

var generateConfigCmd = &cobra.Command{
	Use:   "generate [output-path]",
	Short: "Generate an example configuration file",
	Long: `Generate an example configuration file with all available options and their default values.

If no output path is specified, the default location (~/.assistant-cli.yaml) will be used.
The generated file includes comprehensive comments explaining each configuration option.

Examples:
  assistant-cli config generate
  assistant-cli config generate ./my-config.yaml
  assistant-cli config generate ~/.config/assistant-cli.yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGenerateConfig,
}

var validateConfigCmd = &cobra.Command{
	Use:   "validate [config-file]",
	Short: "Validate configuration file",
	Long: `Validate the configuration file for errors and inconsistencies.

If no config file is specified, the default configuration locations will be checked.
This command will report any validation errors, missing required values, or
configuration inconsistencies.

Examples:
  assistant-cli config validate
  assistant-cli config validate ~/.assistant-cli.yaml
  assistant-cli config validate ./custom-config.yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: runValidateConfig,
}

var showConfigCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long: `Show the current effective configuration loaded from all sources.

This displays the complete configuration as it would be used by the application,
including values from environment variables, configuration files, and defaults.
Sensitive values like API keys and secrets will be masked for security.

Examples:
  assistant-cli config show
  assistant-cli config show --format json
  assistant-cli config show --include-defaults`,
	RunE: runShowConfig,
}

var (
	generateForce  bool
	generateFormat string
	showFormat     string
	showDefaults   bool
	showSources    bool
	maskSensitive  bool
)

func init() {
	// Add subcommands
	configCmd.AddCommand(generateConfigCmd)
	configCmd.AddCommand(validateConfigCmd)
	configCmd.AddCommand(showConfigCmd)

	// Generate command flags
	generateConfigCmd.Flags().BoolVarP(&generateForce, "force", "f", false, "Overwrite existing config file")
	generateConfigCmd.Flags().StringVar(&generateFormat, "format", "yaml", "Output format (yaml, json)")

	// Show command flags
	showConfigCmd.Flags().StringVar(&showFormat, "format", "yaml", "Output format (yaml, json, table)")
	showConfigCmd.Flags().BoolVar(&showDefaults, "include-defaults", false, "Include default values")
	showConfigCmd.Flags().BoolVar(&showSources, "show-sources", false, "Show configuration sources")
	showConfigCmd.Flags().BoolVar(&maskSensitive, "mask-sensitive", true, "Mask sensitive values")
}

func runGenerateConfig(cmd *cobra.Command, args []string) error {
	var outputPath string

	if len(args) > 0 {
		outputPath = args[0]
	} else {
		// Use default config path
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		outputPath = filepath.Join(home, ".assistant-cli.yaml")
	}

	// Expand tilde if present
	if strings.HasPrefix(outputPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		outputPath = filepath.Join(home, outputPath[1:])
	}

	// Check if file exists and handle overwrite
	if _, err := os.Stat(outputPath); err == nil {
		if !generateForce {
			return fmt.Errorf("config file already exists at %s (use --force to overwrite)", outputPath)
		}
	}

	// Create directory if needed
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Generate content based on format
	var content string
	switch generateFormat {
	case "yaml":
		content = config.GenerateExampleConfig()
	case "json":
		// For JSON, we would need to marshal the defaults struct
		// For now, fall back to YAML format
		content = config.GenerateExampleConfig()
		fmt.Fprintf(os.Stderr, "Warning: JSON format not yet implemented, generating YAML instead\n")
	default:
		return fmt.Errorf("unsupported format: %s (supported: yaml, json)", generateFormat)
	}

	// Write the file
	if err := os.WriteFile(outputPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✓ Generated configuration file: %s\n", outputPath)
	fmt.Printf("Edit the file to customize your settings, then run 'assistant-cli config validate' to check for errors.\n")

	return nil
}

func runValidateConfig(cmd *cobra.Command, args []string) error {
	var configFile string

	if len(args) > 0 {
		configFile = args[0]
	}

	// Create config manager and load configuration
	manager := config.NewManager()
	if configFile != "" {
		manager.SetConfigFile(configFile)
	}
	if err := manager.Load(); err != nil {
		fmt.Printf("❌ Configuration validation failed: %v\n", err)
		return err
	}

	// Perform comprehensive validation
	if err := manager.ValidateComprehensive(); err != nil {
		fmt.Printf("❌ Configuration validation failed:\n")

		if validationErrors, ok := err.(config.ValidationErrors); ok {
			for i, validationErr := range validationErrors {
				fmt.Printf("  %d. %s\n", i+1, validationErr.Error())
			}
		} else {
			fmt.Printf("  %v\n", err)
		}
		return err
	}

	configPath := manager.GetConfigFilePath()
	if configPath == "" {
		fmt.Printf("✓ Configuration validation passed (using defaults)\n")
		fmt.Printf("Note: No configuration file found. Run 'assistant-cli config generate' to create one.\n")
	} else {
		fmt.Printf("✓ Configuration validation passed: %s\n", configPath)
	}

	return nil
}

func runShowConfig(cmd *cobra.Command, args []string) error {
	// Use the global config manager
	manager := GetConfig()

	config := manager.Get()

	switch showFormat {
	case "yaml":
		return showConfigYAML(config, manager)
	case "json":
		return showConfigJSON(config, manager)
	case "table":
		return showConfigTable(config, manager)
	default:
		return fmt.Errorf("unsupported format: %s (supported: yaml, json, table)", showFormat)
	}
}

func showConfigYAML(cfg *config.Config, manager *config.Manager) error {
	// Create a copy for display and mask sensitive values if requested
	displayConfig := *cfg

	if maskSensitive {
		maskSensitiveValues(&displayConfig)
	}

	// Print config file source info
	if showSources {
		configPath := manager.GetConfigFilePath()
		if configPath != "" {
			fmt.Printf("# Configuration loaded from: %s\n", configPath)
		} else {
			fmt.Printf("# Configuration: using defaults (no config file found)\n")
		}
		fmt.Printf("# Environment variables with prefix: ASSISTANT_CLI_\n\n")
	}

	// For now, let's manually format key sections
	fmt.Println("# Current Configuration")
	fmt.Println("auth:")
	fmt.Printf("  method: %q\n", displayConfig.Auth.Method)
	fmt.Printf("  timeout: %q\n", displayConfig.Auth.Timeout.String())
	fmt.Printf("  retry_attempts: %d\n", displayConfig.Auth.RetryAttempts)

	fmt.Println("\ntts:")
	fmt.Printf("  language: %q\n", displayConfig.TTS.Language)
	if displayConfig.TTS.Voice != "" {
		fmt.Printf("  voice: %q\n", displayConfig.TTS.Voice)
	}
	fmt.Printf("  speaking_rate: %.2f\n", displayConfig.TTS.SpeakingRate)
	fmt.Printf("  pitch: %.2f\n", displayConfig.TTS.Pitch)
	fmt.Printf("  volume_gain: %.2f\n", displayConfig.TTS.VolumeGain)
	fmt.Printf("  audio_encoding: %q\n", displayConfig.TTS.AudioEncoding)

	fmt.Println("\noutput:")
	fmt.Printf("  default_path: %q\n", displayConfig.Output.DefaultPath)
	fmt.Printf("  format: %q\n", displayConfig.Output.Format)
	fmt.Printf("  overwrite_mode: %q\n", displayConfig.Output.OverwriteMode)
	fmt.Printf("  auto_filename: %t\n", displayConfig.Output.AutoFilename)

	fmt.Println("\nplayback:")
	fmt.Printf("  auto_play: %t\n", displayConfig.Playback.AutoPlay)
	fmt.Printf("  volume: %.2f\n", displayConfig.Playback.Volume)

	return nil
}

func showConfigJSON(cfg *config.Config, manager *config.Manager) error {
	fmt.Println("JSON format display not yet implemented")
	fmt.Println("Please use --format yaml or --format table")
	return nil
}

func showConfigTable(cfg *config.Config, manager *config.Manager) error {
	_ = manager // Manager parameter not used in table format

	displayConfig := *cfg

	if maskSensitive {
		maskSensitiveValues(&displayConfig)
	}

	fmt.Printf("%-30s %-20s %s\n", "Setting", "Value", "Source")
	fmt.Printf("%-30s %-20s %s\n", "-------", "-----", "------")

	// Auth settings
	fmt.Printf("%-30s %-20s %s\n", "auth.method", displayConfig.Auth.Method, getValueSource("auth.method"))
	fmt.Printf("%-30s %-20s %s\n", "auth.timeout", displayConfig.Auth.Timeout.String(), getValueSource("auth.timeout"))
	fmt.Printf("%-30s %-20d %s\n", "auth.retry_attempts", displayConfig.Auth.RetryAttempts,
		getValueSource("auth.retry_attempts"))

	// TTS settings
	fmt.Printf("%-30s %-20s %s\n", "tts.language", displayConfig.TTS.Language, getValueSource("tts.language"))
	if displayConfig.TTS.Voice != "" {
		fmt.Printf("%-30s %-20s %s\n", "tts.voice", displayConfig.TTS.Voice, getValueSource("tts.voice"))
	}
	fmt.Printf("%-30s %-20.2f %s\n", "tts.speaking_rate", displayConfig.TTS.SpeakingRate,
		getValueSource("tts.speaking_rate"))
	fmt.Printf("%-30s %-20.2f %s\n", "tts.pitch", displayConfig.TTS.Pitch, getValueSource("tts.pitch"))
	fmt.Printf("%-30s %-20.2f %s\n", "tts.volume_gain", displayConfig.TTS.VolumeGain,
		getValueSource("tts.volume_gain"))
	fmt.Printf("%-30s %-20s %s\n", "tts.audio_encoding", displayConfig.TTS.AudioEncoding,
		getValueSource("tts.audio_encoding"))

	// Output settings
	fmt.Printf("%-30s %-20s %s\n", "output.default_path", displayConfig.Output.DefaultPath,
		getValueSource("output.default_path"))
	fmt.Printf("%-30s %-20s %s\n", "output.format", displayConfig.Output.Format, getValueSource("output.format"))
	fmt.Printf("%-30s %-20s %s\n", "output.overwrite_mode", displayConfig.Output.OverwriteMode,
		getValueSource("output.overwrite_mode"))
	fmt.Printf("%-30s %-20t %s\n", "output.auto_filename", displayConfig.Output.AutoFilename,
		getValueSource("output.auto_filename"))

	// Playback settings
	fmt.Printf("%-30s %-20t %s\n", "playback.auto_play", displayConfig.Playback.AutoPlay,
		getValueSource("playback.auto_play"))
	fmt.Printf("%-30s %-20.2f %s\n", "playback.volume", displayConfig.Playback.Volume, getValueSource("playback.volume"))

	return nil
}

func maskSensitiveValues(cfg *config.Config) {
	// Mask sensitive configuration values
	if cfg.Auth.APIKey != "" {
		cfg.Auth.APIKey = "***masked***"
	}
	if cfg.Auth.OAuth2ClientSecret != "" {
		cfg.Auth.OAuth2ClientSecret = "***masked***"
	}
}

func getValueSource(key string) string {
	// This is a simplified source detection - in a real implementation,
	// we would track where each value came from during loading
	if os.Getenv("ASSISTANT_CLI_"+strings.ToUpper(strings.ReplaceAll(key, ".", "_"))) != "" {
		return "environment"
	}
	// For now, assume others are from config file or defaults
	return "config/default"
}

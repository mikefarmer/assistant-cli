package cmd

import (
	"fmt"
	"os"

	"github.com/mikefarmer/assistant-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile       string
	globalConfig  *config.Manager
)

var version = "dev" // This will be set by build flags

// SetVersion sets the version for the CLI
func SetVersion(v string) {
	version = v
}

// NewRootCmd creates and returns the root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "assistant-cli",
		Short:   "A personal assistant CLI tool",
		Long: `A personal assistant CLI tool with various capabilities.

This tool provides text-to-speech conversion using Google Cloud Text-to-Speech API,
and will be extended with additional features for Calendar, Gmail, and Drive integration.
It supports multiple authentication methods and provides various customization options.`,
		Example: `  # Basic text-to-speech
  echo "Hello, World!" | assistant-cli synthesize -o hello.mp3

  # Custom voice with playback
  echo "Welcome!" | assistant-cli synthesize --voice en-US-Wavenet-C --play

  # Set up authentication
  export ASSISTANT_CLI_API_KEY="your-api-key"
  assistant-cli login --validate

  # List available voices
  assistant-cli synthesize --list-voices --language en-US

  # Use configuration file
  assistant-cli config generate ~/.assistant-cli.yaml
  assistant-cli --config ~/.assistant-cli.yaml synthesize --help`,
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			// If no subcommand is provided, show help
			cmd.Help()
		},
	}

	// Set up persistent flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.assistant-cli.yaml)")

	// Initialize config when root command is created
	cobra.OnInitialize(initConfig)

	// Add subcommands
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(NewSynthesizeCmd())
	rootCmd.AddCommand(configCmd)

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Initialize the global config manager
	globalConfig = config.NewManager()
	
	// If a specific config file is provided, set it
	if cfgFile != "" {
		globalConfig.SetConfigFile(cfgFile)
	}
	
	// Load the configuration
	if err := globalConfig.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		// Don't exit here, as the app can still work with defaults
	}

	// Keep the old viper functionality for backward compatibility
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Search config in home directory with name ".assistant-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".assistant-cli")
	}

	// Set up environment variable prefix
	viper.SetEnvPrefix("ASSISTANT_CLI")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// Only print if we're in verbose mode (to be implemented)
		// fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// GetConfig returns the global configuration manager
func GetConfig() *config.Manager {
	if globalConfig == nil {
		globalConfig = config.NewManager()
		if err := globalConfig.Load(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error loading configuration: %v\n", err)
		}
	}
	return globalConfig
}
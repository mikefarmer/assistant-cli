package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	texttospeechpb "cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/mikefarmer/assistant-cli/internal/auth"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Google Cloud services",
	Long: `Authenticate with Google Cloud Text-to-Speech API using one of the supported methods:

• API Key authentication (simplest)
• Service Account authentication (for automation)
• OAuth2 authentication (interactive browser flow)

The tool will guide you through the authentication process and store
credentials securely for future use.`,
	Run: runLogin,
}

var (
	loginMethod       string
	loginAPIKey       string
	loginServiceFile  string
	loginClientID     string
	loginClientSecret string
	loginForce        bool
	loginValidate     bool
)

func init() {
	// Add flags for different authentication methods
	loginCmd.Flags().StringVarP(&loginMethod, "method", "m", "",
		"Authentication method: apikey, serviceaccount, or oauth2")
	loginCmd.Flags().StringVar(&loginAPIKey, "api-key", "", "Google Cloud API key")
	loginCmd.Flags().StringVar(&loginServiceFile, "service-account", "", "Path to service account JSON file")
	loginCmd.Flags().StringVar(&loginClientID, "client-id", "", "OAuth2 client ID")
	loginCmd.Flags().StringVar(&loginClientSecret, "client-secret", "", "OAuth2 client secret")
	loginCmd.Flags().BoolVarP(&loginForce, "force", "f", false, "Force re-authentication even if already authenticated")
	loginCmd.Flags().BoolVar(&loginValidate, "validate", true, "Validate authentication by making a test API call")
}

func runLogin(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Determine authentication method
	method, err := determineAuthMethod()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error determining authentication method: %v\n", err)
		cancel() // Ensure context is canceled before exit
		os.Exit(1)
	}

	fmt.Printf("Using authentication method: %s\n", method)

	// Create auth configuration
	authConfig := createAuthConfig(method)

	// Create auth manager
	authManager := auth.NewAuthManager(authConfig)

	// Check if already authenticated (unless force is specified)
	if !loginForce && authManager.IsConfigured() {
		fmt.Println("Already authenticated. Use --force to re-authenticate.")

		if loginValidate {
			fmt.Println("Validating existing authentication...")
			if err := validateAuthentication(ctx, authManager, method); err != nil {
				fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
				fmt.Println("Please run 'assistant-cli login --force' to re-authenticate.")
				os.Exit(1)
			}
			fmt.Println("Authentication is valid!")
		}
		return
	}

	// Perform authentication
	fmt.Println("Starting authentication process...")
	if err := performAuthentication(ctx, authManager, method); err != nil {
		fmt.Fprintf(os.Stderr, "Authentication failed: %v\n", err)
		os.Exit(1)
	}

	// Validate authentication
	if loginValidate {
		fmt.Println("Validating authentication...")
		if err := validateAuthentication(ctx, authManager, method); err != nil {
			fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Authentication validated successfully!")
	}

	// Save configuration
	if err := saveAuthConfig(authConfig, method); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to save configuration: %v\n", err)
	}

	fmt.Println("Authentication completed successfully!")
	fmt.Printf("You can now use 'assistant-cli synthesize' to convert text to speech.\n")
}

// determineAuthMethod determines which authentication method to use
func determineAuthMethod() (auth.AuthMethod, error) {
	// If method is explicitly specified
	if loginMethod != "" {
		switch strings.ToLower(loginMethod) {
		case "apikey", "api-key":
			return auth.AuthMethodAPIKey, nil
		case "serviceaccount", "service-account":
			return auth.AuthMethodServiceAccount, nil
		case "oauth2", "oauth":
			return auth.AuthMethodOAuth2, nil
		default:
			return auth.AuthMethodAPIKey, fmt.Errorf("invalid authentication method: %s", loginMethod)
		}
	}

	// Auto-detect based on provided flags
	if loginAPIKey != "" {
		return auth.AuthMethodAPIKey, nil
	}
	if loginServiceFile != "" {
		return auth.AuthMethodServiceAccount, nil
	}
	if loginClientID != "" && loginClientSecret != "" {
		return auth.AuthMethodOAuth2, nil
	}

	// Check environment variables
	if os.Getenv("ASSISTANT_CLI_API_KEY") != "" {
		return auth.AuthMethodAPIKey, nil
	}
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" {
		return auth.AuthMethodServiceAccount, nil
	}
	if os.Getenv("ASSISTANT_CLI_OAUTH2_CLIENT_ID") != "" && os.Getenv("ASSISTANT_CLI_OAUTH2_CLIENT_SECRET") != "" {
		return auth.AuthMethodOAuth2, nil
	}

	// Default to prompting user
	return promptForAuthMethod()
}

// promptForAuthMethod prompts the user to select an authentication method
func promptForAuthMethod() (auth.AuthMethod, error) {
	fmt.Println("\nSelect an authentication method:")
	fmt.Println("1. API Key (simplest, requires Google Cloud API key)")
	fmt.Println("2. Service Account (for automation, requires JSON key file)")
	fmt.Println("3. OAuth2 (interactive, requires client credentials)")
	fmt.Print("\nEnter your choice (1-3): ")

	var choice string
	if _, err := fmt.Scanln(&choice); err != nil {
		return auth.AuthMethodAPIKey, fmt.Errorf("failed to read choice: %w", err)
	}

	switch choice {
	case "1":
		return auth.AuthMethodAPIKey, nil
	case "2":
		return auth.AuthMethodServiceAccount, nil
	case "3":
		return auth.AuthMethodOAuth2, nil
	default:
		return auth.AuthMethodAPIKey, fmt.Errorf("invalid choice: %s", choice)
	}
}

// createAuthConfig creates an auth configuration based on the selected method
func createAuthConfig(method auth.AuthMethod) auth.AuthConfig {
	config := auth.DefaultAuthConfig()
	config.Method = method

	switch method {
	case auth.AuthMethodAPIKey:
		if loginAPIKey != "" {
			config.APIKey = loginAPIKey
		} else if config.APIKey == "" {
			config.APIKey = promptForAPIKey()
		}

	case auth.AuthMethodServiceAccount:
		if loginServiceFile != "" {
			config.ServiceAccountFile = loginServiceFile
		} else if config.ServiceAccountFile == "" {
			config.ServiceAccountFile = promptForServiceAccountFile()
		}

	case auth.AuthMethodOAuth2:
		if loginClientID != "" {
			config.OAuth2ClientID = loginClientID
		}
		if loginClientSecret != "" {
			config.OAuth2ClientSecret = loginClientSecret
		}
		if config.OAuth2ClientID == "" || config.OAuth2ClientSecret == "" {
			promptForOAuth2Credentials(&config)
		}
	}

	return config
}

// promptForAPIKey prompts the user for an API key
func promptForAPIKey() string {
	fmt.Print("\nEnter your Google Cloud API key: ")
	var apiKey string
	_, _ = fmt.Scanln(&apiKey)
	return strings.TrimSpace(apiKey)
}

// promptForServiceAccountFile prompts the user for a service account file path
func promptForServiceAccountFile() string {
	fmt.Print("\nEnter path to service account JSON file: ")
	var filePath string
	_, _ = fmt.Scanln(&filePath)

	// Expand tilde to home directory
	if strings.HasPrefix(filePath, "~/") {
		home, _ := os.UserHomeDir()
		filePath = filepath.Join(home, filePath[2:])
	}

	return strings.TrimSpace(filePath)
}

// promptForOAuth2Credentials prompts the user for OAuth2 credentials
func promptForOAuth2Credentials(config *auth.AuthConfig) {
	if config.OAuth2ClientID == "" {
		fmt.Print("\nEnter OAuth2 Client ID: ")
		_, _ = fmt.Scanln(&config.OAuth2ClientID)
		config.OAuth2ClientID = strings.TrimSpace(config.OAuth2ClientID)
	}

	if config.OAuth2ClientSecret == "" {
		fmt.Print("Enter OAuth2 Client Secret: ")
		_, _ = fmt.Scanln(&config.OAuth2ClientSecret)
		config.OAuth2ClientSecret = strings.TrimSpace(config.OAuth2ClientSecret)
	}
}

// performAuthentication performs the actual authentication process
func performAuthentication(ctx context.Context, authManager *auth.AuthManager, method auth.AuthMethod) error {
	switch method {
	case auth.AuthMethodAPIKey:
		// API key authentication is immediate - just validate the key format
		if !authManager.IsConfigured() {
			return fmt.Errorf("API key is not properly configured")
		}
		return nil

	case auth.AuthMethodServiceAccount:
		// Service account authentication requires file validation
		if !authManager.IsConfigured() {
			return fmt.Errorf("service account file is not properly configured")
		}
		return nil

	case auth.AuthMethodOAuth2:
		// OAuth2 requires the full flow
		// The auth manager will handle the OAuth2 flow when we try to get a client
		_, err := authManager.GetClient(ctx)
		return err

	default:
		return fmt.Errorf("unsupported authentication method: %s", method)
	}
}

// validateAuthentication validates the authentication by making a test API call
func validateAuthentication(ctx context.Context, authManager *auth.AuthManager, _ auth.AuthMethod) error {

	// Get a client - this will trigger authentication if needed
	client, err := authManager.GetClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get authenticated client: %w", err)
	}
	defer client.Close()

	// Make a test API call to validate credentials
	req := &texttospeechpb.ListVoicesRequest{}
	resp, err := client.ListVoices(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to list voices: %w", err)
	}

	fmt.Printf("Successfully authenticated! Found %d available voices.\n", len(resp.Voices))
	return nil
}

// saveAuthConfig saves the authentication configuration to the config file
func saveAuthConfig(authConfig auth.AuthConfig, method auth.AuthMethod) error {
	// Set configuration values in viper
	viper.Set("auth.method", method.String())

	switch method {
	case auth.AuthMethodAPIKey:
		// Don't save API key to config file for security
		// User should use environment variable or command line flag
		fmt.Println("Note: API key not saved to config file. Use ASSISTANT_CLI_API_KEY environment variable.")

	case auth.AuthMethodServiceAccount:
		viper.Set("auth.service_account_file", authConfig.ServiceAccountFile)

	case auth.AuthMethodOAuth2:
		// Don't save client credentials to config file for security
		// OAuth2 tokens are saved separately by the OAuth2 provider
		fmt.Println("Note: OAuth2 client credentials not saved to config file. Use environment variables.")
	}

	// Get config file path
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configFile = filepath.Join(home, ".assistant-cli.yaml")
		viper.SetConfigFile(configFile)
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Write config file
	return viper.WriteConfig()
}

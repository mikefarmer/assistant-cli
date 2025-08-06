package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/mikefarmer/assistant-cli/internal/auth"
	"github.com/mikefarmer/assistant-cli/internal/output"
	"github.com/mikefarmer/assistant-cli/internal/player"
	"github.com/mikefarmer/assistant-cli/internal/tts"
	"github.com/mikefarmer/assistant-cli/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	voice        string
	languageCode string
	speakingRate float64
	pitch        float64
	volumeGain   float64
	outputFile   string
	audioFormat  string
	playAudio    bool
	listVoices   bool
)

func NewSynthesizeCmd() *cobra.Command {
	synthesizeCmd := &cobra.Command{
		Use:   "synthesize",
		Short: "Convert text to speech using Google Cloud Text-to-Speech",
		Long: `Convert text to speech using Google Cloud Text-to-Speech API.
		
Reads text from STDIN and generates an audio file with customizable voice settings.

Examples:
  echo "Hello, World!" | assistant-cli synthesize -o hello.mp3
  cat story.txt | assistant-cli synthesize --voice en-US-Wavenet-C --play
  echo "<speak>Hello <break time='1s'/> World!</speak>" | assistant-cli synthesize`,
		RunE: runSynthesize,
	}

	synthesizeCmd.Flags().StringVarP(&voice, "voice", "v", "", "Voice name (e.g., en-US-Wavenet-D)")
	synthesizeCmd.Flags().StringVarP(&languageCode, "language", "l", "en-US", "Language code (e.g., en-US, es-ES)")
	synthesizeCmd.Flags().Float64VarP(&speakingRate, "speed", "s", 1.0, "Speaking rate (0.25 to 4.0)")
	synthesizeCmd.Flags().Float64VarP(&pitch, "pitch", "p", 0.0, "Voice pitch (-20.0 to 20.0)")
	synthesizeCmd.Flags().Float64VarP(&volumeGain, "volume", "g", 0.0, "Volume gain in dB (-96.0 to 16.0)")
	synthesizeCmd.Flags().StringVarP(&outputFile, "output", "o", "output.mp3", "Output file path")
	synthesizeCmd.Flags().StringVarP(&audioFormat, "format", "f", "MP3", "Audio format (MP3, LINEAR16, OGG_OPUS, MULAW, ALAW, PCM)")
	synthesizeCmd.Flags().BoolVar(&playAudio, "play", false, "Play audio immediately after synthesis")
	synthesizeCmd.Flags().BoolVar(&listVoices, "list-voices", false, "List available voices for the language")

	viper.BindPFlag("tts.voice", synthesizeCmd.Flags().Lookup("voice"))
	viper.BindPFlag("tts.language", synthesizeCmd.Flags().Lookup("language"))
	viper.BindPFlag("tts.speaking_rate", synthesizeCmd.Flags().Lookup("speed"))
	viper.BindPFlag("tts.pitch", synthesizeCmd.Flags().Lookup("pitch"))
	viper.BindPFlag("tts.volume_gain", synthesizeCmd.Flags().Lookup("volume"))
	viper.BindPFlag("output.default_path", synthesizeCmd.Flags().Lookup("output"))
	viper.BindPFlag("output.format", synthesizeCmd.Flags().Lookup("format"))
	viper.BindPFlag("playback.auto_play", synthesizeCmd.Flags().Lookup("play"))
	
	return synthesizeCmd
}

func runSynthesize(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	authConfig := auth.AuthConfig{
		APIKey:             viper.GetString("auth.api_key"),
		ServiceAccountFile: viper.GetString("auth.service_account_file"),
		OAuth2ClientID:     viper.GetString("auth.oauth2_client_id"),
		OAuth2ClientSecret: viper.GetString("auth.oauth2_client_secret"),
		OAuth2TokenFile:    viper.GetString("auth.oauth2_token_file"),
	}

	if authConfig.APIKey == "" {
		authConfig.APIKey = os.Getenv("ASSISTANT_CLI_API_KEY")
	}
	if authConfig.ServiceAccountFile == "" {
		authConfig.ServiceAccountFile = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	}
	if authConfig.OAuth2ClientID == "" {
		authConfig.OAuth2ClientID = os.Getenv("ASSISTANT_CLI_OAUTH2_CLIENT_ID")
	}
	if authConfig.OAuth2ClientSecret == "" {
		authConfig.OAuth2ClientSecret = os.Getenv("ASSISTANT_CLI_OAUTH2_CLIENT_SECRET")
	}

	authManager := auth.NewAuthManager(authConfig)

	if err := authManager.Validate(ctx); err != nil {
		return fmt.Errorf("authentication failed: %w\nRun 'assistant-cli login' to set up authentication", err)
	}

	ttsConfig := &tts.ClientConfig{
		Voice:         viper.GetString("tts.voice"),
		LanguageCode:  viper.GetString("tts.language"),
		SpeakingRate:  viper.GetFloat64("tts.speaking_rate"),
		Pitch:         viper.GetFloat64("tts.pitch"),
		VolumeGain:    viper.GetFloat64("tts.volume_gain"),
		AudioEncoding: viper.GetString("output.format"),
	}

	if voice != "" {
		ttsConfig.Voice = voice
	}
	if languageCode != "en-US" {
		ttsConfig.LanguageCode = languageCode
	}
	if speakingRate != 1.0 {
		ttsConfig.SpeakingRate = speakingRate
	}
	if pitch != 0.0 {
		ttsConfig.Pitch = pitch
	}
	if volumeGain != 0.0 {
		ttsConfig.VolumeGain = volumeGain
	}
	if audioFormat != "MP3" {
		ttsConfig.AudioEncoding = audioFormat
	}

	ttsClient, err := tts.NewClient(ctx, authManager, ttsConfig)
	if err != nil {
		return fmt.Errorf("failed to create TTS client: %w", err)
	}
	defer ttsClient.Close()

	if listVoices {
		return handleListVoices(ctx, ttsClient, languageCode)
	}

	synthesizer := tts.NewSynthesizer(ttsClient)

	req := &tts.SynthesizeRequest{
		Voice:        ttsConfig.Voice,
		LanguageCode: ttsConfig.LanguageCode,
		SpeakingRate: ttsConfig.SpeakingRate,
		Pitch:        ttsConfig.Pitch,
		VolumeGain:   ttsConfig.VolumeGain,
		OutputFile:   outputFile,
		AudioFormat:  audioFormat,
	}

	fmt.Fprintln(os.Stderr, "Reading text from STDIN...")
	
	// Use enhanced input processing with validation
	inputProcessor := utils.NewInputProcessor(os.Stdin)
	text, err := inputProcessor.ReadText()
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	
	// Validate and potentially sanitize SSML content
	validator := utils.NewSSMLValidator()
	if err := validator.ValidateSSML(text); err != nil {
		return fmt.Errorf("input validation failed: %w", err)
	}
	
	// Display input statistics
	stats := inputProcessor.GetTextStats(text)
	fmt.Fprintf(os.Stderr, "✓ Input processed: %s\n", stats.String())
	
	// Generate safe output filename if not provided
	if outputFile == "output.mp3" {
		// Create a filename based on first few words of input
		safeFilename := output.GetSafeFilename(text[:min(50, len(text))], audioFormat)
		outputFile = safeFilename
	}
	
	req.OutputFile = outputFile
	
	resp, err := synthesizer.SynthesizeText(ctx, text, req)
	if err != nil {
		return fmt.Errorf("synthesis failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✓ Audio synthesized successfully\n")
	fmt.Fprintf(os.Stderr, "  Output: %s\n", resp.OutputFile)
	fmt.Fprintf(os.Stderr, "  Format: %s\n", resp.Format)
	fmt.Fprintf(os.Stderr, "  Size: %d bytes\n", resp.Size)

	if playAudio || viper.GetBool("playback.auto_play") {
		if err := playAudioFile(resp.OutputFile); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to play audio: %v\n", err)
		} else {
			fmt.Fprintln(os.Stderr, "✓ Audio played successfully")
		}
	}

	return nil
}

func handleListVoices(ctx context.Context, client *tts.Client, lang string) error {
	voices, err := client.ListVoices(ctx, lang)
	if err != nil {
		return fmt.Errorf("failed to list voices: %w", err)
	}

	fmt.Printf("Available voices for language '%s':\n\n", lang)
	
	for _, voice := range voices {
		var gender string
		switch voice.SsmlGender {
		case 1:
			gender = "Male"
		case 2:
			gender = "Female"
		case 3:
			gender = "Neutral"
		default:
			gender = "Unspecified"
		}
		
		fmt.Printf("  %s\n", voice.Name)
		fmt.Printf("    Gender: %s\n", gender)
		fmt.Printf("    Languages: %v\n", voice.LanguageCodes)
		fmt.Printf("    Sample Rate: %d Hz\n\n", voice.NaturalSampleRateHertz)
	}

	return nil
}

func playAudioFile(filePath string) error {
	// Check if audio playback is supported on this platform
	if !player.IsSupported() {
		return fmt.Errorf("audio playback is not supported on this platform")
	}
	
	// Create audio player
	audioPlayer, err := player.NewAudioPlayer()
	if err != nil {
		return fmt.Errorf("failed to initialize audio player: %w", err)
	}
	
	// Get player info for debugging
	info := audioPlayer.GetPlayerInfo()
	fmt.Fprintf(os.Stderr, "Playing audio with %s on %s...\n", info.Command, info.Platform)
	
	// Play the audio file
	if err := audioPlayer.Play(filePath); err != nil {
		return fmt.Errorf("failed to play audio: %w", err)
	}
	
	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
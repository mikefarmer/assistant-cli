package tts

import (
	"context"
	"fmt"
	"strings"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/mikefarmer/assistant-cli/internal/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client struct {
	client        *texttospeech.Client
	defaultVoice  *texttospeechpb.VoiceSelectionParams
	defaultAudio  *texttospeechpb.AudioConfig
	retryAttempts int
	retryDelay    time.Duration
	timeout       time.Duration
}

type ClientConfig struct {
	Voice         string
	LanguageCode  string
	SpeakingRate  float64
	Pitch         float64
	VolumeGain    float64
	AudioEncoding string
	RetryAttempts int
	RetryDelay    time.Duration
	Timeout       time.Duration
}

func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		Voice:         "en-US-Wavenet-D",
		LanguageCode:  "en-US",
		SpeakingRate:  1.0,
		Pitch:         0.0,
		VolumeGain:    0.0,
		AudioEncoding: "MP3",
		RetryAttempts: 3,
		RetryDelay:    1 * time.Second,
		Timeout:       30 * time.Second,
	}
}

func NewClient(ctx context.Context, authManager *auth.AuthManager, config *ClientConfig) (*Client, error) {
	if authManager == nil {
		return nil, fmt.Errorf("auth manager is required")
	}

	if config == nil {
		config = DefaultClientConfig()
	}

	ttsClient, err := authManager.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create TTS client: %w", err)
	}

	audioEncoding := texttospeechpb.AudioEncoding_MP3
	switch config.AudioEncoding {
	case "LINEAR16":
		audioEncoding = texttospeechpb.AudioEncoding_LINEAR16
	case "OGG_OPUS":
		audioEncoding = texttospeechpb.AudioEncoding_OGG_OPUS
	case "MP3":
		audioEncoding = texttospeechpb.AudioEncoding_MP3
	case "MULAW":
		audioEncoding = texttospeechpb.AudioEncoding_MULAW
	case "ALAW":
		audioEncoding = texttospeechpb.AudioEncoding_ALAW
	case "PCM":
		audioEncoding = texttospeechpb.AudioEncoding_PCM
	default:
		audioEncoding = texttospeechpb.AudioEncoding_MP3
	}

	client := &Client{
		client: ttsClient,
		defaultVoice: &texttospeechpb.VoiceSelectionParams{
			Name:         config.Voice,
			LanguageCode: config.LanguageCode,
		},
		defaultAudio: &texttospeechpb.AudioConfig{
			AudioEncoding:   audioEncoding,
			SpeakingRate:    config.SpeakingRate,
			Pitch:           config.Pitch,
			VolumeGainDb:    config.VolumeGain,
			EffectsProfileId: []string{"headphone-class-device"},
		},
		retryAttempts: config.RetryAttempts,
		retryDelay:    config.RetryDelay,
		timeout:       config.Timeout,
	}

	return client, nil
}

func (c *Client) Synthesize(ctx context.Context, text string, voice *texttospeechpb.VoiceSelectionParams, audio *texttospeechpb.AudioConfig) ([]byte, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	if voice == nil {
		voice = c.defaultVoice
	}

	if audio == nil {
		audio = c.defaultAudio
	}

	input := &texttospeechpb.SynthesisInput{}
	
	if isSSML(text) {
		input.InputSource = &texttospeechpb.SynthesisInput_Ssml{
			Ssml: text,
		}
	} else {
		input.InputSource = &texttospeechpb.SynthesisInput_Text{
			Text: text,
		}
	}

	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input:       input,
		Voice:       voice,
		AudioConfig: audio,
	}

	var lastErr error
	for attempt := 0; attempt <= c.retryAttempts; attempt++ {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, c.timeout)
		defer cancel()

		resp, err := c.client.SynthesizeSpeech(ctxWithTimeout, req)
		if err == nil {
			return resp.AudioContent, nil
		}

		lastErr = err

		if !isRetryableError(err) {
			return nil, fmt.Errorf("synthesis failed: %w", err)
		}

		if attempt < c.retryAttempts {
			delay := c.retryDelay * time.Duration(attempt+1)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				continue
			}
		}
	}

	return nil, fmt.Errorf("synthesis failed after %d attempts: %w", c.retryAttempts, lastErr)
}

func (c *Client) ListVoices(ctx context.Context, languageCode string) ([]*texttospeechpb.Voice, error) {
	req := &texttospeechpb.ListVoicesRequest{
		LanguageCode: languageCode,
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.ListVoices(ctxWithTimeout, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list voices: %w", err)
	}

	return resp.Voices, nil
}

func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

func isRetryableError(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	switch st.Code() {
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted:
		return true
	default:
		return false
	}
}

func isSSML(text string) bool {
	return len(text) >= 7 && strings.HasPrefix(text, "<speak>")
}
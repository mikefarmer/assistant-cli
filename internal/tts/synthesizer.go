package tts

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

// TTSClient interface for testability
type TTSClient interface {
	Synthesize(ctx context.Context, text string, voice *texttospeechpb.VoiceSelectionParams, audio *texttospeechpb.AudioConfig) ([]byte, error)
	ListVoices(ctx context.Context, languageCode string) ([]*texttospeechpb.Voice, error)
	Close() error
}

type Synthesizer struct {
	client TTSClient
}

type SynthesizeRequest struct {
	Text         string
	Voice        string
	LanguageCode string
	SpeakingRate float64
	Pitch        float64
	VolumeGain   float64
	OutputFile   string
	AudioFormat  string
}

type SynthesizeResponse struct {
	AudioData  []byte
	OutputFile string
	Format     string
	Size       int
}

func NewSynthesizer(client TTSClient) *Synthesizer {
	return &Synthesizer{
		client: client,
	}
}

func (s *Synthesizer) SynthesizeFromReader(ctx context.Context, reader io.Reader, req *SynthesizeRequest) (*SynthesizeResponse, error) {
	textData, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	text := string(textData)
	text = strings.TrimSpace(text)
	
	if text == "" {
		return nil, fmt.Errorf("input text is empty")
	}

	req.Text = text
	return s.Synthesize(ctx, req)
}

// SynthesizeText synthesizes text directly (wrapper around Synthesize)
func (s *Synthesizer) SynthesizeText(ctx context.Context, text string, req *SynthesizeRequest) (*SynthesizeResponse, error) {
	req.Text = text
	return s.Synthesize(ctx, req)
}

func (s *Synthesizer) Synthesize(ctx context.Context, req *SynthesizeRequest) (*SynthesizeResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("synthesis request cannot be nil")
	}

	if req.Text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	if err := s.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	voice := &texttospeechpb.VoiceSelectionParams{}
	if req.Voice != "" {
		voice.Name = req.Voice
	}
	if req.LanguageCode != "" {
		voice.LanguageCode = req.LanguageCode
	} else if req.Voice == "" {
		voice.LanguageCode = "en-US"
	}

	audioEncoding := s.getAudioEncoding(req.AudioFormat)
	audio := &texttospeechpb.AudioConfig{
		AudioEncoding: audioEncoding,
		SpeakingRate:  req.SpeakingRate,
		Pitch:         req.Pitch,
		VolumeGainDb:  req.VolumeGain,
		EffectsProfileId: []string{"headphone-class-device"},
	}

	audioData, err := s.client.Synthesize(ctx, req.Text, voice, audio)
	if err != nil {
		return nil, fmt.Errorf("synthesis failed: %w", err)
	}

	response := &SynthesizeResponse{
		AudioData: audioData,
		Format:    req.AudioFormat,
		Size:      len(audioData),
	}

	if req.OutputFile != "" {
		outputPath, err := s.saveToFile(audioData, req.OutputFile, req.AudioFormat)
		if err != nil {
			return nil, fmt.Errorf("failed to save audio: %w", err)
		}
		response.OutputFile = outputPath
	}

	return response, nil
}

func (s *Synthesizer) validateRequest(req *SynthesizeRequest) error {
	if req.SpeakingRate < 0.25 || req.SpeakingRate > 4.0 {
		return fmt.Errorf("speaking rate must be between 0.25 and 4.0, got %f", req.SpeakingRate)
	}

	if req.Pitch < -20.0 || req.Pitch > 20.0 {
		return fmt.Errorf("pitch must be between -20.0 and 20.0, got %f", req.Pitch)
	}

	if req.VolumeGain < -96.0 || req.VolumeGain > 16.0 {
		return fmt.Errorf("volume gain must be between -96.0 and 16.0, got %f", req.VolumeGain)
	}

	if len(req.Text) > 5000 && !isSSML(req.Text) {
		return fmt.Errorf("text length exceeds 5000 characters")
	}

	if isSSML(req.Text) {
		if err := validateSSML(req.Text); err != nil {
			return fmt.Errorf("invalid SSML: %w", err)
		}
	}

	return nil
}

func (s *Synthesizer) getAudioEncoding(format string) texttospeechpb.AudioEncoding {
	switch strings.ToUpper(format) {
	case "LINEAR16", "WAV":
		return texttospeechpb.AudioEncoding_LINEAR16
	case "OGG_OPUS", "OGG":
		return texttospeechpb.AudioEncoding_OGG_OPUS
	case "MULAW":
		return texttospeechpb.AudioEncoding_MULAW
	case "ALAW":
		return texttospeechpb.AudioEncoding_ALAW
	case "PCM":
		return texttospeechpb.AudioEncoding_PCM
	case "MP3":
		fallthrough
	default:
		return texttospeechpb.AudioEncoding_MP3
	}
}

func (s *Synthesizer) saveToFile(audioData []byte, outputFile string, format string) (string, error) {
	outputFile = filepath.Clean(outputFile)

	if outputFile == "" {
		outputFile = fmt.Sprintf("output.%s", s.getFileExtension(format))
	}

	if !strings.Contains(filepath.Base(outputFile), ".") {
		outputFile = fmt.Sprintf("%s.%s", outputFile, s.getFileExtension(format))
	}

	dir := filepath.Dir(outputFile)
	if dir != "." && dir != ".." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	if err := os.WriteFile(outputFile, audioData, 0644); err != nil {
		return "", fmt.Errorf("failed to write audio file: %w", err)
	}

	absPath, err := filepath.Abs(outputFile)
	if err != nil {
		return outputFile, nil
	}

	return absPath, nil
}

func (s *Synthesizer) getFileExtension(format string) string {
	switch strings.ToUpper(format) {
	case "LINEAR16", "WAV":
		return "wav"
	case "OGG_OPUS", "OGG":
		return "ogg"
	case "MULAW":
		return "mulaw"
	case "ALAW":
		return "alaw"
	case "PCM":
		return "pcm"
	case "MP3":
		fallthrough
	default:
		return "mp3"
	}
}

func validateSSML(text string) error {
	if !strings.HasPrefix(text, "<speak>") {
		return fmt.Errorf("SSML must start with <speak> tag")
	}

	if !strings.HasSuffix(text, "</speak>") {
		return fmt.Errorf("SSML must end with </speak> tag")
	}

	if strings.Contains(text, "<script") || strings.Contains(text, "<iframe") {
		return fmt.Errorf("potentially malicious tags detected")
	}

	openTags := strings.Count(text, "<")
	closeTags := strings.Count(text, ">")
	if openTags != closeTags {
		return fmt.Errorf("mismatched tags in SSML")
	}

	return nil
}
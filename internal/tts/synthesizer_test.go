package tts

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSynthesizer(t *testing.T) {
	client := &Client{}
	synth := NewSynthesizer(client)

	assert.NotNil(t, synth)
	assert.Equal(t, client, synth.client)
}

func TestSynthesizeFromReader(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid input",
			input:       "Hello, World!",
			expectError: false,
		},
		{
			name:        "empty input",
			input:       "",
			expectError: true,
			errorMsg:    "input text is empty",
		},
		{
			name:        "whitespace only",
			input:       "   \n\t   ",
			expectError: true,
			errorMsg:    "input text is empty",
		},
		{
			name:        "input with leading/trailing whitespace",
			input:       "  Hello, World!  \n",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockTTSClient{
				synthesizeResponse: []byte("audio_data"),
			}
			synth := &Synthesizer{client: mockClient}

			reader := strings.NewReader(tt.input)
			req := &SynthesizeRequest{
				SpeakingRate: 1.0,
				Pitch:        0.0,
				VolumeGain:   0.0,
			}

			ctx := context.Background()
			resp, err := synth.SynthesizeFromReader(ctx, reader, req)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestValidateRequest(t *testing.T) {
	synth := &Synthesizer{}

	tests := []struct {
		name        string
		req         *SynthesizeRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid request",
			req: &SynthesizeRequest{
				Text:         "Hello",
				SpeakingRate: 1.0,
				Pitch:        0.0,
				VolumeGain:   0.0,
			},
			expectError: false,
		},
		{
			name: "speaking rate too low",
			req: &SynthesizeRequest{
				Text:         "Hello",
				SpeakingRate: 0.1,
				Pitch:        0.0,
				VolumeGain:   0.0,
			},
			expectError: true,
			errorMsg:    "speaking rate must be between 0.25 and 4.0",
		},
		{
			name: "speaking rate too high",
			req: &SynthesizeRequest{
				Text:         "Hello",
				SpeakingRate: 5.0,
				Pitch:        0.0,
				VolumeGain:   0.0,
			},
			expectError: true,
			errorMsg:    "speaking rate must be between 0.25 and 4.0",
		},
		{
			name: "pitch too low",
			req: &SynthesizeRequest{
				Text:         "Hello",
				SpeakingRate: 1.0,
				Pitch:        -25.0,
				VolumeGain:   0.0,
			},
			expectError: true,
			errorMsg:    "pitch must be between -20.0 and 20.0",
		},
		{
			name: "pitch too high",
			req: &SynthesizeRequest{
				Text:         "Hello",
				SpeakingRate: 1.0,
				Pitch:        25.0,
				VolumeGain:   0.0,
			},
			expectError: true,
			errorMsg:    "pitch must be between -20.0 and 20.0",
		},
		{
			name: "volume gain too low",
			req: &SynthesizeRequest{
				Text:         "Hello",
				SpeakingRate: 1.0,
				Pitch:        0.0,
				VolumeGain:   -100.0,
			},
			expectError: true,
			errorMsg:    "volume gain must be between -96.0 and 16.0",
		},
		{
			name: "volume gain too high",
			req: &SynthesizeRequest{
				Text:         "Hello",
				SpeakingRate: 1.0,
				Pitch:        0.0,
				VolumeGain:   20.0,
			},
			expectError: true,
			errorMsg:    "volume gain must be between -96.0 and 16.0",
		},
		{
			name: "text too long",
			req: &SynthesizeRequest{
				Text:         strings.Repeat("a", 5001),
				SpeakingRate: 1.0,
				Pitch:        0.0,
				VolumeGain:   0.0,
			},
			expectError: true,
			errorMsg:    "text length exceeds 5000 characters",
		},
		{
			name: "valid SSML",
			req: &SynthesizeRequest{
				Text:         "<speak>Hello <break time='1s'/> World!</speak>",
				SpeakingRate: 1.0,
				Pitch:        0.0,
				VolumeGain:   0.0,
			},
			expectError: false,
		},
		{
			name: "invalid SSML - no closing tag",
			req: &SynthesizeRequest{
				Text:         "<speak>Hello World",
				SpeakingRate: 1.0,
				Pitch:        0.0,
				VolumeGain:   0.0,
			},
			expectError: true,
			errorMsg:    "SSML must end with </speak> tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := synth.validateRequest(tt.req)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSSML(t *testing.T) {
	tests := []struct {
		name        string
		ssml        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid SSML",
			ssml:        "<speak>Hello World</speak>",
			expectError: false,
		},
		{
			name:        "missing opening tag",
			ssml:        "Hello World</speak>",
			expectError: true,
			errorMsg:    "SSML must start with <speak> tag",
		},
		{
			name:        "missing closing tag",
			ssml:        "<speak>Hello World",
			expectError: true,
			errorMsg:    "SSML must end with </speak> tag",
		},
		{
			name:        "malicious script tag",
			ssml:        "<speak><script>alert('xss')</script></speak>",
			expectError: true,
			errorMsg:    "potentially malicious tags detected",
		},
		{
			name:        "malicious iframe tag",
			ssml:        "<speak><iframe src='evil.com'></iframe></speak>",
			expectError: true,
			errorMsg:    "potentially malicious tags detected",
		},
		{
			name:        "mismatched tags",
			ssml:        "<speak>Hello <break World</speak>",
			expectError: true,
			errorMsg:    "mismatched tags in SSML",
		},
		{
			name:        "valid with break tag",
			ssml:        "<speak>Hello <break time='1s'/> World</speak>",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSSML(tt.ssml)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetAudioEncodingFromSynthesizer(t *testing.T) {
	synth := &Synthesizer{}

	tests := []struct {
		format   string
		expected texttospeechpb.AudioEncoding
	}{
		{"MP3", texttospeechpb.AudioEncoding_MP3},
		{"mp3", texttospeechpb.AudioEncoding_MP3},
		{"LINEAR16", texttospeechpb.AudioEncoding_LINEAR16},
		{"WAV", texttospeechpb.AudioEncoding_LINEAR16},
		{"OGG_OPUS", texttospeechpb.AudioEncoding_OGG_OPUS},
		{"OGG", texttospeechpb.AudioEncoding_OGG_OPUS},
		{"MULAW", texttospeechpb.AudioEncoding_MULAW},
		{"ALAW", texttospeechpb.AudioEncoding_ALAW},
		{"PCM", texttospeechpb.AudioEncoding_PCM},
		{"unknown", texttospeechpb.AudioEncoding_MP3},
		{"", texttospeechpb.AudioEncoding_MP3},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			result := synth.getAudioEncoding(tt.format)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFileExtension(t *testing.T) {
	synth := &Synthesizer{}

	tests := []struct {
		format   string
		expected string
	}{
		{"MP3", "mp3"},
		{"mp3", "mp3"},
		{"LINEAR16", "wav"},
		{"WAV", "wav"},
		{"OGG_OPUS", "ogg"},
		{"OGG", "ogg"},
		{"MULAW", "mulaw"},
		{"ALAW", "alaw"},
		{"PCM", "pcm"},
		{"unknown", "mp3"},
		{"", "mp3"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			result := synth.getFileExtension(tt.format)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSynthesize_NilRequest(t *testing.T) {
	synth := &Synthesizer{}

	ctx := context.Background()
	resp, err := synth.Synthesize(ctx, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "synthesis request cannot be nil")
	assert.Nil(t, resp)
}

func TestSynthesize_EmptyText(t *testing.T) {
	synth := &Synthesizer{}

	req := &SynthesizeRequest{
		Text: "",
	}

	ctx := context.Background()
	resp, err := synth.Synthesize(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "text cannot be empty")
	assert.Nil(t, resp)
}

// mockTTSClient implements the TTSClient interface
type mockTTSClient struct {
	synthesizeResponse []byte
	synthesizeError    error
	listVoicesResponse []*texttospeechpb.Voice
	listVoicesError    error
}

func (m *mockTTSClient) Synthesize(ctx context.Context, text string, voice *texttospeechpb.VoiceSelectionParams,
	audio *texttospeechpb.AudioConfig) ([]byte, error) {
	return m.synthesizeResponse, m.synthesizeError
}

func (m *mockTTSClient) ListVoices(ctx context.Context, languageCode string) ([]*texttospeechpb.Voice, error) {
	return m.listVoicesResponse, m.listVoicesError
}

func (m *mockTTSClient) Close() error {
	return nil
}

func TestSynthesizeFromReader_Integration(t *testing.T) {
	mockClient := &mockTTSClient{
		synthesizeResponse: []byte("mock_audio_data"),
	}

	synth := &Synthesizer{client: mockClient}

	input := "Hello, this is a test!"
	reader := bytes.NewBufferString(input)

	req := &SynthesizeRequest{
		Voice:        "en-US-Wavenet-D",
		LanguageCode: "en-US",
		SpeakingRate: 1.0,
		Pitch:        0.0,
		VolumeGain:   0.0,
		AudioFormat:  "MP3",
	}

	ctx := context.Background()
	resp, err := synth.SynthesizeFromReader(ctx, reader, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, []byte("mock_audio_data"), resp.AudioData)
	assert.Equal(t, "MP3", resp.Format)
	assert.Equal(t, 15, resp.Size)
}

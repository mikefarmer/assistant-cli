package tts

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/mikefarmer/assistant-cli/internal/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client struct {
	client            *texttospeech.Client
	defaultVoice      *texttospeechpb.VoiceSelectionParams
	defaultAudio      *texttospeechpb.AudioConfig
	retryAttempts     int
	retryDelay       time.Duration
	timeout          time.Duration
	pool             *ConnectionPool
	metrics          *Metrics
	voiceCache       *VoiceCache
	performanceMonitor *PerformanceMonitor
}

type ConnectionPool struct {
	mu          sync.RWMutex
	connections map[string]interface{} // Generic connection pool for future use
	maxSize     int
	idleTimeout time.Duration
	lastUsed    map[string]time.Time
}

type Metrics struct {
	mu               sync.RWMutex
	requestCount     int64
	totalLatency     time.Duration
	failedRequests   int64
	cacheHits        int64
	cacheMisses      int64
	lastRequestTime  time.Time
	avgLatency       time.Duration
}

type ClientConfig struct {
	Voice           string
	LanguageCode    string
	SpeakingRate    float64
	Pitch           float64
	VolumeGain      float64
	AudioEncoding   string
	RetryAttempts   int
	RetryDelay      time.Duration
	Timeout         time.Duration
	PoolMaxSize     int
	PoolIdleTimeout time.Duration
	KeepAliveTime   time.Duration
	KeepAliveTimeout time.Duration
	EnableMetrics   bool
}

func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		Voice:            "en-US-Wavenet-D",
		LanguageCode:     "en-US",
		SpeakingRate:     1.0,
		Pitch:            0.0,
		VolumeGain:       0.0,
		AudioEncoding:    "MP3",
		RetryAttempts:    3,
		RetryDelay:       1 * time.Second,
		Timeout:          30 * time.Second,
		PoolMaxSize:      10,
		PoolIdleTimeout:  5 * time.Minute,
		KeepAliveTime:    30 * time.Second,
		KeepAliveTimeout: 5 * time.Second,
		EnableMetrics:    true,
	}
}

func NewClient(ctx context.Context, authManager *auth.AuthManager, config *ClientConfig) (*Client, error) {
	if authManager == nil {
		return nil, fmt.Errorf("auth manager is required")
	}

	if config == nil {
		config = DefaultClientConfig()
	}

	pool := &ConnectionPool{
		connections: make(map[string]interface{}),
		maxSize:     config.PoolMaxSize,
		idleTimeout: config.PoolIdleTimeout,
		lastUsed:    make(map[string]time.Time),
	}

	var metrics *Metrics
	if config.EnableMetrics {
		metrics = &Metrics{}
	}

	perfMonitor := NewPerformanceMonitor(config.EnableMetrics)

	ttsClient, err := createOptimizedClient(ctx, authManager, config)
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
		retryAttempts:      config.RetryAttempts,
		retryDelay:         config.RetryDelay,
		timeout:           config.Timeout,
		pool:             pool,
		metrics:          metrics,
		performanceMonitor: perfMonitor,
	}

	client.voiceCache = NewVoiceCache(client)
	go client.poolCleanup()

	return client, nil
}

func createOptimizedClient(ctx context.Context, authManager *auth.AuthManager, config *ClientConfig) (*texttospeech.Client, error) {
	// For now, use the existing auth manager method and enhance it with connection optimization
	// We could implement our own client creation with custom gRPC options, but that would require
	// duplicating the authentication logic. For Phase 1.6, we'll focus on other optimizations.
	return authManager.GetClient(ctx)
}

func (c *Client) poolCleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		if c.pool != nil {
			c.pool.cleanup()
		}
	}
}

func (cp *ConnectionPool) cleanup() {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	
	now := time.Now()
	for key, lastUsed := range cp.lastUsed {
		if now.Sub(lastUsed) > cp.idleTimeout {
			if _, exists := cp.connections[key]; exists {
				// For now, just remove from pool - actual connection cleanup would depend on connection type
				delete(cp.connections, key)
				delete(cp.lastUsed, key)
			}
		}
	}
}

func (c *Client) recordMetrics(start time.Time, success bool) {
	if c.metrics == nil {
		return
	}
	
	c.metrics.mu.Lock()
	defer c.metrics.mu.Unlock()
	
	duration := time.Since(start)
	c.metrics.requestCount++
	c.metrics.totalLatency += duration
	c.metrics.lastRequestTime = start
	c.metrics.avgLatency = c.metrics.totalLatency / time.Duration(c.metrics.requestCount)
	
	if !success {
		c.metrics.failedRequests++
	}
}

func (c *Client) GetMetrics() *Metrics {
	if c.metrics == nil {
		return nil
	}
	
	c.metrics.mu.RLock()
	defer c.metrics.mu.RUnlock()
	
	return &Metrics{
		requestCount:    c.metrics.requestCount,
		totalLatency:    c.metrics.totalLatency,
		failedRequests:  c.metrics.failedRequests,
		cacheHits:       c.metrics.cacheHits,
		cacheMisses:     c.metrics.cacheMisses,
		lastRequestTime: c.metrics.lastRequestTime,
		avgLatency:      c.metrics.avgLatency,
	}
}

func (c *Client) Synthesize(ctx context.Context, text string, voice *texttospeechpb.VoiceSelectionParams, audio *texttospeechpb.AudioConfig) ([]byte, error) {
	start := time.Now()
	var success bool
	var benchmarkDone func(bool, string)
	
	if c.performanceMonitor != nil {
		benchmarkDone = c.performanceMonitor.StartBenchmark("synthesize")
	} else {
		benchmarkDone = func(bool, string) {}
	}
	
	defer func() {
		c.recordMetrics(start, success)
		if success {
			benchmarkDone(true, "")
		} else {
			benchmarkDone(false, "synthesis failed")
		}
	}()

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
			success = true
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

	start := time.Now()
	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.ListVoices(ctxWithTimeout, req)
	if err != nil {
		if c.metrics != nil {
			c.recordMetrics(start, false)
		}
		return nil, fmt.Errorf("failed to list voices: %w", err)
	}

	if c.metrics != nil {
		c.recordMetrics(start, true)
	}

	return resp.Voices, nil
}

func (c *Client) ListVoicesCached(ctx context.Context, languageCode string) ([]*texttospeechpb.Voice, error) {
	if c.voiceCache != nil {
		voices, err := c.voiceCache.GetVoices(ctx, languageCode)
		if err == nil {
			if c.metrics != nil {
				c.metrics.mu.Lock()
				c.metrics.cacheHits++
				c.metrics.mu.Unlock()
			}
			return voices, nil
		}
		if c.metrics != nil {
			c.metrics.mu.Lock()
			c.metrics.cacheMisses++
			c.metrics.mu.Unlock()
		}
	}

	return c.ListVoices(ctx, languageCode)
}

func (c *Client) GetCacheStats() *CacheStats {
	if c.voiceCache != nil {
		stats := c.voiceCache.GetStats()
		return &stats
	}
	return nil
}

func (c *Client) ClearCache() {
	if c.voiceCache != nil {
		c.voiceCache.Clear()
	}
}

func (c *Client) GetPerformanceReport() string {
	if c.performanceMonitor != nil {
		return c.performanceMonitor.FormatReport()
	}
	return "Performance monitoring is disabled"
}

func (c *Client) ResetPerformanceStats() {
	if c.performanceMonitor != nil {
		c.performanceMonitor.Reset()
	}
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
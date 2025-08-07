package tts

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

type mockVoiceListClient struct {
	voices      []*texttospeechpb.Voice
	callCount   int
	mu          sync.Mutex
	shouldError bool
}

func (m *mockVoiceListClient) ListVoices(ctx context.Context, languageCode string) ([]*texttospeechpb.Voice, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++

	if m.shouldError {
		return nil, errors.New("mock error")
	}

	return m.voices, nil
}

func TestVoiceCache_GetVoices(t *testing.T) {
	voices := []*texttospeechpb.Voice{
		{Name: "en-US-Wavenet-A", LanguageCodes: []string{"en-US"}},
		{Name: "en-US-Wavenet-B", LanguageCodes: []string{"en-US"}},
	}

	mockClient := &mockVoiceListClient{voices: voices}
	cache := NewVoiceCache(mockClient)

	ctx := context.Background()

	// First call should hit the client
	result1, err := cache.GetVoices(ctx, "en-US")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result1) != 2 {
		t.Errorf("expected 2 voices, got %d", len(result1))
	}

	if mockClient.callCount != 1 {
		t.Errorf("expected 1 API call, got %d", mockClient.callCount)
	}

	// Second call should use cache
	result2, err := cache.GetVoices(ctx, "en-US")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result2) != 2 {
		t.Errorf("expected 2 voices, got %d", len(result2))
	}

	if mockClient.callCount != 1 {
		t.Errorf("expected still 1 API call (cached), got %d", mockClient.callCount)
	}

	// Check cache stats
	stats := cache.GetStats()
	if stats.hits != 1 {
		t.Errorf("expected 1 cache hit, got %d", stats.hits)
	}
	if stats.misses != 1 {
		t.Errorf("expected 1 cache miss, got %d", stats.misses)
	}
}

func TestVoiceCache_GetVoices_Error(t *testing.T) {
	mockClient := &mockVoiceListClient{shouldError: true}
	cache := NewVoiceCache(mockClient)

	ctx := context.Background()

	_, err := cache.GetVoices(ctx, "en-US")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if mockClient.callCount != 1 {
		t.Errorf("expected 1 API call, got %d", mockClient.callCount)
	}

	// Check that error doesn't get cached
	_, err2 := cache.GetVoices(ctx, "en-US")
	if err2 == nil {
		t.Fatal("expected error on second call too, got nil")
	}

	if mockClient.callCount != 2 {
		t.Errorf("expected 2 API calls (no caching on error), got %d", mockClient.callCount)
	}
}

func TestVoiceCache_TTLExpiration(t *testing.T) {
	voices := []*texttospeechpb.Voice{
		{Name: "en-US-Wavenet-A", LanguageCodes: []string{"en-US"}},
	}

	mockClient := &mockVoiceListClient{voices: voices}
	cache := NewVoiceCache(mockClient)

	ctx := context.Background()

	// Get voices first time
	_, err := cache.GetVoices(ctx, "en-US")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Manually expire the cache entry
	cache.mu.Lock()
	for key, entry := range cache.entries {
		entry.Timestamp = time.Now().Add(-20 * time.Minute) // Make it expired
		cache.entries[key] = entry
	}
	cache.mu.Unlock()

	// Next call should hit the client again
	_, err = cache.GetVoices(ctx, "en-US")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mockClient.callCount != 2 {
		t.Errorf("expected 2 API calls (expired cache), got %d", mockClient.callCount)
	}

	stats := cache.GetStats()
	if stats.misses != 2 {
		t.Errorf("expected 2 cache misses, got %d", stats.misses)
	}
}

func TestVoiceCache_GetHitRatio(t *testing.T) {
	voices := []*texttospeechpb.Voice{
		{Name: "en-US-Wavenet-A", LanguageCodes: []string{"en-US"}},
	}

	mockClient := &mockVoiceListClient{voices: voices}
	cache := NewVoiceCache(mockClient)

	ctx := context.Background()

	// First call (miss)
	cache.GetVoices(ctx, "en-US")

	// Two more calls (hits)
	cache.GetVoices(ctx, "en-US")
	cache.GetVoices(ctx, "en-US")

	hitRatio := cache.GetHitRatio()
	expectedRatio := 2.0 / 3.0 // 2 hits out of 3 total

	if hitRatio != expectedRatio {
		t.Errorf("expected hit ratio %.2f, got %.2f", expectedRatio, hitRatio)
	}
}

func TestVoiceCache_Clear(t *testing.T) {
	voices := []*texttospeechpb.Voice{
		{Name: "en-US-Wavenet-A", LanguageCodes: []string{"en-US"}},
	}

	mockClient := &mockVoiceListClient{voices: voices}
	cache := NewVoiceCache(mockClient)

	ctx := context.Background()

	// Add some data to cache
	cache.GetVoices(ctx, "en-US")
	cache.GetVoices(ctx, "en-US") // This should be a hit

	// Clear cache
	cache.Clear()

	// Next call should hit the client again
	cache.GetVoices(ctx, "en-US")

	if mockClient.callCount != 2 {
		t.Errorf("expected 2 API calls (after clear), got %d", mockClient.callCount)
	}

	// Check stats are properly maintained
	stats := cache.GetStats()
	if stats.totalSize != 1 {
		t.Errorf("expected cache size 1 after clear and reload, got %d", stats.totalSize)
	}
}

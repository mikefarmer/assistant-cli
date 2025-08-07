package tts

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

type CacheEntry struct {
	Data      []*texttospeechpb.Voice
	Timestamp time.Time
	TTL       time.Duration
}

type VoiceCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	client  VoiceListClient
	stats   CacheStats
}

type CacheStats struct {
	mu         sync.RWMutex
	hits       int64
	misses     int64
	evictions  int64
	totalSize  int64
}

type VoiceListClient interface {
	ListVoices(ctx context.Context, languageCode string) ([]*texttospeechpb.Voice, error)
}

func NewVoiceCache(client VoiceListClient) *VoiceCache {
	cache := &VoiceCache{
		entries: make(map[string]*CacheEntry),
		client:  client,
	}
	
	go cache.cleanupExpired()
	
	return cache
}

func (vc *VoiceCache) GetVoices(ctx context.Context, languageCode string) ([]*texttospeechpb.Voice, error) {
	cacheKey := fmt.Sprintf("voices:%s", languageCode)
	
	vc.mu.RLock()
	if entry, exists := vc.entries[cacheKey]; exists && !vc.isExpired(entry) {
		vc.mu.RUnlock()
		vc.recordHit()
		return entry.Data, nil
	}
	vc.mu.RUnlock()
	
	vc.recordMiss()
	
	voices, err := vc.client.ListVoices(ctx, languageCode)
	if err != nil {
		return nil, err
	}
	
	vc.mu.Lock()
	vc.entries[cacheKey] = &CacheEntry{
		Data:      voices,
		Timestamp: time.Now(),
		TTL:       15 * time.Minute, // Cache voices for 15 minutes
	}
	vc.mu.Unlock()
	
	return voices, nil
}

func (vc *VoiceCache) isExpired(entry *CacheEntry) bool {
	return time.Since(entry.Timestamp) > entry.TTL
}

func (vc *VoiceCache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		vc.mu.Lock()
		now := time.Now()
		for key, entry := range vc.entries {
			if now.Sub(entry.Timestamp) > entry.TTL {
				delete(vc.entries, key)
				vc.recordEviction()
			}
		}
		vc.mu.Unlock()
	}
}

func (vc *VoiceCache) recordHit() {
	vc.stats.mu.Lock()
	vc.stats.hits++
	vc.stats.mu.Unlock()
}

func (vc *VoiceCache) recordMiss() {
	vc.stats.mu.Lock()
	vc.stats.misses++
	vc.stats.mu.Unlock()
}

func (vc *VoiceCache) recordEviction() {
	vc.stats.mu.Lock()
	vc.stats.evictions++
	vc.stats.mu.Unlock()
}

func (vc *VoiceCache) GetStats() CacheStats {
	vc.stats.mu.RLock()
	defer vc.stats.mu.RUnlock()
	
	vc.mu.RLock()
	totalSize := int64(len(vc.entries))
	vc.mu.RUnlock()
	
	return CacheStats{
		hits:      vc.stats.hits,
		misses:    vc.stats.misses,
		evictions: vc.stats.evictions,
		totalSize: totalSize,
	}
}

func (vc *VoiceCache) Clear() {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	
	vc.entries = make(map[string]*CacheEntry)
}

func (vc *VoiceCache) GetHitRatio() float64 {
	vc.stats.mu.RLock()
	defer vc.stats.mu.RUnlock()
	
	total := vc.stats.hits + vc.stats.misses
	if total == 0 {
		return 0.0
	}
	
	return float64(vc.stats.hits) / float64(total)
}
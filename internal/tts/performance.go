package tts

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type PerformanceMonitor struct {
	mu            sync.RWMutex
	enabled       bool
	startupTime   time.Time
	benchmarks    []Benchmark
	systemMetrics SystemMetrics
}

type Benchmark struct {
	Name        string
	Duration    time.Duration
	Success     bool
	Timestamp   time.Time
	MemoryUsage int64
	ErrorMsg    string
}

type SystemMetrics struct {
	mu                sync.RWMutex
	memStats          runtime.MemStats
	lastGCTime        time.Time
	totalAllocations  uint64
	peakMemoryUsage   uint64
	goroutineCount    int
	gcPauseTotal      time.Duration
}

func NewPerformanceMonitor(enabled bool) *PerformanceMonitor {
	pm := &PerformanceMonitor{
		enabled:     enabled,
		startupTime: time.Now(),
		benchmarks:  make([]Benchmark, 0),
	}
	
	if enabled {
		go pm.collectSystemMetrics()
	}
	
	return pm
}

func (pm *PerformanceMonitor) collectSystemMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		if !pm.enabled {
			return
		}
		
		pm.systemMetrics.mu.Lock()
		runtime.ReadMemStats(&pm.systemMetrics.memStats)
		pm.systemMetrics.goroutineCount = runtime.NumGoroutine()
		
		if pm.systemMetrics.memStats.Alloc > pm.systemMetrics.peakMemoryUsage {
			pm.systemMetrics.peakMemoryUsage = pm.systemMetrics.memStats.Alloc
		}
		
		pm.systemMetrics.totalAllocations = pm.systemMetrics.memStats.TotalAlloc
		pm.systemMetrics.gcPauseTotal = time.Duration(pm.systemMetrics.memStats.PauseTotalNs)
		pm.systemMetrics.mu.Unlock()
	}
}

func (pm *PerformanceMonitor) StartBenchmark(name string) func(success bool, errorMsg string) {
	if !pm.enabled {
		return func(bool, string) {}
	}
	
	start := time.Now()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	startMem := int64(m.Alloc)
	
	return func(success bool, errorMsg string) {
		duration := time.Since(start)
		runtime.ReadMemStats(&m)
		endMem := int64(m.Alloc)
		
		benchmark := Benchmark{
			Name:        name,
			Duration:    duration,
			Success:     success,
			Timestamp:   start,
			MemoryUsage: endMem - startMem,
			ErrorMsg:    errorMsg,
		}
		
		pm.mu.Lock()
		pm.benchmarks = append(pm.benchmarks, benchmark)
		if len(pm.benchmarks) > 1000 {
			pm.benchmarks = pm.benchmarks[100:]
		}
		pm.mu.Unlock()
	}
}

func (pm *PerformanceMonitor) GetReport() PerformanceReport {
	if !pm.enabled {
		return PerformanceReport{Enabled: false}
	}
	
	pm.mu.RLock()
	benchmarksCopy := make([]Benchmark, len(pm.benchmarks))
	copy(benchmarksCopy, pm.benchmarks)
	pm.mu.RUnlock()
	
	pm.systemMetrics.mu.RLock()
	systemMetrics := pm.systemMetrics
	pm.systemMetrics.mu.RUnlock()
	
	return PerformanceReport{
		Enabled:         pm.enabled,
		Uptime:          time.Since(pm.startupTime),
		Benchmarks:      benchmarksCopy,
		SystemMetrics:   systemMetrics,
		SummaryStats:    pm.calculateSummaryStats(benchmarksCopy),
	}
}

type PerformanceReport struct {
	Enabled       bool
	Uptime        time.Duration
	Benchmarks    []Benchmark
	SystemMetrics SystemMetrics
	SummaryStats  SummaryStats
}

type SummaryStats struct {
	TotalRequests       int
	SuccessfulRequests  int
	FailedRequests      int
	AverageLatency      time.Duration
	P50Latency          time.Duration
	P90Latency          time.Duration
	P99Latency          time.Duration
	RequestsPerSecond   float64
	AverageMemoryUsage  int64
	PeakMemoryUsage     int64
	SuccessRate         float64
}

func (pm *PerformanceMonitor) calculateSummaryStats(benchmarks []Benchmark) SummaryStats {
	if len(benchmarks) == 0 {
		return SummaryStats{}
	}
	
	var totalLatency time.Duration
	var totalMemory int64
	var peakMemory int64
	var successful int
	var failed int
	durations := make([]time.Duration, len(benchmarks))
	
	for i, b := range benchmarks {
		durations[i] = b.Duration
		totalLatency += b.Duration
		totalMemory += b.MemoryUsage
		
		if b.MemoryUsage > peakMemory {
			peakMemory = b.MemoryUsage
		}
		
		if b.Success {
			successful++
		} else {
			failed++
		}
	}
	
	total := len(benchmarks)
	avgLatency := totalLatency / time.Duration(total)
	avgMemory := totalMemory / int64(total)
	successRate := float64(successful) / float64(total) * 100
	
	uptime := time.Since(pm.startupTime)
	rps := float64(total) / uptime.Seconds()
	
	return SummaryStats{
		TotalRequests:      total,
		SuccessfulRequests: successful,
		FailedRequests:     failed,
		AverageLatency:     avgLatency,
		P50Latency:         calculatePercentile(durations, 50),
		P90Latency:         calculatePercentile(durations, 90),
		P99Latency:         calculatePercentile(durations, 99),
		RequestsPerSecond:  rps,
		AverageMemoryUsage: avgMemory,
		PeakMemoryUsage:    peakMemory,
		SuccessRate:        successRate,
	}
}

func calculatePercentile(durations []time.Duration, percentile int) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	
	// Simple percentile calculation - sort and find index
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	
	// Simple insertion sort for small arrays
	for i := 1; i < len(sorted); i++ {
		key := sorted[i]
		j := i - 1
		for j >= 0 && sorted[j] > key {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}
	
	index := (percentile * (len(sorted) - 1)) / 100
	return sorted[index]
}

func (pm *PerformanceMonitor) FormatReport() string {
	report := pm.GetReport()
	if !report.Enabled {
		return "Performance monitoring is disabled"
	}
	
	return fmt.Sprintf(`
Performance Report
==================
Uptime: %v
Total Requests: %d
Success Rate: %.2f%%
Failed Requests: %d

Latency Metrics:
  Average: %v
  50th Percentile: %v
  90th Percentile: %v
  99th Percentile: %v

Throughput:
  Requests/sec: %.2f

Memory Usage:
  Average: %d bytes
  Peak: %d bytes
  Current Allocation: %d bytes
  Total Allocations: %d bytes
  GC Pause Total: %v

System:
  Goroutines: %d
  Last GC: %v ago
`,
		report.Uptime,
		report.SummaryStats.TotalRequests,
		report.SummaryStats.SuccessRate,
		report.SummaryStats.FailedRequests,
		report.SummaryStats.AverageLatency,
		report.SummaryStats.P50Latency,
		report.SummaryStats.P90Latency,
		report.SummaryStats.P99Latency,
		report.SummaryStats.RequestsPerSecond,
		report.SummaryStats.AverageMemoryUsage,
		report.SummaryStats.PeakMemoryUsage,
		report.SystemMetrics.memStats.Alloc,
		report.SystemMetrics.totalAllocations,
		report.SystemMetrics.gcPauseTotal,
		report.SystemMetrics.goroutineCount,
		time.Since(report.SystemMetrics.lastGCTime),
	)
}

func (pm *PerformanceMonitor) Reset() {
	if !pm.enabled {
		return
	}
	
	pm.mu.Lock()
	pm.benchmarks = make([]Benchmark, 0)
	pm.startupTime = time.Now()
	pm.mu.Unlock()
}
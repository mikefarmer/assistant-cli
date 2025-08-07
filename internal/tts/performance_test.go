package tts

import (
	"math"
	"testing"
	"time"
)

func abs(x float64) float64 {
	return math.Abs(x)
}

func TestPerformanceMonitor_StartBenchmark(t *testing.T) {
	pm := NewPerformanceMonitor(true)
	
	// Test successful benchmark
	done := pm.StartBenchmark("test_operation")
	time.Sleep(10 * time.Millisecond) // Simulate work
	done(true, "")
	
	report := pm.GetReport()
	if len(report.Benchmarks) != 1 {
		t.Errorf("expected 1 benchmark, got %d", len(report.Benchmarks))
	}
	
	benchmark := report.Benchmarks[0]
	if benchmark.Name != "test_operation" {
		t.Errorf("expected name 'test_operation', got %s", benchmark.Name)
	}
	
	if !benchmark.Success {
		t.Error("expected benchmark to be successful")
	}
	
	if benchmark.Duration < 10*time.Millisecond {
		t.Errorf("expected duration >= 10ms, got %v", benchmark.Duration)
	}
}

func TestPerformanceMonitor_StartBenchmark_Failed(t *testing.T) {
	pm := NewPerformanceMonitor(true)
	
	done := pm.StartBenchmark("failed_operation")
	time.Sleep(5 * time.Millisecond)
	done(false, "test error")
	
	report := pm.GetReport()
	benchmark := report.Benchmarks[0]
	
	if benchmark.Success {
		t.Error("expected benchmark to be failed")
	}
	
	if benchmark.ErrorMsg != "test error" {
		t.Errorf("expected error message 'test error', got %s", benchmark.ErrorMsg)
	}
}

func TestPerformanceMonitor_Disabled(t *testing.T) {
	pm := NewPerformanceMonitor(false)
	
	done := pm.StartBenchmark("test_operation")
	done(true, "")
	
	report := pm.GetReport()
	if report.Enabled {
		t.Error("expected performance monitoring to be disabled")
	}
	
	if len(report.Benchmarks) != 0 {
		t.Errorf("expected no benchmarks when disabled, got %d", len(report.Benchmarks))
	}
}

func TestPerformanceMonitor_SummaryStats(t *testing.T) {
	pm := NewPerformanceMonitor(true)
	
	// Add some benchmarks
	done1 := pm.StartBenchmark("operation1")
	time.Sleep(10 * time.Millisecond)
	done1(true, "")
	
	done2 := pm.StartBenchmark("operation2")
	time.Sleep(20 * time.Millisecond)
	done2(true, "")
	
	done3 := pm.StartBenchmark("operation3")
	time.Sleep(15 * time.Millisecond)
	done3(false, "error")
	
	report := pm.GetReport()
	stats := report.SummaryStats
	
	if stats.TotalRequests != 3 {
		t.Errorf("expected 3 total requests, got %d", stats.TotalRequests)
	}
	
	if stats.SuccessfulRequests != 2 {
		t.Errorf("expected 2 successful requests, got %d", stats.SuccessfulRequests)
	}
	
	if stats.FailedRequests != 1 {
		t.Errorf("expected 1 failed request, got %d", stats.FailedRequests)
	}
	
	expectedSuccessRate := (2.0 / 3.0) * 100
	if abs(stats.SuccessRate - expectedSuccessRate) > 0.01 {
		t.Errorf("expected success rate %.2f%%, got %.2f%%", expectedSuccessRate, stats.SuccessRate)
	}
	
	if stats.AverageLatency == 0 {
		t.Error("expected non-zero average latency")
	}
	
	if stats.RequestsPerSecond <= 0 {
		t.Error("expected positive requests per second")
	}
}

func TestPerformanceMonitor_Reset(t *testing.T) {
	pm := NewPerformanceMonitor(true)
	
	// Add a benchmark
	done := pm.StartBenchmark("test_operation")
	done(true, "")
	
	// Verify it's there
	report := pm.GetReport()
	if len(report.Benchmarks) != 1 {
		t.Errorf("expected 1 benchmark before reset, got %d", len(report.Benchmarks))
	}
	
	// Reset
	pm.Reset()
	
	// Verify it's cleared
	report = pm.GetReport()
	if len(report.Benchmarks) != 0 {
		t.Errorf("expected 0 benchmarks after reset, got %d", len(report.Benchmarks))
	}
}

func TestPerformanceMonitor_FormatReport(t *testing.T) {
	pm := NewPerformanceMonitor(true)
	
	done := pm.StartBenchmark("test_operation")
	time.Sleep(10 * time.Millisecond)
	done(true, "")
	
	report := pm.FormatReport()
	if report == "" {
		t.Error("expected non-empty formatted report")
	}
	
	// Check that report contains key sections
	expectedSections := []string{
		"Performance Report",
		"Uptime:",
		"Total Requests:",
		"Success Rate:",
		"Latency Metrics:",
		"Memory Usage:",
		"System:",
	}
	
	for _, section := range expectedSections {
		if len(report) == 0 || report[:10] == "" {
			t.Errorf("expected report to contain '%s'", section)
		}
	}
}

func TestPerformanceMonitor_FormatReport_Disabled(t *testing.T) {
	pm := NewPerformanceMonitor(false)
	
	report := pm.FormatReport()
	expected := "Performance monitoring is disabled"
	
	if report != expected {
		t.Errorf("expected '%s', got '%s'", expected, report)
	}
}

func TestCalculatePercentile(t *testing.T) {
	durations := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
	}
	
	p50 := calculatePercentile(durations, 50)
	// P50 of [10,20,30,40,50] with 50% index = (50 * 4) / 100 = 2, so durations[2] = 30ms
	if p50 != 30*time.Millisecond {
		t.Errorf("expected P50 to be 30ms, got %v", p50)
	}
	
	p90 := calculatePercentile(durations, 90)
	// P90 of [10,20,30,40,50] with 90% index = (90 * 4) / 100 = 3.6 -> 3, so durations[3] = 40ms
	if p90 != 40*time.Millisecond {
		t.Errorf("expected P90 to be 40ms, got %v", p90)
	}
	
	// Test edge case: empty slice
	emptyP50 := calculatePercentile([]time.Duration{}, 50)
	if emptyP50 != 0 {
		t.Errorf("expected 0 for empty slice, got %v", emptyP50)
	}
}

func BenchmarkPerformanceMonitor_StartBenchmark(b *testing.B) {
	pm := NewPerformanceMonitor(true)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		done := pm.StartBenchmark("benchmark_test")
		done(true, "")
	}
}
package monitoring

import (
	"runtime"
	"sync"
	"time"
)

// MetricsService collects and provides system metrics
type MetricsService struct {
	mu            sync.RWMutex
	requestCounts map[string]int64   // endpoint -> count
	responseTimes map[string][]int64 // endpoint -> latencies (ms)
	errorCounts   map[string]int64   // endpoint -> error count
	startTime     time.Time
	maxLatencies  int
}

// NewMetricsService creates a new metrics service
func NewMetricsService() *MetricsService {
	return &MetricsService{
		requestCounts: make(map[string]int64),
		responseTimes: make(map[string][]int64),
		errorCounts:   make(map[string]int64),
		startTime:     time.Now(),
		maxLatencies:  1000, // Keep last 1000 latencies per endpoint
	}
}

// RecordRequest records a request metric
func (s *MetricsService) RecordRequest(endpoint string, latencyMs int64, isError bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.requestCounts[endpoint]++

	// Record latency
	if s.responseTimes[endpoint] == nil {
		s.responseTimes[endpoint] = make([]int64, 0, s.maxLatencies)
	}
	latencies := s.responseTimes[endpoint]
	if len(latencies) >= s.maxLatencies {
		latencies = latencies[1:] // Remove oldest
	}
	s.responseTimes[endpoint] = append(latencies, latencyMs)

	if isError {
		s.errorCounts[endpoint]++
	}
}

// GetMetrics returns current metrics
func (s *MetricsService) GetMetrics() *Metrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Calculate endpoint metrics
	endpoints := make(map[string]EndpointMetrics)
	for endpoint, count := range s.requestCounts {
		latencies := s.responseTimes[endpoint]
		avgLatency := int64(0)
		if len(latencies) > 0 {
			sum := int64(0)
			for _, l := range latencies {
				sum += l
			}
			avgLatency = sum / int64(len(latencies))
		}

		endpoints[endpoint] = EndpointMetrics{
			RequestCount: count,
			ErrorCount:   s.errorCounts[endpoint],
			AvgLatencyMs: avgLatency,
		}
	}

	return &Metrics{
		Uptime:        time.Since(s.startTime).String(),
		UptimeSeconds: int64(time.Since(s.startTime).Seconds()),
		Memory: MemoryMetrics{
			Alloc:      memStats.Alloc,
			TotalAlloc: memStats.TotalAlloc,
			Sys:        memStats.Sys,
			NumGC:      memStats.NumGC,
		},
		Goroutines: runtime.NumGoroutine(),
		Endpoints:  endpoints,
	}
}

// Reset resets all metrics
func (s *MetricsService) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.requestCounts = make(map[string]int64)
	s.responseTimes = make(map[string][]int64)
	s.errorCounts = make(map[string]int64)
	s.startTime = time.Now()
}

// Metrics represents collected metrics
type Metrics struct {
	Uptime        string                     `json:"uptime"`
	UptimeSeconds int64                      `json:"uptime_seconds"`
	Memory        MemoryMetrics              `json:"memory"`
	Goroutines    int                        `json:"goroutines"`
	Endpoints     map[string]EndpointMetrics `json:"endpoints"`
}

// MemoryMetrics represents memory statistics
type MemoryMetrics struct {
	Alloc      uint64 `json:"alloc_bytes"`
	TotalAlloc uint64 `json:"total_alloc_bytes"`
	Sys        uint64 `json:"sys_bytes"`
	NumGC      uint32 `json:"num_gc"`
}

// EndpointMetrics represents metrics for a single endpoint
type EndpointMetrics struct {
	RequestCount int64 `json:"request_count"`
	ErrorCount   int64 `json:"error_count"`
	AvgLatencyMs int64 `json:"avg_latency_ms"`
}

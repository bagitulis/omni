package monitoring_test

import (
	"sync"
	"testing"

	"github.com/omni/backend/internal/services/monitoring"
)

func TestNewMetricsService(t *testing.T) {
	svc := monitoring.NewMetricsService()
	if svc == nil {
		t.Fatal("NewMetricsService returned nil")
	}
	// Initial state: empty metrics
	m := svc.GetMetrics()
	if m == nil {
		t.Fatal("GetMetrics returned nil on fresh service")
	}
	if len(m.Endpoints) != 0 {
		t.Errorf("expected 0 endpoints, got %d", len(m.Endpoints))
	}
	if m.UptimeSeconds < 0 {
		t.Errorf("uptime seconds should be >= 0, got %d", m.UptimeSeconds)
	}
	if m.Goroutines <= 0 {
		t.Errorf("goroutines should be > 0, got %d", m.Goroutines)
	}
}

func TestRecordRequest_Basic(t *testing.T) {
	svc := monitoring.NewMetricsService()
	svc.RecordRequest("/api/orders", 50, false)

	m := svc.GetMetrics()
	ep, ok := m.Endpoints["/api/orders"]
	if !ok {
		t.Fatal("expected endpoint '/api/orders' not found in metrics")
	}
	if ep.RequestCount != 1 {
		t.Errorf("expected RequestCount=1, got %d", ep.RequestCount)
	}
	if ep.ErrorCount != 0 {
		t.Errorf("expected ErrorCount=0, got %d", ep.ErrorCount)
	}
	if ep.AvgLatencyMs != 50 {
		t.Errorf("expected AvgLatencyMs=50, got %d", ep.AvgLatencyMs)
	}
}

func TestRecordRequest_WithError(t *testing.T) {
	svc := monitoring.NewMetricsService()
	svc.RecordRequest("/api/fail", 100, true)
	m := svc.GetMetrics()
	ep, ok := m.Endpoints["/api/fail"]
	if !ok {
		t.Fatal("expected endpoint '/api/fail' not found")
	}
	if ep.RequestCount != 1 {
		t.Errorf("expected RequestCount=1, got %d", ep.RequestCount)
	}
	if ep.ErrorCount != 1 {
		t.Errorf("expected ErrorCount=1, got %d", ep.ErrorCount)
	}
}

func TestRecordRequest_MultipleRecords_AverageLatency(t *testing.T) {
	svc := monitoring.NewMetricsService()
	// Record 3 requests: latencies 10, 20, 30 -> avg 20
	svc.RecordRequest("/api/avg", 10, false)
	svc.RecordRequest("/api/avg", 20, false)
	svc.RecordRequest("/api/avg", 30, false)
	m := svc.GetMetrics()
	ep := m.Endpoints["/api/avg"]
	if ep.RequestCount != 3 {
		t.Errorf("expected RequestCount=3, got %d", ep.RequestCount)
	}
	if ep.AvgLatencyMs != 20 {
		t.Errorf("expected AvgLatencyMs=20, got %d", ep.AvgLatencyMs)
	}
}

func TestRecordRequest_MultipleEndpoints(t *testing.T) {
	svc := monitoring.NewMetricsService()
	svc.RecordRequest("/api/a", 10, false)
	svc.RecordRequest("/api/b", 20, true)
	svc.RecordRequest("/api/a", 30, false)
	m := svc.GetMetrics()
	if len(m.Endpoints) != 2 {
		t.Errorf("expected 2 endpoints, got %d", len(m.Endpoints))
	}
	if m.Endpoints["/api/a"].RequestCount != 2 {
		t.Errorf("expected /api/a RequestCount=2, got %d", m.Endpoints["/api/a"].RequestCount)
	}
	if m.Endpoints["/api/b"].ErrorCount != 1 {
		t.Errorf("expected /api/b ErrorCount=1, got %d", m.Endpoints["/api/b"].ErrorCount)
	}
}

func TestGetMetrics_EmptyService(t *testing.T) {
	svc := monitoring.NewMetricsService()
	m := svc.GetMetrics()
	if m.Uptime == "" {
		t.Error("expected non-empty Uptime string")
	}
	if m.Memory.Sys == 0 {
		t.Error("expected non-zero Sys memory value")
	}
}

func TestReset_ClearsAllMetrics(t *testing.T) {
	svc := monitoring.NewMetricsService()
	svc.RecordRequest("/api/orders", 50, false)
	svc.RecordRequest("/api/orders", 100, true)
	svc.Reset()

	m := svc.GetMetrics()
	if len(m.Endpoints) != 0 {
		t.Errorf("expected 0 endpoints after Reset, got %d", len(m.Endpoints))
	}
}

func TestRecordRequest_Concurrent(t *testing.T) {
	svc := monitoring.NewMetricsService()
	endpoint := "/api/concurrent"
	n := 50
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(latency int64) {
			defer wg.Done()
			svc.RecordRequest(endpoint, latency, false)
		}(int64(i + 1))
	}
	wg.Wait()
	m := svc.GetMetrics()
	ep, ok := m.Endpoints[endpoint]
	if !ok {
		t.Fatal("endpoint not found after concurrent recording")
	}
	if ep.RequestCount != int64(n) {
		t.Errorf("expected RequestCount=%d, got %d", n, ep.RequestCount)
	}
}

func TestRecordRequest_MaxLatenciesRingBuffer(t *testing.T) {
	svc := monitoring.NewMetricsService()
	// maxLatencies=1000; push 1001 records - oldest (1) should be evicted
	for i := 0; i < 1001; i++ {
		svc.RecordRequest("/api/ring", int64(i+1), false)
	}

	m := svc.GetMetrics()
	ep := m.Endpoints["/api/ring"]
	// RequestCount accumulates all 1001
	if ep.RequestCount != 1001 {
		t.Errorf("expected RequestCount=1001, got %d", ep.RequestCount)
	}
	// Avg should reflect the last 1000 values (2..1001): avg = (2+1001)/2 = 503
	if ep.AvgLatencyMs != 503 {
		t.Errorf("expected AvgLatencyMs=503 (ring buffer avg), got %d", ep.AvgLatencyMs)
	}
}

func TestMetrics_UptimePositive(t *testing.T) {
	svc := monitoring.NewMetricsService()
	m := svc.GetMetrics()
	if m.UptimeSeconds < 0 {
		t.Errorf("UptimeSeconds should be >= 0, got %d", m.UptimeSeconds)
	}
}

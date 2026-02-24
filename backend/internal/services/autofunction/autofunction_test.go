package autofunction_test

import (
	"testing"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/autofunction"
)

// TestNewExecutor_Constructor verifies executor creation with nil DB.
func TestNewExecutor_Constructor(t *testing.T) {
	e := autofunction.NewExecutor(nil)
	if e == nil {
		t.Fatal("expected non-nil Executor")
	}
}

// TestExecutor_RegisterAndGetHandler verifies handler round-trip.
func TestExecutor_RegisterAndGetHandler(t *testing.T) {
	e := autofunction.NewExecutor(nil)

	called := false
	handler := autofunction.FunctionHandler(func(_ interface{}, _ string, _ *models.AutoFunctionConfig) (string, error) {
		called = true
		return "ok", nil
	})

	e.RegisterHandler("my_func", handler)
	got := e.GetHandler("my_func")
	if got == nil {
		t.Fatal("expected non-nil handler after RegisterHandler")
	}
	// Verify it is the same handler by calling it
	_, _ = got(nil, "t1", nil)
	if !called {
		t.Error("expected handler to be called")
	}
}

// TestExecutor_GetHandler_Unregistered verifies nil returned for unknown function.
func TestExecutor_GetHandler_Unregistered(t *testing.T) {
	e := autofunction.NewExecutor(nil)
	if h := e.GetHandler("nonexistent"); h != nil {
		t.Errorf("expected nil handler for unregistered function, got non-nil")
	}
}

// TestExecutor_SetTimeout verifies timeout setter does not panic.
func TestExecutor_SetTimeout(t *testing.T) {
	e := autofunction.NewExecutor(nil)
	// Should not panic; timeout is a private field but SetTimeout must not panic
	e.SetTimeout(30 * time.Second)
	e.SetTimeout(10 * time.Minute)
}

// TestExecutor_RegisterHandler_Overwrite verifies re-registration overwrites.
func TestExecutor_RegisterHandler_Overwrite(t *testing.T) {
	e := autofunction.NewExecutor(nil)

	callCount := 0
	first := autofunction.FunctionHandler(func(_ interface{}, _ string, _ *models.AutoFunctionConfig) (string, error) {
		callCount += 10
		return "", nil
	})
	second := autofunction.FunctionHandler(func(_ interface{}, _ string, _ *models.AutoFunctionConfig) (string, error) {
		callCount += 1
		return "", nil
	})

	e.RegisterHandler("fn", first)
	e.RegisterHandler("fn", second)

	h := e.GetHandler("fn")
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
	_, _ = h(nil, "t", nil)
	if callCount != 1 {
		t.Errorf("expected second handler to be active (callCount=1), got %d", callCount)
	}
}

// TestNewConfigManager_Constructor verifies config manager creation with nil DB.
func TestNewConfigManager_Constructor(t *testing.T) {
	cm := autofunction.NewConfigManager(nil, "tenant1")
	if cm == nil {
		t.Fatal("expected non-nil ConfigManager")
	}
}

// TestNewMultiTenantScheduler_IsRunning verifies scheduler starts as not running.
func TestNewMultiTenantScheduler_IsRunning(t *testing.T) {
	e := autofunction.NewExecutor(nil)
	s := autofunction.NewMultiTenantScheduler(nil, e, "/tmp/base")
	if s == nil {
		t.Fatal("expected non-nil MultiTenantScheduler")
	}
	if s.IsRunning() {
		t.Error("expected IsRunning()=false on freshly created scheduler")
	}
}

// TestNewScheduler_InitialState verifies in-memory initial state.
func TestNewScheduler_InitialState(t *testing.T) {
	e := autofunction.NewExecutor(nil)
	s := autofunction.NewScheduler(nil, e)
	if s == nil {
		t.Fatal("expected non-nil Scheduler")
	}
	if s.IsRunning() {
		t.Error("expected IsRunning()=false on fresh scheduler")
	}
	if s.GetScheduledCount() != 0 {
		t.Errorf("expected GetScheduledCount()=0, got %d", s.GetScheduledCount())
	}
}

// TestScheduler_SetTenantID verifies SetTenantID does not panic (pure setter).
func TestScheduler_SetTenantID(t *testing.T) {
	e := autofunction.NewExecutor(nil)
	s := autofunction.NewScheduler(nil, e)
	s.SetTenantID("tenant42")
	// No observable side effects; just verify it doesn't panic
}

// TestScheduler_AddOrUpdateAndGetNextRun verifies in-memory config management.
func TestScheduler_AddOrUpdateAndGetNextRun(t *testing.T) {
	e := autofunction.NewExecutor(nil)
	s := autofunction.NewScheduler(nil, e)

	next := time.Now().Add(5 * time.Minute)
	cfg := &models.AutoFunctionConfig{
		ID:                     42,
		Name:                   "test_func",
		Enabled:                true,
		IntervalMinutes:        5,
		NextScheduledExecution: &next,
	}

	if err := s.AddOrUpdate(cfg); err != nil {
		t.Fatalf("AddOrUpdate returned unexpected error: %v", err)
	}

	if s.GetScheduledCount() != 1 {
		t.Errorf("expected GetScheduledCount()=1, got %d", s.GetScheduledCount())
	}

	got := s.GetNextRun(42)
	if got == nil {
		t.Fatal("expected non-nil next run time")
	}
	if !got.Equal(next) {
		t.Errorf("expected next run=%v, got %v", next, *got)
	}
}

// TestScheduler_Remove verifies config removal from in-memory store.
func TestScheduler_Remove(t *testing.T) {
	e := autofunction.NewExecutor(nil)
	s := autofunction.NewScheduler(nil, e)

	cfg := &models.AutoFunctionConfig{ID: 10, Name: "fn", Enabled: true}
	_ = s.AddOrUpdate(cfg)
	if s.GetScheduledCount() != 1 {
		t.Fatal("precondition: expected count=1 after AddOrUpdate")
	}

	s.Remove(10)
	if s.GetScheduledCount() != 0 {
		t.Errorf("expected GetScheduledCount()=0 after Remove, got %d", s.GetScheduledCount())
	}
}

// TestScheduler_CancelScheduled verifies NextScheduledExecution is set to nil.
func TestScheduler_CancelScheduled(t *testing.T) {
	e := autofunction.NewExecutor(nil)
	s := autofunction.NewScheduler(nil, e)

	next := time.Now().Add(1 * time.Hour)
	cfg := &models.AutoFunctionConfig{ID: 99, Name: "fn", Enabled: true, NextScheduledExecution: &next}
	_ = s.AddOrUpdate(cfg)

	s.CancelScheduled(99)

	got := s.GetNextRun(99)
	if got != nil {
		t.Errorf("expected nil next run after CancelScheduled, got %v", *got)
	}
}

// TestScheduler_GetNextRun_NotFound verifies nil returned for unknown config.
func TestScheduler_GetNextRun_NotFound(t *testing.T) {
	e := autofunction.NewExecutor(nil)
	s := autofunction.NewScheduler(nil, e)
	if got := s.GetNextRun(9999); got != nil {
		t.Errorf("expected nil for unknown config ID, got %v", *got)
	}
}

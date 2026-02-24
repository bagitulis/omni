package ml_test

import (
	"os"
	"testing"
	"time"

	"github.com/omni/backend/internal/services/ml"
)

// ---------------------------------------------------------------------------
// PythonExecutor
// ---------------------------------------------------------------------------

// TestNewPythonExecutor_Defaults verifies defaults are applied for empty path and zero timeout.
func TestNewPythonExecutor_Defaults(t *testing.T) {
	e := ml.NewPythonExecutor("", 0)
	if e == nil {
		t.Fatal("expected non-nil PythonExecutor")
	}
	// Internal fields are private; we only assert non-nil and no panic.
}

// TestNewPythonExecutor_ExplicitValues verifies constructor with explicit values returns non-nil.
func TestNewPythonExecutor_ExplicitValues(t *testing.T) {
	e := ml.NewPythonExecutor("python3", 30*time.Second)
	if e == nil {
		t.Fatal("expected non-nil PythonExecutor with explicit pythonPath and timeout")
	}
}

// TestNewPythonExecutor_ZeroTimeoutGetsDefault verifies zero timeout is replaced with a default.
// We cannot inspect private fields directly, but the constructor must not panic and must return non-nil.
func TestNewPythonExecutor_ZeroTimeoutGetsDefault(t *testing.T) {
	e := ml.NewPythonExecutor("python", 0)
	if e == nil {
		t.Fatal("expected non-nil PythonExecutor when timeout=0")
	}
}

// ---------------------------------------------------------------------------
// ExecResult struct
// ---------------------------------------------------------------------------

// TestExecResult_FieldAssignment verifies all ExecResult fields can be set.
func TestExecResult_FieldAssignment(t *testing.T) {
	r := ml.ExecResult{
		Stdout:   "hello",
		Stderr:   "warn",
		ExitCode: 1,
		Duration: 250 * time.Millisecond,
	}
	if r.Stdout != "hello" {
		t.Errorf("expected Stdout='hello', got %q", r.Stdout)
	}
	if r.Stderr != "warn" {
		t.Errorf("expected Stderr='warn', got %q", r.Stderr)
	}
	if r.ExitCode != 1 {
		t.Errorf("expected ExitCode=1, got %d", r.ExitCode)
	}
	if r.Duration != 250*time.Millisecond {
		t.Errorf("expected Duration=250ms, got %v", r.Duration)
	}
}

// TestExecResult_ZeroValue verifies zero value is valid.
func TestExecResult_ZeroValue(t *testing.T) {
	var r ml.ExecResult
	if r.ExitCode != 0 {
		t.Errorf("expected ExitCode=0 for zero value, got %d", r.ExitCode)
	}
	if r.Stdout != "" {
		t.Errorf("expected empty Stdout for zero value, got %q", r.Stdout)
	}
}

// ---------------------------------------------------------------------------
// ReportService
// ---------------------------------------------------------------------------

// TestNewReportService_Constructor verifies constructor with nil DB returns non-nil.
// Uses a temp directory to avoid creating output dirs in working directory.
func TestNewReportService_Constructor(t *testing.T) {
	tmpDir := t.TempDir()
	s := ml.NewReportService(nil, tmpDir, tmpDir)
	if s == nil {
		t.Fatal("expected non-nil ReportService")
	}
}

// TestNewReportService_DefaultDirs verifies constructor with empty dirs applies defaults
// and calls os.MkdirAll on "./output" — we verify it returns non-nil and does not panic.
// We change to a temp dir so the default "./output" is created there, not in the project.
func TestNewReportService_DefaultDirs(t *testing.T) {
	tmpDir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir to temp dir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	s := ml.NewReportService(nil, "", "")
	if s == nil {
		t.Fatal("expected non-nil ReportService with empty dirs")
	}

	// Verify "./output" was created by the constructor in tmpDir
	if _, err := os.Stat("output"); os.IsNotExist(err) {
		t.Error("expected NewReportService to create './output' directory")
	}
}

package ml

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// PythonExecutor executes Python scripts
type PythonExecutor struct {
	pythonPath string
	timeout    time.Duration
}

// NewPythonExecutor creates a new executor
func NewPythonExecutor(pythonPath string, timeout time.Duration) *PythonExecutor {
	if pythonPath == "" {
		pythonPath = "python" // default
	}
	if timeout == 0 {
		timeout = 5 * time.Minute // default 5 minutes
	}
	return &PythonExecutor{
		pythonPath: pythonPath,
		timeout:    timeout,
	}
}

// ExecResult represents execution result
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

// Execute runs a Python script with arguments
func (e *PythonExecutor) Execute(ctx context.Context, scriptPath string, args ...string) (*ExecResult, error) {
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	start := time.Now()

	// Build command
	cmdArgs := append([]string{scriptPath}, args...)
	cmd := exec.CommandContext(ctx, e.pythonPath, cmdArgs...)

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute
	err := cmd.Run()

	result := &ExecResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: time.Since(start),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to execute Python script: %w", err)
		}
	}

	return result, nil
}

// ExecuteQuiet runs script without capturing output (faster)
func (e *PythonExecutor) ExecuteQuiet(ctx context.Context, scriptPath string, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	cmdArgs := append([]string{scriptPath}, args...)
	cmd := exec.CommandContext(ctx, e.pythonPath, cmdArgs...)

	return cmd.Run()
}

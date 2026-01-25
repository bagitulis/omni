// Package python provides Python code execution utilities
package python

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Runner executes Python code and returns JSON results
type Runner struct {
	workspaceRoot string
	notebooksDir  string
	moduleName    string
}

// NewRunner creates a Python runner for a specific module
func NewRunner(moduleName string) *Runner {
	workspaceRoot := findWorkspaceRoot(moduleName)
	return &Runner{
		workspaceRoot: workspaceRoot,
		notebooksDir:  filepath.Join(workspaceRoot, "notebooks"),
		moduleName:    moduleName,
	}
}

func findWorkspaceRoot(moduleName string) string {
	// Try current directory
	if fileExists(filepath.Join(".", "notebooks", moduleName, "__init__.py")) {
		abs, _ := filepath.Abs(".")
		return abs
	}

	// Try from executable location
	exe, _ := os.Executable()
	exeDir := filepath.Dir(exe)
	paths := []string{
		filepath.Join(exeDir, "..", ".."),
		filepath.Join(exeDir, "..", "..", ".."),
	}

	for _, p := range paths {
		if fileExists(filepath.Join(p, "notebooks", moduleName, "__init__.py")) {
			abs, _ := filepath.Abs(p)
			return abs
		}
	}

	// Fallback to current directory
	abs, _ := filepath.Abs(".")
	return abs
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// RunCode executes Python code and returns parsed JSON result
func (r *Runner) RunCode(pythonCode string) (interface{}, error) {
	pathSetup := `
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).parent))
`
	fullCode := pathSetup + pythonCode

	tempFile := filepath.Join(r.notebooksDir, fmt.Sprintf("_mcp_%s_temp.py", r.moduleName))

	if err := os.WriteFile(tempFile, []byte(fullCode), 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp script: %w", err)
	}
	defer os.Remove(tempFile)

	cmd := exec.Command("python", filepath.Base(tempFile))
	cmd.Dir = r.notebooksDir
	cmd.Env = append(os.Environ(), "PYTHONIOENCODING=utf-8", "PYTHONUTF8=1")

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("python error: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run python: %w", err)
	}

	result := strings.TrimSpace(string(output))
	var parsed interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w\nOutput: %s", err, result)
	}

	return parsed, nil
}

// OpenInBrowser opens a file in the default browser (Windows)
func OpenInBrowser(filePath string) error {
	cmd := exec.Command("cmd", "/c", "start", "", filePath)
	return cmd.Run()
}

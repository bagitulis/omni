package e2e

import (
	"os"
	"testing"
)

func isStrictMode() bool {
	return os.Getenv("TEST_STRICT") == "1"
}

func skipOrFail(t *testing.T, message string) {
	t.Helper()
	if isStrictMode() {
		t.Fatal(message)
		return
	}
	t.Skip(message)
}

func skipOrFailf(t *testing.T, format string, args ...interface{}) {
	t.Helper()
	if isStrictMode() {
		t.Fatalf(format, args...)
		return
	}
	t.Skipf(format, args...)
}

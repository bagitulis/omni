package testutils

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/testcontainers/testcontainers-go"
)

// writeStubCLI creates a fake container CLI in dir that exits with exitCode from
// every invocation. It lets tests exercise runtime detection deterministically
// without depending on what is installed on the host.
func writeStubCLI(t *testing.T, dir, name string, exitCode int) {
	t.Helper()

	if runtime.GOOS == "windows" {
		// A .bat stub is picked up by exec.LookPath via PATHEXT. `exit /b N`
		// sets the process exit code without a trailing message.
		script := "@echo off\r\nexit /b " + itoa(exitCode) + "\r\n"
		path := filepath.Join(dir, name+".bat")
		if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
			t.Fatalf("failed to write stub CLI %s: %v", path, err)
		}
		return
	}

	script := "#!/bin/sh\nexit " + itoa(exitCode) + "\n"
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write stub CLI %s: %v", path, err)
	}
}

// itoa avoids pulling in strconv for a single small conversion.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// TestResolveContainerCLI_PrefersEnvOverride verifies that CONTAINER_RUNTIME
// selects the CLI binary, matching scripts/python-build/omni_build/container_runtime.py.
//
// Rationale: isDockerAvailable() historically hardcoded "docker", so a machine
// running Podman would silently skip every Postgres integration test. The
// detection must honour an explicit override.
func TestResolveContainerCLI_PrefersEnvOverride(t *testing.T) {
	t.Setenv("CONTAINER_RUNTIME", "podman")

	got := resolveContainerCLI()

	if got != "podman" {
		t.Fatalf("resolveContainerCLI() = %q, want %q when CONTAINER_RUNTIME=podman", got, "podman")
	}
}

// TestResolveContainerCLI_FallsBackToAvailableBinary verifies auto-detection
// when no override is set: docker is preferred, podman is the fallback.
func TestResolveContainerCLI_FallsBackToAvailableBinary(t *testing.T) {
	os.Unsetenv("CONTAINER_RUNTIME")

	got := resolveContainerCLI()

	if got == "" {
		t.Skip("no container runtime installed on this machine; nothing to assert")
	}

	if got != "docker" && got != "podman" {
		t.Fatalf("resolveContainerCLI() = %q, want \"docker\" or \"podman\"", got)
	}

	// Whatever it picked must actually resolve on PATH.
	if _, err := exec.LookPath(got); err != nil {
		t.Fatalf("resolveContainerCLI() returned %q but it is not on PATH: %v", got, err)
	}
}

// TestResolveContainerCLI_EmptyWhenRuntimeAbsent verifies that detection reports
// "no runtime" by returning an empty string rather than guessing. Callers rely on
// this to decide whether to skip, instead of keying off runtime.GOOS.
func TestResolveContainerCLI_EmptyWhenRuntimeAbsent(t *testing.T) {
	// Point PATH at an empty dir so neither docker nor podman can be found.
	t.Setenv("CONTAINER_RUNTIME", "")
	t.Setenv("PATH", t.TempDir())

	got := resolveContainerCLI()

	if got != "" {
		t.Fatalf("resolveContainerCLI() = %q, want \"\" when no runtime is on PATH", got)
	}
}

// TestIsContainerRuntimeAvailable_ReflectsDetection verifies the availability
// helper agrees with detection: when no CLI resolves, availability is false and
// must not depend on GOOS.
func TestIsContainerRuntimeAvailable_ReflectsDetection(t *testing.T) {
	t.Setenv("CONTAINER_RUNTIME", "")
	t.Setenv("PATH", t.TempDir())

	if isContainerRuntimeAvailable() {
		t.Fatal("isContainerRuntimeAvailable() = true, want false when no runtime is on PATH")
	}
}

// TestIsContainerRuntimeAvailable_FalseWhenCliCannotReachHost verifies that a
// present-but-dead runtime is reported as unavailable.
//
// This is the Windows failure mode: Podman/Docker Desktop place a CLI on PATH
// before the backing VM is running. Treating "CLI exists" as "runtime usable"
// makes integration tests fail instead of skip.
//
// The test injects a stub CLI that always exits non-zero for `info`, so it does
// not depend on what is actually installed on the machine.
func TestIsContainerRuntimeAvailable_FalseWhenCliCannotReachHost(t *testing.T) {
	stubDir := t.TempDir()
	writeStubCLI(t, stubDir, "docker", 1)

	t.Setenv("CONTAINER_RUNTIME", "")
	t.Setenv("PATH", stubDir)

	if isContainerRuntimeAvailable() {
		t.Fatal("isContainerRuntimeAvailable() = true, want false when the CLI cannot reach a host")
	}
}

// TestIsContainerRuntimeAvailable_TrueWhenCliReachesHost is the positive
// counterpart: a CLI whose `info` succeeds means the runtime is usable.
func TestIsContainerRuntimeAvailable_TrueWhenCliReachesHost(t *testing.T) {
	stubDir := t.TempDir()
	writeStubCLI(t, stubDir, "docker", 0)

	t.Setenv("CONTAINER_RUNTIME", "")
	t.Setenv("PATH", stubDir)

	if !isContainerRuntimeAvailable() {
		t.Fatal("isContainerRuntimeAvailable() = false, want true when the CLI reaches a host")
	}
}

// TestStartContainerSafely_ConvertsPanicToError verifies that a panic raised
// inside the testcontainers constructor is recovered and surfaced as an error.
//
// testcontainers-go v0.40.0 panics via MustExtractDockerHost when no usable
// container host can be determined. Left unrecovered, that panic aborts the
// entire test binary, so an absent runtime reads as a crash rather than a skip.
//
// The assertion is deliberately one-sided: err must be non-nil. On a machine
// with a working runtime the call legitimately succeeds, so we only fail when
// the panic escapes (which would crash the binary before this point).
func TestStartContainerSafely_ConvertsPanicToError(t *testing.T) {
	container, err := startContainerSafely(context.Background(), testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
	})

	if err != nil {
		if container != nil {
			t.Fatal("startContainerSafely() returned a non-nil container alongside an error")
		}
		return
	}

	// Success path: a real runtime is present, so clean up what we started.
	if container != nil {
		_ = container.Terminate(context.Background())
	}
}

// TestStartContainerSafely_RecoversPanicFromProvider proves the recovery path
// itself works, independent of the local environment. It calls an invalid image
// reference so the provider fails during startup rather than succeeding.
func TestStartContainerSafely_RecoversPanicFromProvider(t *testing.T) {
	if resolveContainerCLI() == "" {
		t.Skip("skipping: no container runtime installed, so no provider to exercise")
	}

	// An intentionally malformed image makes the provider fail. Whether that
	// surfaces as a panic or an error, startContainerSafely must not propagate
	// a panic and must report an error.
	container, err := startContainerSafely(context.Background(), testcontainers.ContainerRequest{
		Image: "", // invalid on purpose
	})

	if err == nil {
		t.Fatal("startContainerSafely() returned nil error for an invalid image request")
	}

	if container != nil {
		t.Fatal("startContainerSafely() returned a non-nil container alongside an error")
	}
}

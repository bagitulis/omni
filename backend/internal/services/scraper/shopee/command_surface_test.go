package shopee

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// The backend and the extension agree on a set of command names, and nothing
// enforces that agreement at compile time: an action is a string on both sides.
//
// A name that exists here but not there fails only when a scrape reaches that
// command against a live Shopee page — the most expensive place to find out.
// This reads both sources and compares them.

// actionsSentByBackend extracts the action names this package sends.
func actionsSentByBackend(t *testing.T) []string {
	t.Helper()

	pattern := regexp.MustCompile(`Send\(ctx, "([a-z_]+)"`)
	found := map[string]bool{}

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for _, m := range pattern.FindAllStringSubmatch(string(src), -1) {
			found[m[1]] = true
		}
	}
	return sortedKeys(found)
}

// actionsHandledByExtension extracts the names the extension can answer: the
// content script's dispatch table plus the worker's own switch.
func actionsHandledByExtension(t *testing.T, extensionDir string) []string {
	t.Helper()

	found := map[string]bool{}

	contentSrc := mustRead(t, filepath.Join(extensionDir, "content-shopee.js"))
	// The COMMANDS map's entries, e.g. "  check_blocked: checkBlocked,".
	commandEntry := regexp.MustCompile(`(?m)^\s{4}([a-z_]+):\s`)
	for _, m := range commandEntry.FindAllStringSubmatch(contentSrc, -1) {
		found[m[1]] = true
	}

	workerSrc := mustRead(t, filepath.Join(extensionDir, "service-worker.js"))
	workerCase := regexp.MustCompile(`case '([a-z_]+)':`)
	for _, m := range workerCase.FindAllStringSubmatch(workerSrc, -1) {
		found[m[1]] = true
	}

	return sortedKeys(found)
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// TestCommandSurface_EveryBackendActionIsImplemented is the gate itself.
func TestCommandSurface_EveryBackendActionIsImplemented(t *testing.T) {
	extensionDir := filepath.Join("..", "..", "..", "..", "..", "extension")
	if _, err := os.Stat(extensionDir); err != nil {
		t.Skipf("extension sources not present at %s: %v", extensionDir, err)
	}

	sent := actionsSentByBackend(t)
	if len(sent) == 0 {
		t.Fatal("no actions found in the backend sources; the extractor is broken, not the code")
	}

	handled := map[string]bool{}
	for _, a := range actionsHandledByExtension(t, extensionDir) {
		handled[a] = true
	}
	if len(handled) == 0 {
		t.Fatal("no actions found in the extension sources; the extractor is broken, not the code")
	}

	var missing []string
	for _, action := range sent {
		if !handled[action] {
			missing = append(missing, action)
		}
	}
	if len(missing) > 0 {
		t.Errorf("the backend sends actions the extension cannot answer: %v\n"+
			"sent: %v\nhandled: %v", missing, sent, sortedKeys(handled))
	}
}

// TestCommandSurface_BlockerCheckIsWired pins the specific command this change
// added, so a later edit cannot quietly drop it from either side and restore the
// original defect.
func TestCommandSurface_BlockerCheckIsWired(t *testing.T) {
	extensionDir := filepath.Join("..", "..", "..", "..", "..", "extension")
	if _, err := os.Stat(extensionDir); err != nil {
		t.Skipf("extension sources not present at %s: %v", extensionDir, err)
	}

	var backendSends bool
	for _, a := range actionsSentByBackend(t) {
		if a == "check_blocked" {
			backendSends = true
		}
	}
	if !backendSends {
		t.Error("the backend no longer sends check_blocked; a captcha would end the run as a success again")
	}

	var extensionHandles bool
	for _, a := range actionsHandledByExtension(t, extensionDir) {
		if a == "check_blocked" {
			extensionHandles = true
		}
	}
	if !extensionHandles {
		t.Error("the extension no longer implements check_blocked")
	}
}

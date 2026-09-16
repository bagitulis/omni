package shopee

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The command-surface gate is only worth having if it can fail.
//
// A test that reads two files and compares them passes trivially when the
// extractor returns nothing from either side — which is exactly what happens if
// a refactor renames the dispatch table or changes how commands are sent. These
// tests feed the extractor a deliberately broken copy and require it to notice.

// writeFixtureExtension creates a throwaway extension directory whose content
// script is missing one command.
func writeFixtureExtension(t *testing.T, omitAction string) string {
	t.Helper()
	dir := t.TempDir()

	content := `(function () {
  const COMMANDS = {
    extract: extract,
    smart_scroll: smartScroll,
    click_next: clickNext,
    check_last_page: checkLastPage,
    check_blocked: checkBlocked,
    install_observer: installObserver,
    observe_network: observeNetwork,
  };
})();`
	if omitAction != "" {
		content = strings.ReplaceAll(content, "    "+omitAction+": ", "    omitted_"+omitAction+": ")
	}
	if err := os.WriteFile(filepath.Join(dir, "content-shopee.js"), []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture content script: %v", err)
	}

	worker := `switch (action) {
      case 'open_tab':
      case 'close_tab':
      case 'goto':
      case 'reload':
      case 'wait':
    }`
	if err := os.WriteFile(filepath.Join(dir, "service-worker.js"), []byte(worker), 0o600); err != nil {
		t.Fatalf("write fixture worker: %v", err)
	}
	return dir
}

// TestCommandSurfaceGate_DetectsAMissingAction proves the comparison bites.
func TestCommandSurfaceGate_DetectsAMissingAction(t *testing.T) {
	dir := writeFixtureExtension(t, "check_blocked")

	handled := map[string]bool{}
	for _, a := range actionsHandledByExtension(t, dir) {
		handled[a] = true
	}

	if handled["check_blocked"] {
		t.Fatal("the fixture was supposed to omit check_blocked; the extractor is matching something else")
	}

	var missing []string
	for _, action := range actionsSentByBackend(t) {
		if !handled[action] {
			missing = append(missing, action)
		}
	}
	if len(missing) == 0 {
		t.Error("the gate found nothing missing against an extension that omits a command it is sent")
	}

	var sawBlocker bool
	for _, m := range missing {
		if m == "check_blocked" {
			sawBlocker = true
		}
	}
	if !sawBlocker {
		t.Errorf("expected check_blocked among the missing actions, got %v", missing)
	}
}

// TestCommandSurfaceGate_PassesOnACompleteFixture guards the other direction: a
// gate that always reports a problem is ignored within a week.
func TestCommandSurfaceGate_PassesOnACompleteFixture(t *testing.T) {
	dir := writeFixtureExtension(t, "")

	handled := map[string]bool{}
	for _, a := range actionsHandledByExtension(t, dir) {
		handled[a] = true
	}

	for _, action := range actionsSentByBackend(t) {
		if !handled[action] {
			t.Errorf("complete fixture reported %q as missing", action)
		}
	}
}

// TestCommandSurfaceGate_ExtractorFindsBackendActions pins the other half. If
// this returned nothing, every comparison above would pass vacuously.
func TestCommandSurfaceGate_ExtractorFindsBackendActions(t *testing.T) {
	sent := actionsSentByBackend(t)

	for _, required := range []string{"extract", "check_blocked", "check_last_page", "observe_network"} {
		var found bool
		for _, a := range sent {
			if a == required {
				found = true
			}
		}
		if !found {
			t.Errorf("the extractor did not find %q among the backend's actions: %v", required, sent)
		}
	}
}

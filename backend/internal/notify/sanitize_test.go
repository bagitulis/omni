package notify

import "testing"

func TestSanitizeForUser_Safe(t *testing.T) {
	safe := []string{
		"Bulk Stock Sync completed",
		"Order 12345 arrived",
		"Shopee sync finished successfully",
	}
	for _, s := range safe {
		if got := SanitizeForUser(s); got != s {
			t.Fatalf("safe string mutated: in=%q out=%q", s, got)
		}
	}
}

func TestSanitizeForUser_Blocked(t *testing.T) {
	blocked := []string{
		"panic: runtime error: invalid memory address",
		"SELECT * FROM users WHERE id = 1",
		"goroutine 12 [running]",
		"Bearer eyJhbGciOi.some.jwt",
		"ECONNREFUSED 127.0.0.1:5432",
		"stack trace at foo.bar (main.go:12)",
		"password=hunter2",
		"apikey=sk-secret",
		"database query failed: relation \"orders\" does not exist",
	}
	for _, s := range blocked {
		if SanitizeForUser(s) == s {
			t.Fatalf("unsafe string NOT sanitized: %q", s)
		}
	}
}

func TestSanitizeForUser_Truncates(t *testing.T) {
	long := ""
	for range 600 {
		long += "a"
	}
	got := SanitizeForUser(long)
	if len(got) > 500 {
		t.Fatalf("length not capped: %d", len(got))
	}
}

func TestSanitizeForUser_EmptyReturnsFallback(t *testing.T) {
	if SanitizeForUser("") == "" {
		t.Fatal("empty must return safe fallback (non-empty)")
	}
}

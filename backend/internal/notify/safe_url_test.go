package notify

import "testing"

func TestSafeActionURL(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
		ok   bool
	}{
		{"empty allowed as no-op", "", "", true},
		{"root path", "/", "/", true},
		{"path with segments", "/orders/123", "/orders/123", true},
		{"path with query", "/notifications?expand=1", "/notifications?expand=1", true},
		{"path with hash", "/products#top", "/products#top", true},
		{"path with query and hash", "/x?q=1&r=2#z", "/x?q=1&r=2#z", true},
		{"unicode-ish encoded", "/o/%E2%9C%93", "/o/%E2%9C%93", true},
		{"reject absolute https", "https://evil.example", "", false},
		{"reject protocol-relative", "//evil.example/path", "", false},
		{"reject javascript scheme", "javascript:alert(1)", "", false},
		{"reject data scheme", "data:text/html,x", "", false},
		{"reject mailto", "mailto:foo@bar", "", false},
		{"reject backslash", `\evil`, "", false},
		{"reject not-starting-with-slash", "orders/1", "", false},
		{"reject whitespace", "   ", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := SafeActionURL(c.in)
			if c.ok && err != nil {
				t.Fatalf("expected ok, got err %v", err)
			}
			if !c.ok && err == nil {
				t.Fatalf("expected err, got value %q", got)
			}
			if got != c.want {
				t.Fatalf("value: got %q want %q", got, c.want)
			}
		})
	}
}

func TestSafeActionURL_TooLong(t *testing.T) {
	long := "/" + repeat("a", 600)
	if _, err := SafeActionURL(long); err == nil {
		t.Fatal("expected reject for >500 chars")
	}
}

func repeat(s string, n int) string {
	out := make([]byte, 0, len(s)*n)
	for range n {
		out = append(out, s...)
	}
	return string(out)
}

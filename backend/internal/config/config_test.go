package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Test default values
	cfg := Load()

	if cfg.Environment != "development" {
		t.Errorf("expected Environment 'development', got '%s'", cfg.Environment)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected Port '8080', got '%s'", cfg.Port)
	}
	if cfg.PGPort != 5432 {
		t.Errorf("expected PGPort 5432, got %d", cfg.PGPort)
	}
}

func TestLoadWithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("GO_ENV", "production")
	os.Setenv("PORT", "3000")
	defer func() {
		os.Unsetenv("GO_ENV")
		os.Unsetenv("PORT")
	}()

	cfg := Load()

	if cfg.Environment != "production" {
		t.Errorf("expected Environment 'production', got '%s'", cfg.Environment)
	}
	if cfg.Port != "3000" {
		t.Errorf("expected Port '3000', got '%s'", cfg.Port)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *Config
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid config",
			cfg: &Config{
				JWTSecret:     "this-is-a-very-long-secret-key-for-testing-purposes",
				EncryptionKey: "some-encryption-key",
				PGHost:        "localhost",
				PGPassword:    "test-pass",
			},
			wantError: false,
		},
		{
			name: "missing JWT secret",
			cfg: &Config{
				JWTSecret:     "",
				EncryptionKey: "some-key",
			},
			wantError: true,
			errorMsg:  "JWT_SECRET is required",
		},
		{
			name: "missing encryption key",
			cfg: &Config{
				JWTSecret:     "this-is-a-very-long-secret-key-for-testing-purposes",
				EncryptionKey: "",
				PGPassword:    "test-pass",
			},
			wantError: true,
			errorMsg:  "ENCRYPTION_KEY is required",
		},
		{
			name: "JWT secret too short",
			cfg: &Config{
				JWTSecret:     "short",
				EncryptionKey: "some-key",
				PGPassword:    "test-pass",
			},
			wantError: true,
			errorMsg:  "JWT_SECRET must be at least 32 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("expected error '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestConfig_IsProduction(t *testing.T) {
	tests := []struct {
		env      string
		expected bool
	}{
		{"production", true},
		{"development", false},
		{"staging", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			cfg := &Config{Environment: tt.env}
			if got := cfg.IsProduction(); got != tt.expected {
				t.Errorf("IsProduction() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	// Test with existing env var
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	if got := getEnv("TEST_VAR", "default"); got != "test_value" {
		t.Errorf("getEnv() = %v, want 'test_value'", got)
	}

	// Test with non-existing env var
	if got := getEnv("NON_EXISTING_VAR", "default"); got != "default" {
		t.Errorf("getEnv() = %v, want 'default'", got)
	}
}

func TestGetEnvInt(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	if got := getEnvInt("TEST_INT", 0); got != 42 {
		t.Errorf("getEnvInt() = %v, want 42", got)
	}

	// Test with invalid int
	os.Setenv("TEST_INVALID_INT", "not-a-number")
	defer os.Unsetenv("TEST_INVALID_INT")

	if got := getEnvInt("TEST_INVALID_INT", 10); got != 10 {
		t.Errorf("getEnvInt() = %v, want 10 (default)", got)
	}

	// Test with non-existing
	if got := getEnvInt("NON_EXISTING_INT", 99); got != 99 {
		t.Errorf("getEnvInt() = %v, want 99", got)
	}
}

func TestGetEnvList(t *testing.T) {
	os.Setenv("TEST_LIST", "a,b,c")
	defer os.Unsetenv("TEST_LIST")

	got := getEnvList("TEST_LIST", []string{"default"})
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("getEnvList() = %v, want [a b c]", got)
	}

	// Test with spaces
	os.Setenv("TEST_LIST_SPACES", "a , b , c")
	defer os.Unsetenv("TEST_LIST_SPACES")

	got = getEnvList("TEST_LIST_SPACES", []string{})
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("getEnvList() with spaces = %v, want [a b c]", got)
	}

	// Test with non-existing
	got = getEnvList("NON_EXISTING_LIST", []string{"default"})
	if len(got) != 1 || got[0] != "default" {
		t.Errorf("getEnvList() = %v, want [default]", got)
	}
}

func TestSplitString(t *testing.T) {
	tests := []struct {
		input    string
		sep      string
		expected []string
	}{
		{"a,b,c", ",", []string{"a", "b", "c"}},
		{"a::b::c", "::", []string{"a", "b", "c"}},
		{"single", ",", []string{"single"}},
		{"", ",", []string{""}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := splitString(tt.input, tt.sep)
			if len(got) != len(tt.expected) {
				t.Errorf("splitString() len = %v, want %v", len(got), len(tt.expected))
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("splitString()[%d] = %v, want %v", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestTrimSpace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"\thello\t", "hello"},
		{"hello", "hello"},
		{"  ", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := trimSpace(tt.input); got != tt.expected {
				t.Errorf("trimSpace() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetDataDir(t *testing.T) {
	// Test default
	os.Unsetenv("DATABASE_PATH")
	if got := GetDataDir(); got != "./data" {
		t.Errorf("GetDataDir() = %v, want './data'", got)
	}

	// Test with env var
	os.Setenv("DATABASE_PATH", "/custom/path")
	defer os.Unsetenv("DATABASE_PATH")

	if got := GetDataDir(); got != "/custom/path" {
		t.Errorf("GetDataDir() = %v, want '/custom/path'", got)
	}
}

func TestHasStandaloneUniqueTag(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		expected bool
	}{
		{
			name:     "unique index only",
			tag:      "column:state;uniqueIndex;not null",
			expected: false,
		},
		{
			name:     "explicit unique tag",
			tag:      "column:username;unique;not null",
			expected: true,
		},
		{
			name:     "named unique constraint",
			tag:      "column:email;unique:idx_users_email",
			expected: true,
		},
		{
			name:     "empty tag",
			tag:      "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasStandaloneUniqueTag(tt.tag); got != tt.expected {
				t.Errorf("hasStandaloneUniqueTag(%q) = %v, want %v", tt.tag, got, tt.expected)
			}
		})
	}
}

package env

import (
	"os"
	"testing"
	"time"
)

func TestGetString(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		setValue string
		setEnv   bool
		def      string
		expected string
	}{
		{
			name:     "returns env value when set",
			key:      "TEST_STRING_VAR",
			setValue: "hello",
			setEnv:   true,
			def:      "default",
			expected: "hello",
		},
		{
			name:     "returns default when not set",
			key:      "TEST_STRING_VAR_UNSET",
			setEnv:   false,
			def:      "default",
			expected: "default",
		},
		{
			name:     "returns empty string when set to empty",
			key:      "TEST_STRING_VAR_EMPTY",
			setValue: "",
			setEnv:   true,
			def:      "default",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before and after
			os.Unsetenv(tt.key)
			defer os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.setValue)
			}

			result := GetString(tt.key, tt.def)
			if result != tt.expected {
				t.Errorf("GetString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetBool(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		setValue string
		setEnv   bool
		def      bool
		expected bool
	}{
		{
			name:     "returns true for 'true'",
			key:      "TEST_BOOL_VAR_TRUE",
			setValue: "true",
			setEnv:   true,
			def:      false,
			expected: true,
		},
		{
			name:     "returns false for 'false'",
			key:      "TEST_BOOL_VAR_FALSE",
			setValue: "false",
			setEnv:   true,
			def:      true,
			expected: false,
		},
		{
			name:     "returns true for '1'",
			key:      "TEST_BOOL_VAR_ONE",
			setValue: "1",
			setEnv:   true,
			def:      false,
			expected: true,
		},
		{
			name:     "returns false for '0'",
			key:      "TEST_BOOL_VAR_ZERO",
			setValue: "0",
			setEnv:   true,
			def:      true,
			expected: false,
		},
		{
			name:     "returns true for empty string when var is set",
			key:      "TEST_BOOL_VAR_EMPTY",
			setValue: "",
			setEnv:   true,
			def:      false,
			expected: true,
		},
		{
			name:     "returns default when not set",
			key:      "TEST_BOOL_VAR_UNSET",
			setEnv:   false,
			def:      true,
			expected: true,
		},
		{
			name:     "returns default for invalid value",
			key:      "TEST_BOOL_VAR_INVALID",
			setValue: "invalid",
			setEnv:   true,
			def:      false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.key)
			defer os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.setValue)
			}

			result := GetBool(tt.key, tt.def)
			if result != tt.expected {
				t.Errorf("GetBool() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		setValue string
		setEnv   bool
		def      int
		expected int
	}{
		{
			name:     "returns parsed int",
			key:      "TEST_INT_VAR",
			setValue: "42",
			setEnv:   true,
			def:      0,
			expected: 42,
		},
		{
			name:     "returns negative int",
			key:      "TEST_INT_VAR_NEG",
			setValue: "-10",
			setEnv:   true,
			def:      0,
			expected: -10,
		},
		{
			name:     "returns zero",
			key:      "TEST_INT_VAR_ZERO",
			setValue: "0",
			setEnv:   true,
			def:      99,
			expected: 0,
		},
		{
			name:     "returns default when not set",
			key:      "TEST_INT_VAR_UNSET",
			setEnv:   false,
			def:      123,
			expected: 123,
		},
		{
			name:     "returns default for invalid value",
			key:      "TEST_INT_VAR_INVALID",
			setValue: "not-a-number",
			setEnv:   true,
			def:      456,
			expected: 456,
		},
		{
			name:     "returns default for empty string",
			key:      "TEST_INT_VAR_EMPTY",
			setValue: "",
			setEnv:   true,
			def:      789,
			expected: 789,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.key)
			defer os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.setValue)
			}

			result := GetInt(tt.key, tt.def)
			if result != tt.expected {
				t.Errorf("GetInt() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestGetDuration(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		setValue string
		setEnv   bool
		def      time.Duration
		expected time.Duration
	}{
		{
			name:     "returns seconds duration",
			key:      "TEST_DUR_VAR_SEC",
			setValue: "5s",
			setEnv:   true,
			def:      0,
			expected: 5 * time.Second,
		},
		{
			name:     "returns minutes duration",
			key:      "TEST_DUR_VAR_MIN",
			setValue: "2m",
			setEnv:   true,
			def:      0,
			expected: 2 * time.Minute,
		},
		{
			name:     "returns hours duration",
			key:      "TEST_DUR_VAR_HOUR",
			setValue: "1h",
			setEnv:   true,
			def:      0,
			expected: time.Hour,
		},
		{
			name:     "returns milliseconds duration",
			key:      "TEST_DUR_VAR_MS",
			setValue: "500ms",
			setEnv:   true,
			def:      0,
			expected: 500 * time.Millisecond,
		},
		{
			name:     "returns combined duration",
			key:      "TEST_DUR_VAR_COMBO",
			setValue: "1h30m",
			setEnv:   true,
			def:      0,
			expected: time.Hour + 30*time.Minute,
		},
		{
			name:     "returns default when not set",
			key:      "TEST_DUR_VAR_UNSET",
			setEnv:   false,
			def:      10 * time.Second,
			expected: 10 * time.Second,
		},
		{
			name:     "returns default for invalid value",
			key:      "TEST_DUR_VAR_INVALID",
			setValue: "invalid",
			setEnv:   true,
			def:      15 * time.Second,
			expected: 15 * time.Second,
		},
		{
			name:     "returns default for empty string",
			key:      "TEST_DUR_VAR_EMPTY",
			setValue: "",
			setEnv:   true,
			def:      20 * time.Second,
			expected: 20 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.key)
			defer os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.setValue)
			}

			result := GetDuration(tt.key, tt.def)
			if result != tt.expected {
				t.Errorf("GetDuration() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAllFunctionsWithRealEnvVars(t *testing.T) {
	// Test all functions together with environment variables
	const (
		strKey  = "TEST_REAL_STR"
		boolKey = "TEST_REAL_BOOL"
		intKey  = "TEST_REAL_INT"
		durKey  = "TEST_REAL_DUR"
	)

	// Clean up
	defer func() {
		os.Unsetenv(strKey)
		os.Unsetenv(boolKey)
		os.Unsetenv(intKey)
		os.Unsetenv(durKey)
	}()

	// Set values
	os.Setenv(strKey, "test-value")
	os.Setenv(boolKey, "true")
	os.Setenv(intKey, "999")
	os.Setenv(durKey, "30s")

	// Verify all work correctly
	if s := GetString(strKey, ""); s != "test-value" {
		t.Errorf("GetString() = %q, want %q", s, "test-value")
	}
	if b := GetBool(boolKey, false); b != true {
		t.Errorf("GetBool() = %v, want %v", b, true)
	}
	if i := GetInt(intKey, 0); i != 999 {
		t.Errorf("GetInt() = %d, want %d", i, 999)
	}
	if d := GetDuration(durKey, 0); d != 30*time.Second {
		t.Errorf("GetDuration() = %v, want %v", d, 30*time.Second)
	}
}

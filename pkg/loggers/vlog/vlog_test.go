package vlog

import (
	"bytes"
	"strings"
	"testing"
)

func TestUnescapeMultilineAttrs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple log without escapes",
			input:    `time=2025-10-23T19:29:28+02:00 level=DEBUG msg="Response Body"` + "\n",
			expected: `time=2025-10-23T19:29:28+02:00 level=DEBUG msg="Response Body"` + "\n",
		},
		{
			name:  "log with escaped JSON in attribute",
			input: `time=2025-10-23T19:29:28+02:00 level=DEBUG msg="Response Body" body="{\n  \"name\": \"test\",\n  \"value\": 123\n}"` + "\n",
			expected: `time=2025-10-23T19:29:28+02:00 level=DEBUG msg="Response Body" body={
  "name": "test",
  "value": 123
}` + "\n",
		},
		{
			name:  "log with multiple escaped JSON attributes",
			input: `time=2025-10-23T19:29:28+02:00 level=DEBUG msg="Test" data="{\n  \"a\": 1\n}" extra="{\n  \"b\": 2\n}"` + "\n",
			expected: `time=2025-10-23T19:29:28+02:00 level=DEBUG msg="Test" data={
  "a": 1
} extra={
  "b": 2
}` + "\n",
		},
		{
			name:  "log with escaped quotes inside JSON",
			input: `time=2025-10-23T19:29:28+02:00 level=DEBUG body="{\n  \"key\": \"value with \\\"quotes\\\"\"\n}"` + "\n",
			expected: `time=2025-10-23T19:29:28+02:00 level=DEBUG body={
  "key": "value with \"quotes\""
}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unescapeMultilineAttrs([]byte(tt.input))
			if string(result) != tt.expected {
				t.Errorf("unescapeMultilineAttrs() =\n%s\nwant:\n%s", string(result), tt.expected)
			}
		})
	}
}

func TestUnescapeMultilineAttrsRealWorld(t *testing.T) {
	// Simulated real-world log from controller-runtime with large JSON body
	input := `time=2025-10-23T19:29:28+02:00 level=DEBUG msg="Response Body" controller=networkconfiguration body="{\n  \"apiVersion\": \"vitistack.io/v1alpha1\",\n  \"kind\": \"NetworkConfiguration\",\n  \"metadata\": {\n    \"name\": \"test-networkconfiguration\",\n    \"namespace\": \"default\"\n  }\n}"` + "\n"

	result := unescapeMultilineAttrs([]byte(input))
	resultStr := string(result)

	// Verify that escaped newlines were converted to real newlines
	if !strings.Contains(resultStr, "\"apiVersion\": \"vitistack.io/v1alpha1\",\n  \"kind\":") {
		t.Error("Expected real newlines in output")
	}

	// Verify that the quotes around the JSON were removed
	if strings.Contains(resultStr, `body="{\n`) {
		t.Error("Expected escaped newlines to be unescaped")
	}

	// Verify structure
	if !strings.Contains(resultStr, "body={") {
		t.Error("Expected body={ format")
	}
}

func BenchmarkUnescapeMultilineAttrs(b *testing.B) {
	input := []byte(`time=2025-10-23T19:29:28+02:00 level=DEBUG msg="Response Body" body="{\n  \"apiVersion\": \"v1\",\n  \"kind\": \"Pod\",\n  \"metadata\": {\n    \"name\": \"test\"\n  }\n}"` + "\n")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = unescapeMultilineAttrs(input)
	}
}

func BenchmarkUnescapeMultilineAttrsNoEscapes(b *testing.B) {
	input := []byte(`time=2025-10-23T19:29:28+02:00 level=INFO msg="Simple message" key=value` + "\n")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = unescapeMultilineAttrs(input)
	}
}

func TestRebuildMultilineValue(t *testing.T) {
	orig := []byte(`time=2025-10-23T19:29:28+02:00 body="old" extra=value`)
	expanded := []byte("new\nmultiline\nvalue")
	eqIdx := bytes.Index(orig, []byte(`body=`))
	end := bytes.Index(orig, []byte(`" extra`))

	result := rebuildMultilineValue(orig, eqIdx+4, end, expanded)
	expected := `time=2025-10-23T19:29:28+02:00 body=new
multiline
value extra=value`

	if string(result) != expected {
		t.Errorf("rebuildMultilineValue() =\n%s\nwant:\n%s", string(result), expected)
	}
}

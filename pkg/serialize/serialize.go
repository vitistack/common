package serialize

import (
	"encoding/json"
	"fmt"
	"strings"
)

// JSON returns a compact JSON string representation of v.
// On error, it returns a best-effort fallback using fmt with the error appended.
func JSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fallback(v, err)
	}
	return string(b)
}

// JSONIndent pretty-prints v as JSON with the provided indent string (e.g., "  " or "\t").
// On error, it returns a best-effort fallback using fmt with the error appended.
func JSONIndent(v any, indent string) string {
	b, err := json.MarshalIndent(v, "", indent)
	if err != nil {
		return fallback(v, err)
	}
	return string(b)
}

// JSONIndentN pretty-prints v as JSON with n spaces of indentation (n <= 0 defaults to 2).
func JSONIndentN(v any, n int) string {
	if n <= 0 {
		n = 2
	}
	return JSONIndent(v, strings.Repeat(" ", n))
}

// Pretty is a convenience alias for JSONIndentN(v, 2).
func Pretty(v any) string { return JSONIndentN(v, 2) }

// BytesJSON returns the compact JSON bytes and any error encountered.
func BytesJSON(v any) ([]byte, error) { return json.Marshal(v) }

// BytesJSONIndent returns the indented JSON bytes and any error encountered.
func BytesJSONIndent(v any, indent string) ([]byte, error) { return json.MarshalIndent(v, "", indent) }

// As returns JSON string representation of v; when indent > 0, it pretty-prints with that many spaces.
func As(v any, indent int) string {
	if indent > 0 {
		return JSONIndentN(v, indent)
	}
	return JSON(v)
}

func fallback(v any, err error) string {
	return fmt.Sprintf("%v (serialize error: %v)", v, err)
}

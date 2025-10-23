package vlog

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"
)

// prepare a sample encoded slog text line similar to what the TextHandler would emit.
// We'll focus on the post-processing steps: unescapeMultilineAttrs + applyMultilineColor.
func sampleLine(single bool) []byte {
	if single {
		return []byte("time=2025-09-26T12:00:00Z level=INFO msg=\"hello world\"\n")
	}
	// multiline JSON pretty printed inside msg string (escaped newlines)
	return []byte("time=2025-09-26T12:00:00Z level=INFO msg=\"{\\n  \\\"a\\\": 1,\\n  \\\"b\\\": 2\\n}\"\n")
}

func BenchmarkUnescapeSingleLine(b *testing.B) {
	line := sampleLine(true)
	for i := 0; i < b.N; i++ {
		_ = unescapeMultilineAttrs(line)
	}
}

func BenchmarkUnescapeMultiline(b *testing.B) {
	line := sampleLine(false)
	for i := 0; i < b.N; i++ {
		_ = unescapeMultilineAttrs(line)
	}
}

func BenchmarkFullColorHandlerSingle(b *testing.B) {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	h := newColorTextHandler(bytes.NewBuffer(nil), opts)
	rec := slog.NewRecord(testTime(), slog.LevelInfo, "hello world", 0)
	b.ResetTimer()
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_ = h.Handle(ctx, rec)
	}
}

func BenchmarkFullColorHandlerMultiline(b *testing.B) {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	h := newColorTextHandler(bytes.NewBuffer(nil), opts)
	msg := prettyValue{v: map[string]int{"a": 1, "b": 2}}.String()
	rec := slog.NewRecord(testTime(), slog.LevelInfo, msg, 0)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h.Handle(ctx, rec)
	}
}

// testTime returns a constant time to reduce allocations.
func testTime() time.Time { return time.Unix(0, 0).UTC() }

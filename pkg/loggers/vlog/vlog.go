package vlog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"log/slog"

	"github.com/go-logr/logr"
	"github.com/vitistack/common/pkg/loggers"
)

// Package-level logger with lazy default initialization.
var (
	base       *slog.Logger
	once       sync.Once
	addCaller  bool
	doUnescape bool
)

// Options configures the vlog logger (now backed by Go's slog).
// ColorizeLine and DisableStacktrace are currently no-ops for slog.
type Options struct {
	// Level sets the minimum log level. One of: "debug", "info", "warn", "error".
	// Values like "dpanic", "panic", "fatal" are treated as "error" for slog.
	Level string
	// JSON switches the encoder to JSON instead of human-readable text.
	JSON bool
	// AddCaller includes caller information (file:line) when true.
	AddCaller bool
	// DisableStacktrace is kept for compatibility; slog does not emit stacktraces by default.
	DisableStacktrace bool
	// ColorizeLine is kept for compatibility; not applied with slog's standard handlers.
	ColorizeLine bool
	// UnescapeMultiline when true will post-process console (non-JSON) lines to turn escaped \n inside
	// msg="..." into real multi-line output (removing surrounding quotes). Adds a small per-log overhead.
	// Default: false (favor performance); can be enabled when human readability of large multi-line messages matters.
	UnescapeMultiline bool
}

// Setup initializes the global slog-based logger with the provided options.
func Setup(opts Options) error {
	addCaller = opts.AddCaller
	doUnescape = opts.UnescapeMultiline
	handlerOpts := &slog.HandlerOptions{
		AddSource: false, // we add caller manually to control the skip depth
		Level:     slogLevelFromString(opts.Level),
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Format time as RFC3339 to match previous output style
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format(time.RFC3339))
				}
			}
			return a
		},
	}

	var h slog.Handler
	switch {
	case opts.JSON:
		h = slog.NewJSONHandler(os.Stdout, handlerOpts)
	case opts.ColorizeLine:
		h = newColorTextHandler(os.Stdout, handlerOpts)
	default:
		h = slog.NewTextHandler(os.Stdout, handlerOpts)
	}

	base = slog.New(h)
	return nil
}

// ensure ensures the logger is initialized with sensible defaults.
func ensure() {
	if base != nil {
		return
	}
	once.Do(func() {
		if base != nil {
			return
		}
		_ = Setup(Options{ // defaults: JSON (efficient), info level
			Level:             "info",
			JSON:              true,
			AddCaller:         false,
			ColorizeLine:      false,
			UnescapeMultiline: false,
		})
	})
}

// Sync is kept for API compatibility; slog's standard handlers don't buffer.
func Sync() error { return nil }

// Logr returns a logr.Logger backed by the slog logger, for controller-runtime integration.
func Logr() logr.Logger {
	ensure()
	sink := &slogSink{logger: base}
	return logr.New(sink)
}

// Logger returns the generic loggers.Logger backed by this package's slog logger.
func Logger() loggers.Logger {
	ensure()
	return &SugaredLogger{logger: base}
}

// Debug logs at Debug level. Accepts mixed arguments.
func Debug(args ...any) { logArgs(slog.LevelDebug, args...) }

// Info logs at Info level. Accepts mixed arguments.
func Info(args ...any) { logArgs(slog.LevelInfo, args...) }

// Warn logs at Warn level. Accepts mixed arguments.
func Warn(args ...any) { logArgs(slog.LevelWarn, args...) }

// Error logs at Error level. Accepts mixed arguments.
func Error(args ...any) { logArgs(slog.LevelError, args...) }

// DPanic logs at Error level (closest mapping in slog).
func DPanic(args ...any) { logArgs(slog.LevelError, args...) }

// Panic logs at Error level then panics.
func Panic(args ...any) { logArgs(slog.LevelError, args...); panic(fmt.Sprint(args...)) }

// Fatal logs at Error level then exits(1).
func Fatal(args ...any) { logArgs(slog.LevelError, args...); os.Exit(1) }

// Formatted variants.
func Debugf(format string, args ...any)  { logMsg(slog.LevelDebug, fmt.Sprintf(format, args...)) }
func Infof(format string, args ...any)   { logMsg(slog.LevelInfo, fmt.Sprintf(format, args...)) }
func Warnf(format string, args ...any)   { logMsg(slog.LevelWarn, fmt.Sprintf(format, args...)) }
func Errorf(format string, args ...any)  { logMsg(slog.LevelError, fmt.Sprintf(format, args...)) }
func DPanicf(format string, args ...any) { logMsg(slog.LevelError, fmt.Sprintf(format, args...)) }
func Panicf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	logMsg(slog.LevelError, msg)
	panic(msg)
}
func Fatalf(format string, args ...any) {
	logMsg(slog.LevelError, fmt.Sprintf(format, args...))
	os.Exit(1)
}

// With returns a child logger with additional structured context provided as key-value pairs.
// Example: vlog.With("pod", podName, "ns", namespace).Info("created")
func With(keysAndValues ...any) *SugaredLogger {
	ensure()
	return &SugaredLogger{logger: base.With(convertKVs(keysAndValues)...)}
}

// SugaredLogger provides chainable methods similar to zap's SugaredLogger.
type SugaredLogger struct{ logger *slog.Logger }

func (s *SugaredLogger) Debug(args ...any) {
	s.logger.Log(context.Background(), slog.LevelDebug, fmt.Sprint(args...))
}
func (s *SugaredLogger) Info(args ...any) {
	s.logger.Log(context.Background(), slog.LevelInfo, fmt.Sprint(args...))
}
func (s *SugaredLogger) Warn(args ...any) {
	s.logger.Log(context.Background(), slog.LevelWarn, fmt.Sprint(args...))
}
func (s *SugaredLogger) Error(args ...any) {
	s.logger.Log(context.Background(), slog.LevelError, fmt.Sprint(args...))
}
func (s *SugaredLogger) Debugf(f string, a ...any) {
	s.logger.Log(context.Background(), slog.LevelDebug, fmt.Sprintf(f, a...))
}
func (s *SugaredLogger) Infof(f string, a ...any) {
	s.logger.Log(context.Background(), slog.LevelInfo, fmt.Sprintf(f, a...))
}
func (s *SugaredLogger) Warnf(f string, a ...any) {
	s.logger.Log(context.Background(), slog.LevelWarn, fmt.Sprintf(f, a...))
}
func (s *SugaredLogger) Errorf(f string, a ...any) {
	s.logger.Log(context.Background(), slog.LevelError, fmt.Sprintf(f, a...))
}
func (s *SugaredLogger) With(kv ...any) loggers.Logger {
	return &SugaredLogger{logger: s.logger.With(convertKVs(kv)...)}
}
func (s *SugaredLogger) WithGroup(name string) *SugaredLogger {
	return &SugaredLogger{logger: s.logger.WithGroup(name)}
}

// Ensure SugaredLogger implements the generic loggers.Logger interface.
var _ loggers.Logger = (*SugaredLogger)(nil)

// --- helpers ---

func logArgs(level slog.Level, args ...any) {
	ensure()
	if len(args) == 0 {
		return
	}
	writeRecord(base, level, fmt.Sprint(args...))
}

func logMsg(level slog.Level, msg string) {
	ensure()
	writeRecord(base, level, msg)
}

// writeRecord constructs a slog.Record with a caller pointing at the first frame outside this package.
func writeRecord(logger *slog.Logger, level slog.Level, msg string) {
	h := logger.Handler()
	pc := uintptr(0)
	file := ""
	line := 0
	if addCaller {
		pc, file, line = findExternalCaller()
	}
	rec := slog.NewRecord(time.Now(), level, msg, pc)
	if addCaller && file != "" {
		short := shortenPath(file)
		rec.AddAttrs(slog.String("caller", fmt.Sprintf("%s:%d", short, line)))
	}
	_ = h.Handle(context.Background(), rec)
}

// findExternalCaller returns the (pc,file,line) for the first stack frame not in this vlog package.
func findExternalCaller() (uintptr, string, int) {
	// Skip: runtime.Callers, findExternalCaller, writeRecord/logArgs/logMsg wrappers.
	const skip = 4
	pcs := make([]uintptr, 32)
	n := runtime.Callers(skip, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		fr, more := frames.Next()
		if fr.Function == "" || !strings.Contains(fr.Function, "/pkg/loggers/vlog.") {
			return fr.PC, fr.File, fr.Line
		}
		if !more {
			break
		}
	}
	return 0, "", 0
}

// shortenPath returns last two path components for brevity.
func shortenPath(p string) string {
	if p == "" {
		return p
	}
	p = filepath.ToSlash(p)
	parts := strings.Split(p, "/")
	if len(parts) <= 2 {
		return p
	}
	return strings.Join(parts[len(parts)-2:], "/")
}

func slogLevelFromString(lvl string) slog.Leveler {
	switch strings.ToLower(lvl) {
	case "debug":
		return slog.LevelDebug
	case "", "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error", "dpanic", "panic", "fatal":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func convertKVs(kv []any) []any {
	// Ensure even-length list; if odd, append a value placeholder
	if len(kv)%2 == 1 {
		kv = append(kv, "<missing>")
	}
	return kv
}

// Pretty wraps any value and, when logged, attempts to pretty-print JSON or YAML (structs, maps, slices, or raw JSON/YAML strings).
// Usage: vlog.Info("object", vlog.Pretty(obj))
// Works for both text and JSON logging modes (in JSON mode the pretty text is still embedded as the message or value string).
func Pretty(v any) any { return prettyValue{v: v} }

type prettyValue struct{ v any }

func (p prettyValue) String() string {
	if p.v == nil {
		return "null"
	}
	// If it's already a string that looks like JSON or YAML, try to reformat.
	if s, ok := p.v.(string); ok {
		trimmed := strings.TrimSpace(s)
		if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") { // JSON guess
			if pretty, ok := reformatJSONBytes([]byte(trimmed)); ok {
				return pretty
			}
		}
		if looksLikeYAML(trimmed) {
			if pretty, ok := reformatYAMLBytes([]byte(trimmed)); ok {
				return pretty
			}
		}
		// fallback original
		return s
	}

	// Try JSON marshal with indent first.
	if b, err := json.MarshalIndent(p.v, "", "  "); err == nil {
		return string(b)
	}
	// Try YAML marshal.
	if b, err := yaml.Marshal(p.v); err == nil {
		return string(b)
	}
	// Fallback verbose formatting.
	return fmt.Sprintf("%+v", p.v)
}

func reformatJSONBytes(b []byte) (string, bool) {
	var anyVal any
	if err := json.Unmarshal(b, &anyVal); err != nil {
		return "", false
	}
	out, err := json.MarshalIndent(anyVal, "", "  ")
	if err != nil {
		return "", false
	}
	return string(out), true
}

func reformatYAMLBytes(b []byte) (string, bool) {
	var anyVal any
	if err := yaml.Unmarshal(b, &anyVal); err != nil {
		return "", false
	}
	// Marshal back to YAML (yaml lib already emits multi-line with indentation)
	out, err := yaml.Marshal(anyVal)
	if err != nil {
		return "", false
	}
	return string(out), true
}

func looksLikeYAML(s string) bool {
	// Heuristic: contains ':' early (key: value) and not pure JSON braces.
	if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
		return false
	}
	colon := strings.IndexByte(s, ':')
	if colon > 0 && colon < 80 { // arbitrary small window
		return true
	}
	return false
}

// --- colorized text handler for slog ---

// ANSI codes
const (
	ansiReset  = "\x1b[0m"
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
	ansiBlue   = "\x1b[34m"
)

func levelColorSlog(lvl slog.Level) string {
	switch {
	case lvl <= slog.LevelDebug:
		return ansiBlue
	case lvl < slog.LevelWarn:
		return ansiGreen // info
	case lvl < slog.LevelError:
		return ansiYellow // warn
	default:
		return ansiRed // error and above
	}
}

type colorTextHandler struct {
	w      *syncWriter
	opts   *slog.HandlerOptions
	attrs  []slog.Attr
	groups []string
}

func newColorTextHandler(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	return &colorTextHandler{w: &syncWriter{w: w}, opts: opts}
}

func (h *colorTextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// Delegate to an inner TextHandler for level filtering
	var th slog.Handler = slog.NewTextHandler(io.Discard, h.opts)
	return th.Enabled(ctx, level)
}

//nolint:gocritic // slog.Handler requires a value parameter for Record
func (h *colorTextHandler) Handle(ctx context.Context, r slog.Record) error {
	// Render the record using a TextHandler into a buffer
	var buf bytes.Buffer
	var th slog.Handler = slog.NewTextHandler(&buf, h.opts)
	// Apply accumulated groups and attrs
	for _, g := range h.groups {
		th = th.WithGroup(g)
	}
	if len(h.attrs) > 0 {
		th = th.WithAttrs(h.attrs)
	}
	if !th.Enabled(ctx, r.Level) {
		return nil
	}
	if err := th.Handle(ctx, r); err != nil {
		return err
	}
	b := buf.Bytes()
	// Optional multiline unescape (enabled only when UnescapeMultiline option is set and using text mode).
	if doUnescape {
		b = unescapeMultilineMsg(b)
	}
	color := levelColorSlog(r.Level)
	// If multi-line, ensure each line starts with color and ends with reset to keep coloring consistent.
	b = applyMultilineColor(b, color)
	if _, err := h.w.Write(b); err != nil {
		return err
	}
	return nil
}

func (h *colorTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nh := *h
	nh.attrs = append(append([]slog.Attr(nil), h.attrs...), attrs...)
	return &nh
}

func (h *colorTextHandler) WithGroup(name string) slog.Handler {
	nh := *h
	nh.groups = append(append([]string(nil), h.groups...), name)
	return &nh
}

// syncWriter serializes writes to avoid color interleaving across goroutines.
type syncWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (s *syncWriter) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.w.Write(p)
}

func (s *syncWriter) WriteString(str string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.w.Write([]byte(str))
}

// applyMultilineColor wraps each line (including last even if empty) with color/reset.
func applyMultilineColor(b []byte, color string) []byte {
	if len(b) == 0 {
		return b
	}
	// Remove trailing newline, remember to re-add.
	hadNL := false
	if b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
		hadNL = true
	}
	lines := bytes.Split(b, []byte{'\n'})
	out := make([]byte, 0, len(b)+len(lines)*len(color)+len(lines)*len(ansiReset)+2)
	for i, ln := range lines {
		out = append(out, color...)
		out = append(out, ln...)
		out = append(out, ansiReset...)
		if i < len(lines)-1 {
			out = append(out, '\n')
		}
	}
	if hadNL {
		out = append(out, '\n')
	}
	return out
}

// unescapeMultilineMsg finds the msg="..." segment, and if it contains escaped newlines (\n),
// it replaces the quoted, escaped content with an unquoted, real multiline block.
func unescapeMultilineMsg(b []byte) []byte {
	key := []byte(" msg=\"")
	idx := bytes.Index(b, key)
	if idx == -1 {
		return b
	}
	start := idx + len(key)
	end := findClosingQuote(b, start)
	if end == -1 {
		return b
	}
	segment := b[start:end]
	if !bytes.Contains(segment, []byte(`\n`)) {
		return b
	}
	expanded := unescapeMsgSegment(segment)
	return rebuildMultilineLine(b, idx, end, expanded)
}

func findClosingQuote(b []byte, start int) int {
	esc := false
	for i := start; i < len(b); i++ {
		c := b[i]
		if esc {
			esc = false
			continue
		}
		if c == '\\' {
			esc = true
			continue
		}
		if c == '"' {
			return i
		}
	}
	return -1
}

func unescapeMsgSegment(seg []byte) []byte {
	var out bytes.Buffer
	out.Grow(len(seg) + 8)
	esc := false
	for i := 0; i < len(seg); i++ {
		c := seg[i]
		if esc {
			switch c {
			case 'n':
				out.WriteByte('\n')
			case 't':
				out.WriteByte('\t')
			case '\\', '"':
				out.WriteByte(c)
			default:
				out.WriteByte(c)
			}
			esc = false
			continue
		}
		if c == '\\' {
			esc = true
			continue
		}
		out.WriteByte(c)
	}
	return out.Bytes()
}

func rebuildMultilineLine(orig []byte, keyIdx, end int, expanded []byte) []byte {
	var rebuilt bytes.Buffer
	rebuilt.Grow(len(orig) + len(expanded))
	_, _ = rebuilt.Write(orig[:keyIdx])
	_, _ = rebuilt.WriteString(" msg=")
	_, _ = rebuilt.Write(expanded)
	_, _ = rebuilt.Write(orig[end+1:])
	return rebuilt.Bytes()
}

// slogSink adapts slog.Logger to logr.LogSink for controller-runtime compatibility.
type slogSink struct {
	logger *slog.Logger
	name   string
	kv     []any
}

func (s *slogSink) Init(_ logr.RuntimeInfo) {}

func (s *slogSink) Enabled(_ int) bool {
	// logr will call Enabled with a verbosity level; we delegate to slog's handler for Info level.
	// We conservatively report true; handler will still filter by level.
	return true
}

func (s *slogSink) Info(level int, msg string, keysAndValues ...any) {
	l := s.logger
	if s.name != "" {
		l = l.WithGroup(s.name)
	}
	if len(s.kv) > 0 {
		l = l.With(convertKVs(s.kv)...)
	}
	// Map logr verbosity: V(0) -> Info, V(1+) -> Debug.
	lvl := slog.LevelInfo
	if level > 0 {
		lvl = slog.LevelDebug
	}
	l.Log(context.Background(), lvl, msg, convertKVs(keysAndValues)...)
}

func (s *slogSink) Error(err error, msg string, keysAndValues ...any) {
	l := s.logger
	if s.name != "" {
		l = l.WithGroup(s.name)
	}
	if len(s.kv) > 0 {
		l = l.With(convertKVs(s.kv)...)
	}
	attrs := append(convertKVs(keysAndValues), "err", err)
	l.Log(context.Background(), slog.LevelError, msg, attrs...)
}

func (s *slogSink) WithValues(keysAndValues ...any) logr.LogSink {
	return &slogSink{logger: s.logger, name: s.name, kv: append(append([]any(nil), s.kv...), keysAndValues...)}
}

func (s *slogSink) WithName(name string) logr.LogSink {
	// Chain groups using '/'
	newName := name
	if s.name != "" {
		newName = s.name + "/" + name
	}
	return &slogSink{logger: s.logger, name: newName, kv: append([]any(nil), s.kv...)}
}

package vlog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"log/slog"

	"github.com/go-logr/logr"
	"github.com/vitistack/common/pkg/loggers"
)

// Package-level logger with lazy default initialization.
var (
	base *slog.Logger
	once sync.Once
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
}

// Setup initializes the global slog-based logger with the provided options.
func Setup(opts Options) error {
	handlerOpts := &slog.HandlerOptions{
		AddSource: opts.AddCaller,
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
		_ = Setup(Options{ // defaults: console text, info level
			Level:        "info",
			JSON:         false,
			AddCaller:    false,
			ColorizeLine: false,
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
	base.Log(context.Background(), level, fmt.Sprint(args...))
}

func logMsg(level slog.Level, msg string) {
	ensure()
	base.Log(context.Background(), level, msg)
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
	color := levelColorSlog(r.Level)
	// Wrap entire line in color, ensuring reset before trailing newline.
	if n := len(b); n > 0 && b[n-1] == '\n' {
		if _, err := h.w.WriteString(color); err != nil {
			return err
		}
		if _, err := h.w.Write(b[:n-1]); err != nil {
			return err
		}
		if _, err := h.w.WriteString(ansiReset); err != nil {
			return err
		}
		_, err := h.w.Write([]byte{'\n'})
		return err
	}
	if _, err := h.w.WriteString(color); err != nil {
		return err
	}
	if _, err := h.w.Write(b); err != nil {
		return err
	}
	_, err := h.w.WriteString(ansiReset)
	return err
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

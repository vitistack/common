package vlog

import (
	"os"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Package-level sugared logger with lazy default initialization.
var (
	sugar *zap.SugaredLogger
	once  sync.Once
)

// Options configures the vlog logger.
type Options struct {
	// Level sets the minimum log level. One of: "debug", "info", "warn", "error", "dpanic", "panic", "fatal".
	Level string
	// JSON switches the encoder to JSON instead of console.
	JSON bool
	// AddCaller includes caller information in logs.
	AddCaller bool
	// DisableStacktrace disables automatic stacktraces at Error level and above.
	DisableStacktrace bool
	// ColorizeLine applies ANSI color to the entire log line (console mode only).
	// When true, the level text itself won't be separately colorized to avoid nested codes.
	ColorizeLine bool
}

// Setup initializes the global logger with the provided options.
func Setup(opts Options) error {
	// Default to info when not specified. Set "debug" explicitly if desired.
	level := parseLevel(opts.Level)

	encCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder, // default: plain; we may colorize whole line below
		EncodeTime:     func(t time.Time, enc zapcore.PrimitiveArrayEncoder) { enc.AppendString(t.Format(time.RFC3339)) },
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if opts.JSON {
		// Note: JSON encoder won't use colors; colors are for console encoder only.
		// We keep the same level/time encoders for consistency.
		encCfg.EncodeLevel = zapcore.CapitalLevelEncoder
		encoder = zapcore.NewJSONEncoder(encCfg)
	} else {
		// Console mode: if not coloring the whole line, colorize the level for readability.
		if !opts.ColorizeLine {
			encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		encoder = zapcore.NewConsoleEncoder(encCfg)
	}

	ws := zapcore.AddSync(os.Stdout)
	var core zapcore.Core
	if !opts.JSON && opts.ColorizeLine {
		// Use a custom core that wraps the entire encoded line in a color based on the level.
		core = newColorCore(encoder, ws, level)
	} else {
		core = zapcore.NewCore(encoder, ws, level)
	}

	var zapOpts []zap.Option
	if opts.AddCaller {
		zapOpts = append(zapOpts, zap.AddCaller(), zap.AddCallerSkip(1))
	}
	if opts.DisableStacktrace {
		zapOpts = append(zapOpts, zap.AddStacktrace(zapcore.InvalidLevel))
	}

	z := zap.New(core, zapOpts...)
	sugar = z.Sugar()
	return nil
}

// ensure ensures the logger is initialized with sensible defaults.
func ensure() {
	if sugar != nil {
		return
	}
	once.Do(func() {
		if sugar != nil {
			return
		}
		_ = Setup(Options{ // defaults: console, colored, info level
			Level:        "info",
			JSON:         false,
			AddCaller:    false,
			ColorizeLine: true,
		})
	})
}

// Sync flushes any buffered log entries.
func Sync() error {
	if sugar == nil {
		return nil
	}
	return sugar.Sync()
}

// Logr returns a logr.Logger backed by vlog's zap logger.
// This allows integration with controller-runtime (kubebuilder) via ctrl.SetLogger(vlog.Logr()).
func Logr() logr.Logger {
	ensure()
	// Desugar to get the underlying *zap.Logger and wrap with zapr.
	return zapr.NewLogger(sugar.Desugar())
}

// Debug logs at Debug level. Accepts mixed arguments (strings, structs, numbers, errors, etc.).
func Debug(args ...any) {
	ensure()
	sugar.Debug(args...)
}

// Info logs at Info level. Accepts mixed arguments (strings, structs, numbers, errors, etc.).
func Info(args ...any) {
	ensure()
	sugar.Info(args...)
}

// Warn logs at Warn level. Accepts mixed arguments (strings, structs, numbers, errors, etc.).
func Warn(args ...any) {
	ensure()
	sugar.Warn(args...)
}

// Error logs at Error level. Accepts mixed arguments (strings, structs, numbers, errors, etc.).
// When an error is included among args, it's rendered in the message; for structured error fields use WithError.
func Error(args ...any) {
	ensure()
	sugar.Error(args...)
}

// DPanic logs at DPanic level.
func DPanic(args ...any) { ensure(); sugar.DPanic(args...) }

// Panic logs at Panic level then panics.
func Panic(args ...any) { ensure(); sugar.Panic(args...) }

// Fatal logs at Fatal level then exits.
func Fatal(args ...any) { ensure(); sugar.Fatal(args...) }

// Formatted variants.
func Debugf(format string, args ...any)  { ensure(); sugar.Debugf(format, args...) }
func Infof(format string, args ...any)   { ensure(); sugar.Infof(format, args...) }
func Warnf(format string, args ...any)   { ensure(); sugar.Warnf(format, args...) }
func Errorf(format string, args ...any)  { ensure(); sugar.Errorf(format, args...) }
func DPanicf(format string, args ...any) { ensure(); sugar.DPanicf(format, args...) }
func Panicf(format string, args ...any)  { ensure(); sugar.Panicf(format, args...) }
func Fatalf(format string, args ...any)  { ensure(); sugar.Fatalf(format, args...) }

// With returns a child logger with additional structured context provided as key-value pairs.
// Example: vlog.With("pod", podName, "ns", namespace).Info("created")
func With(keysAndValues ...any) *zap.SugaredLogger {
	ensure()
	return sugar.With(keysAndValues...)
}

// (no helper)

func parseLevel(lvl string) zapcore.LevelEnabler {
	switch lvl {
	case "debug":
		return zap.DebugLevel
	case "info", "":
		return zap.InfoLevel
	case "warn", "warning":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "dpanic":
		return zap.DPanicLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

// color codes
const (
	ansiReset  = "\x1b[0m"
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
	ansiBlue   = "\x1b[34m"
)

// newColorCore creates a Core that colorizes the entire encoded line based on level.
func newColorCore(enc zapcore.Encoder, ws zapcore.WriteSyncer, enab zapcore.LevelEnabler) zapcore.Core {
	return &colorCore{enc: enc, ws: ws, enab: enab}
}

type colorCore struct {
	enc  zapcore.Encoder
	ws   zapcore.WriteSyncer
	enab zapcore.LevelEnabler
}

func (c *colorCore) Enabled(lvl zapcore.Level) bool { return c.enab.Enabled(lvl) }

func (c *colorCore) With(fields []zapcore.Field) zapcore.Core {
	// Clone encoder and add fields to it so they are included in each subsequent entry.
	enc := c.enc.Clone()
	for i := range fields {
		fields[i].AddTo(enc)
	}
	return &colorCore{enc: enc, ws: c.ws, enab: c.enab}
}

//nolint:gocritic // signature is defined by zapcore.Core interface
func (c *colorCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

//nolint:gocritic // signature is defined by zapcore.Core interface
func (c *colorCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	buf, err := c.enc.EncodeEntry(ent, fields)
	if err != nil {
		return err
	}
	// Choose color per level
	color := levelColor(ent.Level)
	b := buf.Bytes()
	// Ensure reset occurs before trailing newline so colors don't bleed.
	var out []byte
	if n := len(b); n > 0 && b[n-1] == '\n' {
		out = make([]byte, 0, len(color)+n+len(ansiReset))
		out = append(out, color...)
		out = append(out, b[:n-1]...)
		out = append(out, ansiReset...)
		out = append(out, '\n')
	} else {
		out = make([]byte, 0, len(color)+len(b)+len(ansiReset))
		out = append(out, color...)
		out = append(out, b...)
		out = append(out, ansiReset...)
	}
	// Write and free buffer
	_, werr := c.ws.Write(out)
	buf.Free()
	return werr
}

func (c *colorCore) Sync() error { return c.ws.Sync() }

func levelColor(lvl zapcore.Level) string {
	switch lvl {
	case zapcore.DebugLevel:
		return ansiBlue
	case zapcore.InfoLevel:
		return ansiGreen
	case zapcore.WarnLevel:
		return ansiYellow
	case zapcore.ErrorLevel, zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return ansiRed
	default:
		return ansiReset
	}
}

package loggers

// Logger defines a small logging interface to allow multiple implementations.
// It's intentionally minimal and aligned with typical needs.
type Logger interface {
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)

	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)

	// With attaches structured key-value pairs and returns a derived logger.
	With(keysAndValues ...any) Logger
}

// Factory creates a Logger with provided options.
// Concrete packages can expose their own Options type and translate here as needed.
type Factory interface {
	New() Logger
}

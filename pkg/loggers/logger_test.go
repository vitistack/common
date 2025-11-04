package loggers

import (
	"testing"
)

// mockLogger implements the Logger interface for testing
type mockLogger struct {
	debugCalls  int
	infoCalls   int
	warnCalls   int
	errorCalls  int
	debugfCalls int
	infofCalls  int
	warnfCalls  int
	errorfCalls int
	withCalls   int
	lastArgs    []any
	lastFormat  string
	lastKVs     []any
}

func (m *mockLogger) Debug(args ...any) {
	m.debugCalls++
	m.lastArgs = args
}

func (m *mockLogger) Info(args ...any) {
	m.infoCalls++
	m.lastArgs = args
}

func (m *mockLogger) Warn(args ...any) {
	m.warnCalls++
	m.lastArgs = args
}

func (m *mockLogger) Error(args ...any) {
	m.errorCalls++
	m.lastArgs = args
}

func (m *mockLogger) Debugf(format string, args ...any) {
	m.debugfCalls++
	m.lastFormat = format
	m.lastArgs = args
}

func (m *mockLogger) Infof(format string, args ...any) {
	m.infofCalls++
	m.lastFormat = format
	m.lastArgs = args
}

func (m *mockLogger) Warnf(format string, args ...any) {
	m.warnfCalls++
	m.lastFormat = format
	m.lastArgs = args
}

func (m *mockLogger) Errorf(format string, args ...any) {
	m.errorfCalls++
	m.lastFormat = format
	m.lastArgs = args
}

func (m *mockLogger) With(keysAndValues ...any) Logger {
	m.withCalls++
	m.lastKVs = keysAndValues
	// Return a new instance to simulate chaining
	return &mockLogger{}
}

func TestLoggerInterface(t *testing.T) {
	logger := &mockLogger{}

	// Verify that mockLogger implements Logger
	var _ Logger = logger

	t.Run("Debug", func(t *testing.T) {
		logger.Debug("test message", 123)
		if logger.debugCalls != 1 {
			t.Errorf("Debug() calls = %d, want 1", logger.debugCalls)
		}
		if len(logger.lastArgs) != 2 {
			t.Errorf("Debug() args length = %d, want 2", len(logger.lastArgs))
		}
	})

	t.Run("Info", func(t *testing.T) {
		logger = &mockLogger{}
		logger.Info("info message")
		if logger.infoCalls != 1 {
			t.Errorf("Info() calls = %d, want 1", logger.infoCalls)
		}
	})

	t.Run("Warn", func(t *testing.T) {
		logger = &mockLogger{}
		logger.Warn("warning")
		if logger.warnCalls != 1 {
			t.Errorf("Warn() calls = %d, want 1", logger.warnCalls)
		}
	})

	t.Run("Error", func(t *testing.T) {
		logger = &mockLogger{}
		logger.Error("error message")
		if logger.errorCalls != 1 {
			t.Errorf("Error() calls = %d, want 1", logger.errorCalls)
		}
	})

	t.Run("Debugf", func(t *testing.T) {
		logger = &mockLogger{}
		logger.Debugf("debug: %s %d", "test", 42)
		if logger.debugfCalls != 1 {
			t.Errorf("Debugf() calls = %d, want 1", logger.debugfCalls)
		}
		if logger.lastFormat != "debug: %s %d" {
			t.Errorf("Debugf() format = %q, want %q", logger.lastFormat, "debug: %s %d")
		}
	})

	t.Run("Infof", func(t *testing.T) {
		logger = &mockLogger{}
		logger.Infof("info: %v", "value")
		if logger.infofCalls != 1 {
			t.Errorf("Infof() calls = %d, want 1", logger.infofCalls)
		}
	})

	t.Run("Warnf", func(t *testing.T) {
		logger = &mockLogger{}
		logger.Warnf("warning: %d", 123)
		if logger.warnfCalls != 1 {
			t.Errorf("Warnf() calls = %d, want 1", logger.warnfCalls)
		}
	})

	t.Run("Errorf", func(t *testing.T) {
		logger = &mockLogger{}
		logger.Errorf("error: %s", "fail")
		if logger.errorfCalls != 1 {
			t.Errorf("Errorf() calls = %d, want 1", logger.errorfCalls)
		}
	})

	t.Run("With", func(t *testing.T) {
		logger = &mockLogger{}
		newLogger := logger.With("key1", "value1", "key2", 2)
		if logger.withCalls != 1 {
			t.Errorf("With() calls = %d, want 1", logger.withCalls)
		}
		if len(logger.lastKVs) != 4 {
			t.Errorf("With() key-value pairs = %d, want 4", len(logger.lastKVs))
		}
		// Verify it returns a Logger
		var _ Logger = newLogger
	})

	t.Run("With returns chainable logger", func(t *testing.T) {
		logger = &mockLogger{}
		chainedLogger := logger.With("key", "value")

		// Should be able to call methods on the returned logger
		chainedLogger.Info("chained message")

		// The returned logger should also be a valid Logger interface
		var _ Logger = chainedLogger
	})
}

// mockFactory implements the Factory interface for testing
type mockFactory struct {
	newCalls int
}

func (m *mockFactory) New() Logger {
	m.newCalls++
	return &mockLogger{}
}

func TestFactoryInterface(t *testing.T) {
	factory := &mockFactory{}

	// Verify that mockFactory implements Factory
	var _ Factory = factory

	t.Run("New creates logger", func(t *testing.T) {
		logger := factory.New()
		if factory.newCalls != 1 {
			t.Errorf("New() calls = %d, want 1", factory.newCalls)
		}

		// Verify returned logger implements Logger interface
		var _ Logger = logger
	})

	t.Run("New can be called multiple times", func(t *testing.T) {
		factory = &mockFactory{}

		logger1 := factory.New()
		logger2 := factory.New()

		if factory.newCalls != 2 {
			t.Errorf("New() calls = %d, want 2", factory.newCalls)
		}

		// Both should be valid loggers
		var _ Logger = logger1
		var _ Logger = logger2
	})
}

func TestLoggerWithVariadicArgs(t *testing.T) {
	logger := &mockLogger{}

	t.Run("handles no arguments", func(t *testing.T) {
		logger.Debug()
		if len(logger.lastArgs) != 0 {
			t.Errorf("Debug() with no args should have empty lastArgs")
		}
	})

	t.Run("handles single argument", func(t *testing.T) {
		logger = &mockLogger{}
		logger.Info("single")
		if len(logger.lastArgs) != 1 {
			t.Errorf("Info() args length = %d, want 1", len(logger.lastArgs))
		}
	})

	t.Run("handles multiple arguments", func(t *testing.T) {
		logger = &mockLogger{}
		logger.Warn("msg", 1, 2, 3, "end")
		if len(logger.lastArgs) != 5 {
			t.Errorf("Warn() args length = %d, want 5", len(logger.lastArgs))
		}
	})
}

func TestLoggerFormattedWithVariadicArgs(t *testing.T) {
	logger := &mockLogger{}

	t.Run("handles no format args", func(t *testing.T) {
		logger.Debugf("no args")
		if len(logger.lastArgs) != 0 {
			t.Errorf("Debugf() with no args should have empty lastArgs")
		}
		if logger.lastFormat != "no args" {
			t.Errorf("Debugf() format = %q, want %q", logger.lastFormat, "no args")
		}
	})

	t.Run("handles multiple format args", func(t *testing.T) {
		logger = &mockLogger{}
		logger.Infof("format %s %d %v", "str", 42, true)
		if len(logger.lastArgs) != 3 {
			t.Errorf("Infof() args length = %d, want 3", len(logger.lastArgs))
		}
	})
}

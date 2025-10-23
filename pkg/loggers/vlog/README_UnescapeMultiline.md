# VLog - UnescapeMultiline Feature

## Overview

The `vlog` package includes an **UnescapeMultiline** feature that transforms escaped JSON strings with `\n` characters into beautifully formatted, actual multi-line output in your logs.

## The Problem

When logging complex JSON structures (especially from Kubernetes resources), the output often looks like this:

```
time=2025-10-23T19:29:28+02:00 level=DEBUG msg="Response Body" body="{\n  \"apiVersion\": \"vitistack.io/v1alpha1\",\n  \"kind\": \"NetworkConfiguration\",\n  \"metadata\": {\n    \"name\": \"test-networkconfiguration\",\n    \"namespace\": \"default\"\n  }\n}"
```

The escaped newlines (`\n`) and quotes (`\"`) make it hard to read!

## The Solution

Enable `UnescapeMultiline: true` in your vlog setup:

```go
vlog.Setup(vlog.Options{
    Level:             "debug",
    JSON:              false,
    ColorizeLine:      true,
    UnescapeMultiline: true,  // ✨ Enable this!
})
```

Now the same log becomes:

```
time=2025-10-23T19:29:28+02:00 level=DEBUG msg="Response Body" body={
  "apiVersion": "vitistack.io/v1alpha1",
  "kind": "NetworkConfiguration",
  "metadata": {
    "name": "test-networkconfiguration",
    "namespace": "default"
  }
}
```

Much better! 🎉

## Usage Examples

### Basic Usage

```go
import (
    "github.com/vitistack/common/pkg/loggers/vlog"
    "github.com/vitistack/common/pkg/serialize"
)

func main() {
    // Setup with UnescapeMultiline enabled
    vlog.Setup(vlog.Options{
        Level:             "debug",
        JSON:              false,
        UnescapeMultiline: true,
    })

    config := MyStruct{...}

    // Use serialize.Pretty() for 2-space indented JSON
    vlog.Info("Config loaded", "config", serialize.Pretty(config))
}
```

### Auto-Formatting (No serialize.Pretty needed)

The vlog package now automatically formats JSON structures:

```go
// Maps are auto-formatted
configMap := map[string]any{
    "database": "postgres",
    "port": 5432,
}
vlog.Info("Database config", "config", configMap)

// Structs are auto-formatted
vlog.Info("User created", "user", userStruct)

// JSON strings are auto-formatted
jsonStr := `{"status":"ok","data":{"id":1}}`
vlog.Info("API response", "response", jsonStr)
```

### Controller-Runtime Integration

Perfect for Kubernetes operators:

```go
func (r *MyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := vlog.With(
        "controller", "mycontroller",
        "namespace", req.Namespace,
        "name", req.Name,
    )

    // Log the full resource with beautiful formatting
    logger.Debug("Reconciling resource", "resource", serialize.Pretty(myResource))

    return ctrl.Result{}, nil
}
```

### Multiple JSON Attributes

```go
vlog.Info("Resource comparison",
    "current", serialize.Pretty(currentState),
    "desired", serialize.Pretty(desiredState),
    "diff", serialize.Pretty(differences),
)
```

## When to Use

### ✅ Use UnescapeMultiline when:

- Developing/debugging locally
- You need to read complex JSON structures
- Logging Kubernetes resources
- Console/terminal output (not JSON logging mode)
- Human readability is important

### ❌ Don't use UnescapeMultiline when:

- Logging to structured log aggregators (use `JSON: true` instead)
- Performance is critical (adds small overhead)
- You need machine-parseable logs
- Logs are being processed by scripts

## Performance

The `UnescapeMultiline` feature:

- **Fast path**: If no `\n` escapes are detected, returns immediately (minimal overhead)
- **Overhead**: Only processes lines that contain escaped newlines
- **Benchmarked**: See `vlog_bench_test.go` for performance measurements

```bash
# Run benchmarks
go test ./pkg/loggers/vlog/... -bench=.
```

## Configuration Options

```go
type Options struct {
    Level             string  // "debug", "info", "warn", "error"
    JSON              bool    // true = JSON output, false = text output
    AddCaller         bool    // Include file:line in logs
    ColorizeLine      bool    // Colorize entire log lines
    UnescapeMultiline bool    // Transform \n to real newlines (text mode only)
}
```

## Tips

1. **Combine with ColorizeLine**: Makes logs even more readable

   ```go
   vlog.Setup(vlog.Options{
       ColorizeLine:      true,
       UnescapeMultiline: true,
   })
   ```

2. **Use serialize.Pretty()**: For explicit control

   ```go
   vlog.Info("data", "obj", serialize.Pretty(obj))    // JSON with 2-space indent
   ```

3. **Use serialize.YAML()**: For YAML output

   ```go
   vlog.Info("data", "obj", serialize.YAML(obj))      // YAML format
   ```

4. **Auto-formatting**: Works automatically for maps, slices, and structs
   ```go
   vlog.Info("data", "obj", myStruct)  // Auto-formatted!
   ```

## Related Functions

### serialize package

```go
serialize.Pretty(v any) string         // JSON with 2-space indent
serialize.JSON(v any) string           // Compact JSON
serialize.JSONIndentN(v any, n int)    // JSON with n-space indent
serialize.YAML(v any) string           // YAML format
serialize.PrettyYAML(v any) string     // Alias for YAML
```

### vlog package

```go
vlog.Pretty(v any)    // Wraps value for explicit pretty-printing
vlog.With(kv ...any)  // Create logger with context
```

## Complete Example

```go
package main

import (
    "github.com/vitistack/common/pkg/loggers/vlog"
    "github.com/vitistack/common/pkg/serialize"
)

type Config struct {
    Database DatabaseConfig `json:"database"`
    Features []string       `json:"features"`
}

type DatabaseConfig struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}

func main() {
    // Setup vlog with pretty-printing enabled
    vlog.Setup(vlog.Options{
        Level:             "debug",
        JSON:              false,
        AddCaller:         true,
        ColorizeLine:      true,
        UnescapeMultiline: true,
    })

    config := Config{
        Database: DatabaseConfig{
            Host: "localhost",
            Port: 5432,
        },
        Features: []string{"auth", "api", "cache"},
    }

    // All of these will be beautifully formatted:
    vlog.Info("Config loaded", "config", serialize.Pretty(config))
    vlog.Debug("Database config", "db", config.Database)
    vlog.Info("Features", "list", config.Features)
}
```

## Summary

The `UnescapeMultiline` feature makes your logs **human-readable** by converting escaped JSON strings into properly formatted multi-line output. Combined with auto-formatting and `serialize.Pretty()`, you get beautiful, easy-to-read logs for local development and debugging! 🚀

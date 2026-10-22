# Quick Reference: Prettifying Logs with vlog

## TL;DR

```go
// Enable pretty-printed logs
vlog.Setup(vlog.Options{
    Level:             "debug",
    UnescapeMultiline: true,  // ← This makes logs readable!
    ColorizeLine:      true,   // ← Optional: adds colors
})

// Log with auto-formatting (no extra work needed!)
vlog.Info("Resource", "data", myStruct)

// Or use serialize.Pretty() for explicit control
vlog.Info("Resource", "data", serialize.Pretty(myStruct))
```

## Before vs After

### Before (UnescapeMultiline: false)

```
time=2025-10-23T19:29:28+02:00 level=DEBUG body="{\n  \"name\": \"test\",\n  \"value\": 123\n}"
```

### After (UnescapeMultiline: true)

```
time=2025-10-23T19:29:28+02:00 level=DEBUG body={
  "name": "test",
  "value": 123
}
```

## Common Use Cases

### Kubernetes Operators

```go
vlog.Setup(vlog.Options{
    Level:             os.Getenv("LOG_LEVEL"),
    JSON:              os.Getenv("LOG_JSON") == "true",
    UnescapeMultiline: os.Getenv("LOG_PRETTY") == "true",
    ColorizeLine:      os.Getenv("LOG_COLOR") == "true",
})

// In your reconciler
logger := vlog.With("controller", "mycontroller", "namespace", req.Namespace)
logger.Debug("Reconciling", "resource", serialize.Pretty(resource))
```

### API Responses

```go
vlog.Info("API Response",
    "status", resp.StatusCode,
    "body", serialize.Pretty(responseBody),
)
```

### Configuration Debugging

```go
vlog.Debug("Configuration loaded",
    "database", serialize.Pretty(cfg.Database),
    "features", cfg.Features,
)
```

## Functions

| Function                     | Purpose                  | Example                        |
| ---------------------------- | ------------------------ | ------------------------------ |
| `serialize.Pretty(v)`        | JSON with 2-space indent | `serialize.Pretty(config)`     |
| `serialize.YAML(v)`          | YAML format              | `serialize.YAML(config)`       |
| `vlog.With(kv...)`           | Add context to logger    | `vlog.With("req_id", id)`      |
| `vlog.Info/Debug/Warn/Error` | Log at level             | `vlog.Info("msg", "key", val)` |

## Environment Variables

```bash
# Typical dev setup
export LOG_LEVEL=debug
export LOG_JSON=false
export LOG_PRETTY=true
export LOG_COLOR=true
```

## When NOT to Use

❌ Don't use `UnescapeMultiline: true` in production if:

- You're sending logs to a structured log aggregator (use `JSON: true`)
- You need parseable logs for automation
- Performance is critical

✅ Perfect for:

- Local development
- Debugging
- Reading complex JSON structures
- Terminal output

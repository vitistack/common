# Go Libraries

Vitistack Common provides a set of small, focused Go libraries for common tasks in Kubernetes operators and cloud-native applications.

## Installation

```bash
go get github.com/vitistack/common@latest
```

## Available Libraries

- [vlog](#vlog---structured-logging) - Structured logging with Zap
- [serialize](#serialize---json-helpers) - JSON serialization helpers
- [k8sclient](#k8sclient---kubernetes-client) - Kubernetes client initialization
- [crdcheck](#crdcheck---crd-validation) - CRD prerequisite checking
- [dotenv](#dotenv---environment-configuration) - Smart .env file loading

---

## vlog - Structured Logging

Thin Zap setup with nice console colors and JSON mode, plus a logr adapter for controller-runtime.

### Quick Start

```go
import "github.com/vitistack/common/pkg/loggers/vlog"

func main() {
	_ = vlog.Setup(vlog.Options{
		Level:             "info", // debug|info|warn|error|dpanic|panic|fatal
		JSON:              true,    // default: structured JSON (fastest to parse)
		AddCaller:         true,    // include caller file:line
		DisableStacktrace: false,
		ColorizeLine:      false,   // set true only for human console viewing
		UnescapeMultiline: false,   // set true only if you need pretty multi-line msg rendering in text mode
	})
	defer vlog.Sync()

	vlog.Info("hello world")
	vlog.Debugf("count=%d", 42)
	vlog.With("user", "alice", "req", 123).Warn("something odd")
}
```

### Options

- **Level**: `string` — one of `debug`, `info`, `warn` (or `warning`), `error`, `dpanic`, `panic`, `fatal` (default: `info`)
- **JSON**: `bool` — switch to JSON encoder (no ANSI colors). Default: `true`
- **AddCaller**: `bool` — include short caller information
- **DisableStacktrace**: `bool` — turn off auto stack traces at Error+
- **ColorizeLine**: `bool` — when using console encoder, colorize the entire line by level
- **UnescapeMultiline**: `bool` — when using console text mode, turn escaped '\n' inside msg="..." into real multi-line output (costs a tiny bit of CPU). Default: `false`

### Use with controller-runtime (Kubebuilder)

`vlog` exposes a logr-compatible adapter so you can wire it into controller-runtime.

```go
import (
	ctrl "sigs.k8s.io/controller-runtime"
	"github.com/vitistack/common/pkg/loggers/vlog"
)

func main() {
	_ = vlog.Setup(vlog.Options{Level: "info", ColorizeLine: true, AddCaller: true})
	defer func() {
		_ = vlog.Sync()
	}()

	ctrl.SetLogger(vlog.Logr())

	// proceed with manager setup ...
}
```

### Advanced Usage

```go
// Create a logger with fields
logger := vlog.With("component", "database", "version", "1.0")
logger.Info("connection established")

// Use structured fields
vlog.With(
	"user_id", 123,
	"action", "login",
	"ip", "192.168.1.1",
).Info("user logged in")

// Format strings
vlog.Infof("processing %d items", count)
vlog.Warnf("retry attempt %d of %d", attempt, maxRetries)

// Error logging
if err != nil {
	vlog.Error("failed to connect", err)
}
```

### Best Practices

1. **Initialize once** at application startup
2. **Use structured fields** instead of string formatting for better log parsing
3. **Enable JSON mode** in production for log aggregation tools
4. **Use ColorizeLine** only for local development
5. **Call Sync()** before application exit to flush buffers

See [vlog documentation](../pkg/loggers/vlog/) for more details.

---

## serialize - JSON Helpers

`serialize` aims to make quick JSON rendering trivial for logs and debug prints.

### Quick Start

```go
import "github.com/vitistack/common/pkg/serialize"

// Compact JSON string
s := serialize.JSON(map[string]any{"a": 1, "b": "x"})

// Pretty with 2-space indent
pretty := serialize.Pretty(map[string]any{"a": 1, "b": []int{1,2,3}})

// Pretty with N spaces
pretty4 := serialize.JSONIndentN(struct{ X int }{X: 7}, 4)

// Pretty with custom indent string
tab := serialize.JSONIndent(map[string]string{"k": "v"}, "\t")

// Conditional: indent when n>0, otherwise compact
dyn := serialize.As([]int{1,2,3}, 0)    // compact
dyn2 := serialize.As([]int{1,2,3}, 2)   // indented 2 spaces

// Bytes variants
b, err := serialize.BytesJSON(map[string]int{"n": 10})
bi, err := serialize.BytesJSONIndent(map[string]int{"n": 10}, "  ")
```

### Notes

- On marshal error, string helpers return a best-effort `fmt` representation with the error appended
- Bytes helpers return the error so you can handle it explicitly
- Useful for logging complex objects during debugging

### Usage Examples

```go
// Debug logging
vlog.Debugf("config: %s", serialize.Pretty(config))

// HTTP response
w.Header().Set("Content-Type", "application/json")
w.Write([]byte(serialize.JSON(response)))

// Conditional formatting based on environment
indent := 0
if os.Getenv("ENV") == "development" {
	indent = 2
}
output := serialize.As(data, indent)
```

---

## k8sclient - Kubernetes Client

Quick client-go setup using the in-cluster or KUBECONFIG context.

### Quick Start

```go
import (
	"context"
	"github.com/vitistack/common/pkg/clients/k8sclient"
	"github.com/vitistack/common/pkg/loggers/vlog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	_ = vlog.Setup(vlog.Options{ Level: "info", ColorizeLine: true })
	defer vlog.Sync()

	k8sclient.Init()

	pods, err := k8sclient.Kubernetes.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		vlog.Error("list pods failed", err)
		return
	}
	vlog.Infof("pods in default: %d", len(pods.Items))
}
```

### Features

- Automatic detection of in-cluster vs out-of-cluster configuration
- KUBECONFIG support for local development
- Initializes both typed and dynamic clients
- Error handling with helpful messages

### Usage Patterns

```go
// Initialize once at startup
k8sclient.Init()

// Use typed client
deployments, err := k8sclient.Kubernetes.AppsV1().
	Deployments("default").
	List(context.Background(), metav1.ListOptions{})

// Use dynamic client
k8sclient.Dynamic // Available for unstructured access

// Check configuration
if k8sclient.Config != nil {
	fmt.Printf("Connected to: %s\n", k8sclient.Config.Host)
}
```

See `cmd/examples/main.go` for a runnable sample combining `vlog`, `serialize`, and the k8s client.

---

## crdcheck - CRD Validation

Verify a set of CRD-backed API resources are served by the cluster before your operator starts reconciling. It uses the Discovery API and will log and panic when required CRDs are missing.

### Quick Start

```go
import (
	"context"
	"github.com/vitistack/common/pkg/clients/k8sclient"
	"github.com/vitistack/common/pkg/loggers/vlog"
	"github.com/vitistack/common/pkg/operator/crdcheck"
)

func main() {
	_ = vlog.Setup(vlog.Options{Level: "info", ColorizeLine: true})
	defer vlog.Sync()

	// Initialize k8s clients (KUBECONFIG or in-cluster)
	k8sclient.Init()

	// Panic if these CRDs/resources are not available
	crdcheck.MustEnsureInstalled(context.TODO(),
		crdcheck.Ref{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"},
		crdcheck.Ref{Group: "vitistack.io", Version: "v1alpha1", Resource: "machines"},
		crdcheck.Ref{Group: "vitistack.io", Version: "v1alpha1", Resource: "kubernetesclusters"},
	)

	// continue with manager/controllers...
}
```

### Features

- Validates CRD availability at startup
- Prevents operator from starting with missing dependencies
- Clear error messages indicating which CRDs are missing
- Uses Kubernetes Discovery API

### Best Practices

1. Call during operator initialization, before starting controllers
2. Include all CRDs your operator depends on
3. Use fully qualified resource names (plural form)
4. Consider adding a timeout context for cluster connectivity issues

---

## dotenv - Environment Configuration

Smart `.env` file loading with environment-specific overrides and upward directory searching. It follows the principle of not overriding existing OS environment variables.

### Quick Start

```go
import "github.com/vitistack/common/pkg/settings/dotenv"

func main() {
	// Load .env files - call this early in main()
	dotenv.LoadDotEnv()

	// Now you can use environment variables as usual
	dbURL := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")
}
```

### How it works

1. **File discovery**: Searches for `.env` files starting from the current working directory and executable directory, walking upwards until found

2. **Environment-specific files**: Environment-specific `.env-<ENV>` files (e.g., `.env-production`, `.env-development`) are **only loaded when the `ENV` environment variable is set**. Without setting `ENV`, only the base `.env` file will be loaded

3. **Load order and precedence**:

   - Base `.env` file is loaded first
   - Environment-specific `.env-<ENV>` file overrides base file values
   - Existing OS environment variables are **never overridden**

4. **Example file structure**:
   ```
   project/
   ├── .env              # Base configuration
   ├── .env-development  # Development overrides
   ├── .env-production   # Production overrides
   └── cmd/app/main.go
   ```

### Example .env files

**.env** (base configuration):

```bash
PORT=3000
DATABASE_URL=postgresql://localhost:5432/myapp
LOG_LEVEL=info
DEBUG=false
```

**.env-development**:

```bash
DATABASE_URL=postgresql://localhost:5432/myapp_dev
LOG_LEVEL=debug
DEBUG=true
```

**.env-production**:

```bash
DATABASE_URL=postgresql://prod-server:5432/myapp_prod
LOG_LEVEL=warn
```

### Usage patterns

**Basic usage** (loads only `.env`):

```go
func main() {
	dotenv.LoadDotEnv()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback
	}
}
```

**With environment switching** (loads `.env` + `.env-<ENV>`):

```bash
# Development - loads .env and .env-development
ENV=development go run main.go

# Production - loads .env and .env-production
ENV=production go run main.go
```

> **Important**: Environment-specific files like `.env-production` are only loaded when you set the `ENV` environment variable. Without `ENV` set, only the base `.env` file will be loaded.

**Integration with vlog**:

```go
import (
	"os"
	"github.com/vitistack/common/pkg/settings/dotenv"
	"github.com/vitistack/common/pkg/loggers/vlog"
)

func main() {
	// Load environment first
	dotenv.LoadDotEnv()

	// Configure logging based on environment
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	_ = vlog.Setup(vlog.Options{
		Level: logLevel,
		JSON:  os.Getenv("ENV") == "production",
	})
	defer vlog.Sync()

	vlog.Info("application starting")
}
```

### Notes

- Files are searched upwards from both the current working directory and the executable directory
- Missing `.env` files are silently ignored
- Successfully loaded files are logged at info level showing their paths
- Based on [joho/godotenv](https://github.com/joho/godotenv) library

---

## Complete Example

Here's a complete example combining multiple libraries:

```go
package main

import (
	"context"
	"os"

	"github.com/vitistack/common/pkg/clients/k8sclient"
	"github.com/vitistack/common/pkg/loggers/vlog"
	"github.com/vitistack/common/pkg/operator/crdcheck"
	"github.com/vitistack/common/pkg/serialize"
	"github.com/vitistack/common/pkg/settings/dotenv"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	// Load environment configuration
	dotenv.LoadDotEnv()

	// Setup logging
	_ = vlog.Setup(vlog.Options{
		Level: os.Getenv("LOG_LEVEL"),
		JSON:  os.Getenv("ENV") == "production",
		AddCaller: true,
	})
	defer vlog.Sync()

	vlog.Info("application starting")

	// Initialize Kubernetes client
	k8sclient.Init()

	// Ensure required CRDs exist
	crdcheck.MustEnsureInstalled(context.Background(),
		crdcheck.Ref{Group: "vitistack.io", Version: "v1alpha1", Resource: "machines"},
	)

	// List and log resources
	machines, err := k8sclient.Kubernetes.CoreV1().
		Pods("default").
		List(context.Background(), metav1.ListOptions{})

	if err != nil {
		vlog.Error("failed to list machines", err)
		return
	}

	vlog.With("count", len(machines.Items)).Info("machines listed")
	vlog.Debugf("machines: %s", serialize.Pretty(machines))
}
```

## Troubleshooting

### vlog

- **No logs appear**: Ensure `vlog.Setup` is called before use and `Level` is high enough
- **Colors not showing**: `ColorizeLine` only affects console encoder; if `JSON: true`, output is uncolored
- **controller-runtime logs still plain**: Verify `ctrl.SetLogger(vlog.Logr())` is called before creating the manager

### k8sclient

- **Connection refused**: Check KUBECONFIG is set correctly for out-of-cluster usage
- **Permission denied**: Ensure service account has necessary RBAC permissions in-cluster
- **Client is nil**: Call `k8sclient.Init()` before using the client

### crdcheck

- **Panic on startup**: CRDs are not installed. Install them with `kubectl apply -f crds/` or Helm
- **Timeout errors**: Check cluster connectivity and increase context timeout

### dotenv

- **Variables not loading**: Ensure `.env` file exists and is readable
- **Variables not overriding**: Check if OS environment variable is already set (dotenv won't override)
- **Environment-specific file not loading**: Set the `ENV` environment variable

## Additional Resources

- [Examples](../examples/) - Working code examples
- [API Reference](https://pkg.go.dev/github.com/vitistack/common) - Full Go package documentation
- [vlog Package](../pkg/loggers/vlog/) - Detailed vlog documentation

# Vitistack common

[![Build and Test](https://github.com/vitistack/common/actions/workflows/build-and-test.yml/badge.svg)](https://github.com/vitistack/common/actions/workflows/build-and-test.yml)
[![Security Scan](https://github.com/vitistack/common/actions/workflows/security-scan.yml/badge.svg)](https://github.com/vitistack/common/actions/workflows/security-scan.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/vitistack/common)](https://goreportcard.com/report/github.com/vitistack/common)
[![Go Reference](https://pkg.go.dev/badge/github.com/vitistack/common.svg)](https://pkg.go.dev/github.com/vitistack/common)

Shared components for Vitistack infrastructure management:

### Custom Resource Definitions (CRDs)
- **Infrastructure as Code**: Declarative management of machines, networks, and Kubernetes clusters
- **Multi-Provider Support**: Proxmox, KubeVirt, and more
- See [CRD Documentation](docs/crds.md) for complete reference

### Go Libraries
- **Logging**: `vlog` — thin Zap setup with nice console colors and JSON mode, plus a logr adapter for controller-runtime
- **Serialization**: `serialize` — tiny helpers to turn Go values into JSON strings/bytes quickly
- **K8s client helper**: `k8sclient` — convenience initializer around client-go/controller-runtime config
- **Operator utils**: `crdcheck` — verify required CRDs/API resources exist during startup and panic if missing
- **Environment config**: `dotenv` — smart .env file loading with environment-specific overrides and directory traversal

## Install

### CRDs (Helm)

```bash
helm install vitistack-crds oci://ghcr.io/vitistack/vitistack-crds --version 0.1.0
```

### Go Library

```bash
go get github.com/vitistack/common@latest
```

## Logging with vlog

Initialize once at startup, then use the package-level functions. Defaults now favor structured JSON output for efficiency.

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

- Level: string — one of `debug`, `info`, `warn` (or `warning`), `error`, `dpanic`, `panic`, `fatal` (default: `info`).
- JSON: bool — switch to JSON encoder (no ANSI colors). Default: true.
- AddCaller: bool — include short caller information.
- DisableStacktrace: bool — turn off auto stack traces at Error+.
- ColorizeLine: bool — when using console encoder, colorize the entire line by level.
- UnescapeMultiline: bool — when using console text mode, turn escaped '\n' inside msg="..." into real multi-line output (costs a tiny bit of CPU). Default: false.

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

## JSON helpers with serialize

`serialize` aims to make quick JSON rendering trivial for logs and debug prints.

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

Notes:

- On marshal error, string helpers return a best-effort `fmt` representation with the error appended.
- Bytes helpers return the error so you can handle it explicitly.

## Kubernetes client helper

If you want a quick client-go setup using the in-cluster or KUBECONFIG context, use `k8sclient`:

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

See `cmd/examples/main.go` for a runnable sample combining `vlog`, `serialize`, and the k8s client.

## Operator prerequisite: ensure CRDs installed

Use `crdcheck` to verify a set of CRD-backed API resources are served by the cluster before your operator starts reconciling. It uses the Discovery API and will log and panic when required CRDs are missing.

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
		crdcheck.Ref{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}, // example core CRD API
		crdcheck.Ref{Group: "example.com", Version: "v1alpha1", Resource: "widgets"},                      // your CRD plural
	)

	// continue with manager/controllers...
}
```

## Environment configuration with dotenv

The `dotenv` package provides smart `.env` file loading with environment-specific overrides and upward directory searching. It follows the principle of not overriding existing OS environment variables.

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

1. **File discovery**: Searches for `.env` files starting from the current working directory and executable directory, walking upwards until found.

2. **Environment-specific files**: Environment-specific `.env-<ENV>` files (e.g., `.env-production`, `.env-development`) are **only loaded when the `ENV` environment variable is set**. Without setting `ENV`, only the base `.env` file will be loaded.

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

## Troubleshooting

- No logs appear: ensure `vlog.Setup` is called before use and `Level` is high enough.
- Colors not showing: `ColorizeLine` only affects console encoder; if `JSON: true`, output is uncolored.
- controller-runtime logs still plain: verify `ctrl.SetLogger(vlog.Logr())` is called before creating the manager.

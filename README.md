# Vitistack common

Small, focused helpers shared across Vitistack projects:

- Logging: `vlog` — thin Zap setup with nice console colors and JSON mode, plus a logr adapter for controller-runtime.
- Serialization: `serialize` — tiny helpers to turn Go values into JSON strings/bytes quickly.
- K8s client helper: `k8sclient` — convenience initializer around client-go/controller-runtime config.
- Operator utils: `crdcheck` — verify required CRDs/API resources exist during startup and panic if missing.

## Install

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

## Troubleshooting

- No logs appear: ensure `vlog.Setup` is called before use and `Level` is high enough.
- Colors not showing: `ColorizeLine` only affects console encoder; if `JSON: true`, output is uncolored.
- controller-runtime logs still plain: verify `ctrl.SetLogger(vlog.Logr())` is called before creating the manager.

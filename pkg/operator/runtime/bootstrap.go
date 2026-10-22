package runtime

import (
	"net/http"

	"github.com/vitistack/common/pkg/loggers/vlog"
	krt "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ManagerOptions wraps a subset of ctrl.Options for convenience.
type ManagerOptions struct {
	Scheme         *krt.Scheme
	LeaderElection bool
}

// NewManagerWithDefaults sets up vlog as the logger and builds a controller-runtime manager
// with health and ready probes registered.
func NewManagerWithDefaults(cfg *rest.Config, o ManagerOptions) (ctrl.Manager, error) {
	// Ensure vlog is the logger for controller-runtime
	ctrl.SetLogger(vlog.Logr())

	opts := ctrl.Options{
		Scheme:         o.Scheme,
		LeaderElection: o.LeaderElection,
	}
	mgr, err := ctrl.NewManager(cfg, opts)
	if err != nil {
		return nil, err
	}
	// Register default health and ready checks
	_ = mgr.AddHealthzCheck("ping", func(_ *http.Request) error { return nil })
	_ = mgr.AddReadyzCheck("ping", func(_ *http.Request) error { return nil })
	return mgr, nil
}

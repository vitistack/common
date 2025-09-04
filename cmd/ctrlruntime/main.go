package main

import (
    "github.com/vitistack/common/pkg/loggers/vlog"
    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Minimal controller-runtime bootstrap showing vlog integration via ctrl.SetLogger.
func main() {
    // Initialize vlog with your preferences
    _ = vlog.Setup(vlog.Options{
        Level:        "info",
        JSON:         false,
        AddCaller:    true,
        ColorizeLine: true,
    })
    defer func() { _ = vlog.Sync() }()

    // Wire vlog into controller-runtime
    ctrl.SetLogger(vlog.Logr())

    // Get kube rest.Config (respects in-cluster or KUBECONFIG)
    cfg, err := config.GetConfig()
    if err != nil {
        vlog.Error("unable to load kube config", err)
        return
    }

    // Minimal Scheme; add your APIs as needed
    scheme := runtime.NewScheme()

    // Create manager (use defaults; customize as needed)
    mgr, err := ctrl.NewManager(cfg, ctrl.Options{
        Scheme:         scheme,
        LeaderElection: false,
    })
    if err != nil {
        vlog.Error("unable to create manager", err)
        return
    }

    // Start manager until signaled to stop
    ctx := ctrl.SetupSignalHandler()
    if err := mgr.Start(ctx); err != nil {
        // Use Fatal to exit non-zero
        vlog.Fatalf("manager exited: %v", err)
    }

    // Manager stopped cleanly
}

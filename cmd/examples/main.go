package main

import (
	"context"

	"github.com/vitistack/common/pkg/clients/k8sclient"
	"github.com/vitistack/common/pkg/loggers/vlog"
	"github.com/vitistack/common/pkg/serialize"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	// Initialize the logger
	err := vlog.Setup(vlog.Options{
		Level:        "debug", // debug|info|warn|error|dpanic|panic|fatal
		ColorizeLine: true,    // whole-line color
		JSON:         false,   // console output (supports ANSI colors)
		AddCaller:    true,
	})
	if err != nil {
		panic(err)
	}
	defer func() { _ = vlog.Sync() }()

	vlog.Info("This is an info message")
	vlog.Debug("This is a debug message")
	vlog.Warn("This is a warning message")
	vlog.Error("This is an error message")

	test := 42
	vlog.Debug("Log line ", "with extra parameters ", test)

	// Initialize Kubernetes client
	k8sclient.Init()
	vlog.Info("Kubernetes client initialized successfully")
	pods, err := k8sclient.Kubernetes.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		vlog.Error("Failed to list pods", err)
		return
	}
	for i := range pods.Items {
		pod := pods.Items[i]
		vlog.Debug("Pod:", pod.Name, "Pod labels:", serialize.Pretty(pod.Labels))
	}
	vlog.Info("Number of pods in default namespace:", len(pods.Items))
}

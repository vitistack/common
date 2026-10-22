package k8sclient

import (
	"github.com/vitistack/common/pkg/loggers/vlog"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	Kubernetes      *kubernetes.Clientset
	DiscoveryClient *discovery.DiscoveryClient

	DynamicClient dynamic.Interface
)

func Init() {
	k8sconfig, err := config.GetConfig()
	if err != nil {
		vlog.Error("Failed to get Kubernetes config:", err)
		panic(err)
	}

	Kubernetes, err = kubernetes.NewForConfig(k8sconfig)
	if err != nil {
		vlog.Error("Failed to create Kubernetes client:", err)
		panic(err)
	}

	DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(k8sconfig)
	if err != nil {
		vlog.Error("Failed to create Kubernetes discovery client:", err)
		panic(err)
	}

	DynamicClient, err = dynamic.NewForConfig(k8sconfig)
	if err != nil {
		vlog.Error("Failed to create Kubernetes dynamic client:", err)
		panic(err)
	}
}

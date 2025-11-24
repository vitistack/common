package v1alpha1

import (
	"slices"
)

const (
	ClusterNameAnnotation        = "vitistack.io/clustername"              // The name of the cluster
	ClusterWorkspaceAnnotation   = "vitistack.io/clusterworkspace"         // The workspace of the cluster
	CountryAnnotation            = "vitistack.io/country"                  // The country of the cluster
	RegionAnnotation             = "vitistack.io/region"                   // The region of the cluster
	InfrastructureAnnotation     = "vitistack.io/infrastructure"           // The infrastructure of the cluster
	AzAnnotation                 = "vitistack.io/az"                       // The availability zone of the cluster
	VMProviderAnnotation         = "vitistack.io/vmprovider"               // The VM provider of the cluster
	VMIdAnnotation               = "vitistack.io/vmid"                     // The VM ID of the cluster
	KubernetesProviderAnnotation = "vitistack.io/kubernetesprovider"       // The Kubernetes provider of the cluster
	ClusterIdAnnotation          = "vitistack.io/clusterid"                // The ID of the cluster, this is the uuid in ror
	K8sEndpointAnnotation        = "vitistack.io/kubernetes-endpoint-addr" // The endpoint of the kubernetes api server
	ClusterFQDNAnnotation        = "vitistack.io/cluster-fqdn"             // The FQDN of the cluster
	NodeFQDNAnnotation           = "vitistack.io/node-fqdn"                // The FQDN of the node
	ManagedByAnnotation          = "vitistack.io/managed-by"               // The operator managing the resource

)

func GetAllAnnotations() []string {
	return []string{
		ClusterNameAnnotation,
		ClusterWorkspaceAnnotation,
		CountryAnnotation,
		RegionAnnotation,
		InfrastructureAnnotation,
		AzAnnotation,
		VMProviderAnnotation,
		VMIdAnnotation,
		KubernetesProviderAnnotation,
		ClusterIdAnnotation,
		K8sEndpointAnnotation,
		ClusterFQDNAnnotation,
		NodeFQDNAnnotation,
	}
}

func IsVitisStackAnnotation(key string) bool {
	return slices.Contains(GetAllAnnotations(), key)
}

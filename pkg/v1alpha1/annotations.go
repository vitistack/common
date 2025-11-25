package v1alpha1

import (
	"slices"
)

const (
	// The ID of the cluster, this is the uuid in ror
	ClusterIdAnnotation = "vitistack.io/clusterid"

	// The name of the cluster
	ClusterNameAnnotation = "vitistack.io/clustername"

	// The workspace of the cluster
	ClusterWorkspaceAnnotation = "vitistack.io/clusterworkspace"

	// The FQDN of the cluster
	ClusterFQDNAnnotation = "vitistack.io/cluster-fqdn"

	// The country of the cluster
	CountryAnnotation = "vitistack.io/country"

	// The region of the cluster
	RegionAnnotation = "vitistack.io/region"

	// The availability zone of the cluster
	AzAnnotation = "vitistack.io/az"

	// The infrastructure of the cluster
	InfrastructureAnnotation = "vitistack.io/infrastructure"

	// The VM provider of the cluster
	VMProviderAnnotation = "vitistack.io/vmprovider"

	// The VM ID of the cluster
	VMIdAnnotation = "vitistack.io/vmid"

	// The Kubernetes provider of the cluster
	KubernetesProviderAnnotation = "vitistack.io/kubernetesprovider"

	// The endpoint of the kubernetes api server
	K8sEndpointAnnotation = "vitistack.io/kubernetes-endpoint-addr"

	// The FQDN of the node
	NodeFQDNAnnotation = "vitistack.io/node-fqdn"

	// The role of the node
	NodeRoleAnnotation = "vitistack.io/node-role"

	// The name of the node pool
	NodePoolAnnotation = "vitistack.io/nodepool"

	// The operator managing the resource
	ManagedByAnnotation = "vitistack.io/managed-by"
)

func GetAllVitistackAnnotations() []string {
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
	return slices.Contains(GetAllVitistackAnnotations(), key)
}

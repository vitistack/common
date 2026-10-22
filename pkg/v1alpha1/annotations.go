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

	// The project of the cluster
	ClusterProjectAnnotation = "vitistack.io/clusterproject"

	// The environment of the cluster (e.g. dev, test, qa, prod)
	EnvironmentAnnotation = "vitistack.io/environment"

	// The infrastructure tier of the machine (defaults to "prod")
	MachineInfrastructureAnnotation = "vitistack.io/machineinfrastructure"

	// The machine provider (e.g. kubevirt, proxmox)
	MachineProviderAnnotation = "vitistack.io/machineprovider"

	// The machine class (e.g. small, medium, large)
	MachineClassAnnotation = "vitistack.io/machineclass"

	// The machine ID (name of the machine in the provider)
	MachineIdAnnotation = "vitistack.io/machineid"

	// Deprecated: use MachineInfrastructureAnnotation instead. Will be removed in a future release.
	InfrastructureAnnotation = "vitistack.io/infrastructure"

	// Deprecated: use MachineProviderAnnotation instead. Will be removed in a future release.
	VMProviderAnnotation = "vitistack.io/vmprovider"

	// Deprecated: use MachineIdAnnotation instead. Will be removed in a future release.
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
		ClusterIdAnnotation,
		ClusterNameAnnotation,
		ClusterWorkspaceAnnotation,
		ClusterProjectAnnotation,
		ClusterFQDNAnnotation,
		EnvironmentAnnotation,
		CountryAnnotation,
		RegionAnnotation,
		AzAnnotation,
		KubernetesProviderAnnotation,
		K8sEndpointAnnotation,
		MachineProviderAnnotation,
		MachineClassAnnotation,
		MachineIdAnnotation,
		MachineInfrastructureAnnotation,
		NodeFQDNAnnotation,
		NodeRoleAnnotation,
		NodePoolAnnotation,
		// Deprecated: kept for backward compatibility during transition
		InfrastructureAnnotation,
		VMProviderAnnotation,
		VMIdAnnotation,
	}
}

func IsVitisStackAnnotation(key string) bool {
	return slices.Contains(GetAllVitistackAnnotations(), key)
}

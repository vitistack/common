package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KubernetesCluster is the Schema for the Machines API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=kubernetesclusters,scope=Namespaced,shortName=kc
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`,description="The phase of the Kubernetes cluster"
// +kubebuilder:printcolumn:name="Provider",type=string,JSONPath=`.spec.data.provider`,description="The cloud provider of the Kubernetes cluster"
// +kubebuilder:printcolumn:name="Region",type=string,JSONPath=`.spec.data.region`,description="The region of the Kubernetes cluster"
// +kubebuilder:printcolumn:name="ControlPlaneReplicas",type=integer,JSONPath=`.spec.topology.controlplane.replicas`,description="The number of control plane replicas"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="The age of the Kubernetes cluster"
type KubernetesCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubernetesClusterSpec   `json:"spec,omitempty"`
	Status KubernetesClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// MachineList contains a list of Machine
type KubernetesClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubernetesCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubernetesCluster{}, &KubernetesClusterList{})
}

type KubernetesClusterSpec struct {
	// +kubebuilder:validation:Required
	Cluster KubernetesClusterSpecData `json:"data,omitzero"`

	// +kubebuilder:validation:Required
	Topology KubernetesClusterSpecTopology `json:"topology,omitzero"`
}

type KubernetesClusterSpecData struct {
	ClusterUID string `json:"clusterUid"` // ClusterUID is a unique identifier for the cluster, e.g., "12345678-1234-1234-1234-123456789012"

	// +kubebuilder:validation:Required
	ClusterId string `json:"clusterId"`

	// +kubebuilder:validation:Required
	Provider   KubernetesProviderType `json:"provider"`
	Datacenter string                 `json:"datacenter"`

	// +kubebuilder:validation:Required
	Region string `json:"region"`

	// +kubebuilder:validation:Required
	Zone      string `json:"zone"`
	Project   string `json:"project"`
	Workspace string `json:"workspace"`
	Workorder string `json:"workorder"`

	// +kubebuilder:validation:Required
	Environment string `json:"environment"`
}

type KubernetesClusterSpecTopology struct {
	Version string `json:"version"` // Kubernetes version, e.g., "1.23.0"

	// +kubebuilder:validation:Required
	ControlPlane KubernetesClusterSpecControlPlane `json:"controlplane"` // ControlPlane contains the control plane configuration.

	Workers KubernetesClusterWorkers `json:"workers"` // Workers contains the worker nodes configuration.
}

type KubernetesClusterSpecControlPlane struct {
	// +kubebuilder:validation:Required
	Replicas int    `json:"replicas"`
	Version  string `json:"version"` // Kubernetes version, e.g., "1.23.0"

	// +kubebuilder:validation:Required
	Provider KubernetesProviderType `json:"provider"`

	MachineClass string                               `json:"machineClass"`
	Metadata     KubernetesClusterSpecMetadataDetails `json:"metadata"`
	Storage      []KubernetesClusterStorage           `json:"storage"`
}

type KubernetesClusterSpecMetadataDetails struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

type KubernetesClusterStorage struct {
	Class string `json:"class"`
	Path  string `json:"path"`
	Size  string `json:"size"`
}

type KubernetesClusterWorkers struct {
	NodePools []KubernetesClusterNodePool `json:"nodePools"`
}

type KubernetesClusterNodePool struct {
	MachineClass string                               `json:"machineClass"`
	Provider     KubernetesProviderType               `json:"provider"`
	Version      string                               `json:"version"` // Kubernetes version, e.g., "1.23.0"
	Name         string                               `json:"name"`
	Replicas     int                                  `json:"replicas"`
	Autoscaling  KubernetesClusterAutoscalingSpec     `json:"autoscaling"`
	Metadata     KubernetesClusterSpecMetadataDetails `json:"metadata"`
	Taint        []KubernetesClusterTaint             `json:"taint"`
	Storage      []KubernetesClusterStorage           `json:"storage"`
}

type KubernetesClusterTaint struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Effect string `json:"effect"`
}

type KubernetesClusterAutoscalingConfig struct {
	Enabled     bool `json:"enabled"`
	MinReplicas int  `json:"minReplicas"`
	MaxReplicas int  `json:"maxReplicas"`
}
type KubernetesClusterAutoscalingSpec struct {
	KubernetesClusterAutoscalingConfig `json:",inline"`
	ScalingRules                       []string `json:"scalingRules"`
}

// KubernetesClusterStatus represents the status of a Kubernetes cluster.
// It contains the current state, phase, and conditions of the cluster.
type KubernetesClusterStatus struct {
	State      KubernetesClusterClusterState `json:"state"`
	Phase      string                        `json:"phase"` // Provisioning, Running, Deleting, Failed, Updating
	Conditions []KubernetesClusterCondition  `json:"conditions"`
}

type KubernetesClusterClusterState struct {
	Cluster       KubernetesClusterClusterDetails `json:"cluster"`
	Versions      []KubernetesClusterVersion      `json:"versions"`
	Endpoints     []KubernetesClusterEndpoint     `json:"endpoints"`
	EgressIP      string                          `json:"egressIP"`
	LastUpdated   metav1.Time                     `json:"lastUpdated"`
	LastUpdatedBy string                          `json:"lastUpdatedBy"`
	Created       metav1.Time                     `json:"created"`
}

type KubernetesClusterEndpoint struct {
	Name    string `json:"name"`    // Name is the name of the endpoint, e.g., "controllplane", "kubernetes", "api", "dashboard, grafana, argocd", "datacenter"
	Address string `json:"address"` // Address is the address of the endpoint, e.g., "https://api.example.com", "http://dashboard.example.com"
}

type KubernetesClusterStatusCondition struct {
	Type               string `json:"type" example:"ClusterReady"`                                   // Type is the type of the condition. For example, "ready", "available", etc.
	Status             string `json:"status"  example:"ok" enums:"ok,warning,error,working,unknown"` // Status is the status of the condition. Valid vales are: ok, warning, error, working, unknown.
	LastTransitionTime string `json:"lastTransitionTime"`                                            // LastTransitionTime is the last time the condition transitioned from one status to another.
	Reason             string `json:"reason"`                                                        // Reason is a brief reason for the condition's last transition.
	Message            string `json:"message"`                                                       // Message is a human-readable message indicating details about the condition.
}

type KubernetesClusterStatusPrice struct {
	Monthly int `json:"monthly"` // Monthly is the monthly price of the cluster in your currency, e.g., "1000"
	Yearly  int `json:"yearly"`  // Yearly is the yearly price of the cluster, e.g., "12000"
}

type KubernetesClusterClusterDetails struct {
	ExternalId         string                                        `json:"externalId"`
	Resources          KubernetesClusterStatusClusterStatusResources `json:"resources"`
	Price              KubernetesClusterStatusPrice                  `json:"price"` // Price is the price of the cluster, e.g., "1000 NOK/month"
	ControlPlaneStatus KubernetesClusterControlPlaneStatus           `json:"controlplane"`
	NodePools          []KubernetesClusterNodePoolStatus             `json:"nodepools"` // TODO
}

type KubernetesClusterStatusClusterStatusResources struct {
	CPU    KubernetesClusterStatusClusterStatusResource `json:"cpu,omitzero"`    // CPU is the total CPU capacity of the cluster, if not specified in millicores, e.g., "16 cores", "8000 millicores"
	Memory KubernetesClusterStatusClusterStatusResource `json:"memory,omitzero"` // Memory is the total memory capacity of the cluster, if not specified in bytes, e.g., "64 GB", "128000 MB", "25600000000 bytes"
	GPU    KubernetesClusterStatusClusterStatusResource `json:"gpu,omitzero"`    // GPU is the total GPU capacity of the cluster, if not specified in number of GPUs"
	Disk   KubernetesClusterStatusClusterStatusResource `json:"disk,omitzero"`   // Disk is the total disk capacity of the cluster, if not specified in bytes"
}

type KubernetesClusterStatusClusterStatusResource struct {
	Capacity  resource.Quantity `json:"capacity"`   // Capacity is the total capacity of the resource."
	Used      resource.Quantity `json:"used"`       // Used is the amount of the resource that is currently used."
	Percetage int               `json:"percentage"` // Percentage is the percentage of the resource that is currently used as an int.
}

type KubernetesClusterControlPlaneStatus struct {
	Status       string                                        `json:"status"`
	Message      string                                        `json:"message"`
	Scale        int                                           `json:"scale"`        // Scale is the number of replicas of the control plane.
	MachineClass string                                        `json:"machineClass"` // MachineClass is the machine class of the control plane, e.g., "c5.large", "m5.xlarge"
	Resources    KubernetesClusterStatusClusterStatusResources `json:"resources"`    // Resources is the resources of the control plane, e.g., CPU, Memory, Disk, GPU
	Nodes        []string                                      `json:"nodes"`        // Nodes is the list of the uuids of the nodes in the control plane
}

type KubernetesClusterNodePoolStatus struct {
	Name         string                                        `json:"name"`
	Status       string                                        `json:"status"`
	Message      string                                        `json:"message"`
	Scale        int                                           `json:"scale"`        // Scale is the number of replicas of the nodepool.
	MachineClass string                                        `json:"machineClass"` // MachineClass is the machine class of the nodepool, e.g., "c5.large", "m5.xlarge"
	Autoscaling  KubernetesClusterAutoscalingConfig            `json:"autoscaling"`  // Autoscaling is the autoscaling configuration of the node pool.
	Resources    KubernetesClusterStatusClusterStatusResources `json:"resources"`    // Resources is the resources of the node pool, e.g., CPU, Memory, Disk, GPU
	Nodes        []string                                      `json:"nodes"`        // Nodes is the list of the uuids of the nodes in the node pool
}

type KubernetesClusterVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Branch  string `json:"branch"`
}

type KubernetesClusterCondition struct {
	Type               string `json:"type" example:"ClusterReady"`                                   // Type is the type of the condition. For example, "ready", "available", etc.
	Status             string `json:"status"  example:"ok" enums:"ok,warning,error,working,unknown"` // Status is the status of the condition. Valid vales are: ok, warning, error, working, unknown.
	LastTransitionTime string `json:"lastTransitionTime"`                                            // LastTransitionTime is the last time the condition transitioned from one status to another.
	Reason             string `json:"reason"`                                                        // Reason is a brief reason for the condition's last transition.
	Message            string `json:"message"`                                                       // Message is a human-readable message indicating details about the condition.
}

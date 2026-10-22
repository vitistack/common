package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineProvider is the Schema for the MachineProviders API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=machineproviders,scope=Cluster,shortName=mp
// +kubebuilder:printcolumn:name="Provider",type=string,JSONPath=`.spec.providerType`
// +kubebuilder:printcolumn:name="Region",type=string,JSONPath=`.spec.region`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type MachineProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineProviderSpec   `json:"spec,omitempty"`
	Status MachineProviderStatus `json:"status,omitempty"`
}

// MachineProviderSpec defines the desired state of MachineProvider
type MachineProviderSpec struct {
	// Provider type (aws, azure, gcp, vsphere, openstack, libvirt, proxmox)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=kubevirt;proxmox;aws;azure;gcp;vsphere;openstack;libvirt;proxmox;cloudstack;nutanix;ovirt
	ProviderType string `json:"providerType"`

	// Human-readable name for this provider instance
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	DisplayName string `json:"displayName"`

	// Region where this provider operates
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Region string `json:"region"`

	// Available zones in this region
	Zones []string `json:"zones,omitempty"`

	// Provider-specific endpoint configuration
	Endpoint ProviderEndpoint `json:"endpoint,omitempty"`

	// Authentication configuration
	Authentication ProviderAuthentication `json:"authentication"`

	// Provider capabilities and limits
	Capabilities ProviderCapabilities `json:"capabilities,omitempty"`

	// Network configuration for this provider
	Network ProviderNetworkConfig `json:"network,omitempty"`

	// Storage configuration for this provider
	Storage ProviderStorageConfig `json:"storage,omitempty"`

	// Compute configuration and limits
	Compute ProviderComputeConfig `json:"compute,omitempty"`

	// Default tags to apply to all resources
	DefaultTags map[string]string `json:"defaultTags,omitempty"`

	// Provider-specific configuration
	Config map[string]string `json:"config,omitempty"`
}

type ProviderEndpoint struct {
	// Primary endpoint URL
	URL string `json:"url,omitempty"`

	// Whether to skip TLS verification
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`

	// Custom CA certificate bundle
	CABundle string `json:"caBundle,omitempty"`

	// Connection timeout in seconds
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=300
	TimeoutSeconds int `json:"timeoutSeconds,omitempty"`

	// Number of retry attempts
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10
	RetryAttempts int `json:"retryAttempts,omitempty"`
}

type ProviderAuthentication struct {
	// Type of authentication (credentials, serviceAccount, instanceProfile, managedIdentity)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=credentials;serviceAccount;instanceProfile;managedIdentity;certificate;token
	Type string `json:"type"`

	// Reference to credentials secret
	CredentialsRef *CredentialsReference `json:"credentialsRef,omitempty"`

	// Service account configuration (for GCP, Azure, etc.)
	ServiceAccount *ServiceAccountConfig `json:"serviceAccount,omitempty"`

	// Additional authentication parameters
	Parameters map[string]string `json:"parameters,omitempty"`
}

type ServiceAccountConfig struct {
	// Service account email or ID
	AccountID string `json:"accountID"`

	// Key file content or reference to secret
	KeyRef *CredentialsReference `json:"keyRef,omitempty"`

	// Scopes for the service account
	Scopes []string `json:"scopes,omitempty"`
}

type ProviderCapabilities struct {
	// Supported machineclasses
	MachineClasses []MachineClass `json:"machineClasses,omitempty"`

	// Supported operating systems
	OperatingSystems []OSInfo `json:"operatingSystems,omitempty"`

	// Supported storage types
	StorageTypes []StorageTypeInfo `json:"storageTypes,omitempty"`

	// Supported network features
	NetworkFeatures []string `json:"networkFeatures,omitempty"`

	// Maximum number of machines per provider
	// +kubebuilder:validation:Minimum=1
	MaxMachines int `json:"maxMachines,omitempty"`

	// Whether this provider supports auto-scaling
	AutoScaling bool `json:"autoScaling,omitempty"`

	// Whether this provider supports load balancers
	LoadBalancers bool `json:"loadBalancers,omitempty"`

	// Whether this provider supports persistent volumes
	PersistentVolumes bool `json:"persistentVolumes,omitempty"`
}

type OSInfo struct {
	// OS family
	Family string `json:"family"`

	// OS distribution
	Distribution string `json:"distribution"`

	// Available versions
	Versions []string `json:"versions"`

	// Supported architectures
	Architectures []string `json:"architectures"`

	// Default image ID
	DefaultImageID string `json:"defaultImageID,omitempty"`
}

type StorageTypeInfo struct {
	// Storage type name
	Name string `json:"name"`

	// Display name
	DisplayName string `json:"displayName,omitempty"`

	// Storage class (SSD, HDD, NVMe)
	Class string `json:"class"`

	// IOPS range
	IOPSRange *IOPSRange `json:"iopsRange,omitempty"`

	// Throughput range in MB/s
	ThroughputRange *ThroughputRange `json:"throughputRange,omitempty"`

	// Whether encryption is supported
	EncryptionSupported bool `json:"encryptionSupported,omitempty"`
}

type IOPSRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type ThroughputRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type ProviderNetworkConfig struct {
	// Default VPC/Virtual Network
	DefaultVPC string `json:"defaultVPC,omitempty"`

	// Available VPCs/Virtual Networks
	AvailableVPCs []VPCInfo `json:"availableVPCs,omitempty"`

	// Default security groups
	DefaultSecurityGroups []string `json:"defaultSecurityGroups,omitempty"`

	// Whether public IPs are supported
	PublicIPSupport bool `json:"publicIPSupport,omitempty"`

	// Whether IPv6 is supported
	IPv6Support bool `json:"ipv6Support,omitempty"`

	// Whether load balancers are supported
	LoadBalancerSupport bool `json:"loadBalancerSupport,omitempty"`
}

type VPCInfo struct {
	// VPC ID
	ID string `json:"id"`

	// VPC name
	Name string `json:"name"`

	// CIDR block
	CIDR string `json:"cidr"`

	// Available subnets
	Subnets []SubnetInfo `json:"subnets,omitempty"`
}

type SubnetInfo struct {
	// Subnet ID
	ID string `json:"id"`

	// Subnet name
	Name string `json:"name"`

	// CIDR block
	CIDR string `json:"cidr"`

	// Availability zone
	Zone string `json:"zone"`

	// Whether this is a public subnet
	Public bool `json:"public"`
}

type ProviderStorageConfig struct {
	// Default storage type
	DefaultType string `json:"defaultType,omitempty"`

	// Available storage classes
	StorageClasses []StorageClassInfo `json:"storageClasses,omitempty"`

	// Whether encryption is enabled by default
	DefaultEncryption bool `json:"defaultEncryption,omitempty"`

	// Maximum storage size in GB
	// +kubebuilder:validation:Minimum=1
	MaxStorageGB int `json:"maxStorageGB,omitempty"`
}

type StorageClassInfo struct {
	// Storage class name
	Name string `json:"name"`

	// Display name
	DisplayName string `json:"displayName,omitempty"`

	// Provisioner
	Provisioner string `json:"provisioner"`

	// Parameters
	Parameters map[string]string `json:"parameters,omitempty"`

	// Whether this is the default storage class
	Default bool `json:"default,omitempty"`
}

type ProviderComputeConfig struct {
	// Default machine class
	DefaultMachineClass string `json:"defaultMachineClass,omitempty"`

	// Maximum CPU cores per machine
	// +kubebuilder:validation:Minimum=1
	MaxCPUs int `json:"maxCPUs,omitempty"`

	// Maximum memory in GB per machine
	// +kubebuilder:validation:Minimum=1
	MaxMemoryGB int `json:"maxMemoryGB,omitempty"`

	// Whether GPU instances are available
	GPUSupport bool `json:"gpuSupport,omitempty"`

	// Available GPU types
	GPUTypes []string `json:"gpuTypes,omitempty"`

	// Whether nested virtualization is supported
	NestedVirtualization bool `json:"nestedVirtualization,omitempty"`
}

// MachineProviderStatus defines the observed state of MachineProvider
type MachineProviderStatus struct {
	// Current phase of the provider (Pending, Ready, Failed, Offline)
	Phase string `json:"phase,omitempty"`

	// Human-readable status message
	Message string `json:"message,omitempty"`

	// Last time the provider was verified/tested
	LastVerified *metav1.Time `json:"lastVerified,omitempty"`

	// Current quota usage
	Quota ProviderQuotaStatus `json:"quota,omitempty"`

	// Health check information
	Health ProviderHealthStatus `json:"health,omitempty"`

	// Available resources
	AvailableResources ProviderResourcesStatus `json:"availableResources,omitempty"`

	// Active machines count
	ActiveMachines int `json:"activeMachines,omitempty"`

	// Provider-specific status information
	ProviderStatus map[string]string `json:"providerStatus,omitempty"`

	// Conditions represent the latest available observations
	Conditions []ProviderCondition `json:"conditions,omitempty"`
}

type ProviderQuotaStatus struct {
	// CPU quota (total cores)
	CPUQuota int `json:"cpuQuota,omitempty"`

	// CPU usage (used cores)
	CPUUsed int `json:"cpuUsed,omitempty"`

	// Memory quota in GB
	MemoryQuotaGB int `json:"memoryQuotaGB,omitempty"`

	// Memory usage in GB
	MemoryUsedGB int `json:"memoryUsedGB,omitempty"`

	// Storage quota in GB
	StorageQuotaGB int `json:"storageQuotaGB,omitempty"`

	// Storage usage in GB
	StorageUsedGB int `json:"storageUsedGB,omitempty"`

	// Instance quota
	InstanceQuota int `json:"instanceQuota,omitempty"`

	// Instance usage
	InstanceUsed int `json:"instanceUsed,omitempty"`

	// Network quota (e.g., VPCs, Security Groups)
	NetworkQuota map[string]int `json:"networkQuota,omitempty"`

	// Network usage
	NetworkUsed map[string]int `json:"networkUsed,omitempty"`
}

type ProviderHealthStatus struct {
	// Overall health status (Healthy, Degraded, Unhealthy)
	Status string `json:"status,omitempty"`

	// API connectivity status
	APIConnectivity string `json:"apiConnectivity,omitempty"`

	// Authentication status
	Authentication string `json:"authentication,omitempty"`

	// Service availability by region/zone
	ServiceAvailability map[string]string `json:"serviceAvailability,omitempty"`

	// Last health check time
	LastCheck *metav1.Time `json:"lastCheck,omitempty"`

	// Response time in milliseconds
	ResponseTimeMs int `json:"responseTimeMs,omitempty"`
}

type ProviderResourcesStatus struct {
	// Available machine classes
	MachineClasses []string `json:"machineClasses,omitempty"`

	// Available zones
	Zones []string `json:"zones,omitempty"`

	// Available storage types
	StorageTypes []string `json:"storageTypes,omitempty"`

	// Available images
	Images []ImageInfo `json:"images,omitempty"`

	// Resource limits
	Limits ProviderLimits `json:"limits,omitempty"`
}

type ImageInfo struct {
	// Image ID
	ID string `json:"id"`

	// Image name
	Name string `json:"name"`

	// OS family
	OSFamily string `json:"osFamily"`

	// OS distribution
	OSDistribution string `json:"osDistribution"`

	// OS version
	OSVersion string `json:"osVersion"`

	// Architecture
	Architecture string `json:"architecture"`

	// Whether this is a public image
	Public bool `json:"public"`

	// Creation date
	CreationDate *metav1.Time `json:"creationDate,omitempty"`
}

type ProviderLimits struct {
	// Maximum machines per zone
	MaxMachinesPerZone int `json:"maxMachinesPerZone,omitempty"`

	// Maximum storage per machine in GB
	MaxStoragePerMachineGB int `json:"maxStoragePerMachineGB,omitempty"`

	// Maximum network interfaces per machine
	MaxNetworkInterfaces int `json:"maxNetworkInterfaces,omitempty"`

	// Rate limits
	RateLimits map[string]string `json:"rateLimits,omitempty"`
}

type ProviderCondition struct {
	// Type of condition
	Type string `json:"type"`

	// Status of the condition (True, False, Unknown)
	Status string `json:"status"`

	// Last time the condition transitioned
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`

	// Reason for the condition's last transition
	Reason string `json:"reason"`

	// Human-readable message
	Message string `json:"message"`
}

// Common provider phases
const (
	MachineProviderPhasePending = "Pending"
	MachineProviderPhaseReady   = "Ready"
	MachineProviderPhaseFailed  = "Failed"
	MachineProviderPhaseOffline = "Offline"
)

// Common provider condition types
const (
	MachineProviderConditionReady         = "Ready"
	MachineProviderConditionHealthy       = "Healthy"
	MachineProviderConditionAuthenticated = "Authenticated"
	MachineProviderConditionConnected     = "Connected"
	MachineProviderConditionQuotaValid    = "QuotaValid"
)

// Health status values
const (
	ProviderHealthStatusHealthy   = "Healthy"
	ProviderHealthStatusDegraded  = "Degraded"
	ProviderHealthStatusUnhealthy = "Unhealthy"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// MachineProviderList contains a list of MachineProvider
type MachineProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachineProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MachineProvider{}, &MachineProviderList{})
}

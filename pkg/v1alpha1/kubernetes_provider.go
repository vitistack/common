package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KubernetesProvider is the Schema for the KubernetesProviders API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=kubernetesproviders,scope=Cluster,shortName=kp
// +kubebuilder:printcolumn:name="Provider",type=string,JSONPath=`.spec.providerType`
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Nodes",type=string,JSONPath=`.status.nodeCount`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type KubernetesProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubernetesProviderSpec   `json:"spec,omitempty"`
	Status KubernetesProviderStatus `json:"status,omitempty"`
}

// KubernetesProviderSpec defines the desired state of KubernetesProvider
type KubernetesProviderSpec struct {
	// Provider type (eks: talos, gke, aks, openshift, rancher, k3s, kubeadm, managed)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=aks;talos
	ProviderType string `json:"providerType"`

	// Human-readable name for this Kubernetes provider
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	DisplayName string `json:"displayName"`

	// Kubernetes version
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=^v?[0-9]+\.[0-9]+\.[0-9]+$
	Version string `json:"version"`

	// Region where this Kubernetes cluster operates
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Region string `json:"region"`

	// Available zones for node placement
	Zones []string `json:"zones,omitempty"`

	// Cluster configuration
	Cluster KubernetesClusterConfig `json:"cluster"`

	// Node pool configurations
	NodePools []NodePoolConfig `json:"nodePools,omitempty"`

	// Network configuration
	Network KubernetesNetworkConfig `json:"network,omitempty"`

	// Addons configuration
	Addons KubernetesAddonsConfig `json:"addons,omitempty"`

	// Security configuration
	Security KubernetesSecurityConfig `json:"security,omitempty"`

	// Monitoring and logging configuration
	Observability KubernetesObservabilityConfig `json:"observability,omitempty"`

	// Backup and disaster recovery configuration
	Backup KubernetesBackupConfig `json:"backup,omitempty"`

	// Machine provider reference for node provisioning
	MachineProviderRef *MachineProviderReference `json:"machineProviderRef,omitempty"`

	// Authentication configuration
	Authentication KubernetesAuthConfig `json:"authentication,omitempty"`

	// Default tags to apply to all resources
	DefaultTags map[string]string `json:"defaultTags,omitempty"`

	// Provider-specific configuration
	Config map[string]string `json:"config,omitempty"`
}

type KubernetesClusterConfig struct {
	// Cluster endpoint URL (for managed clusters)
	Endpoint string `json:"endpoint,omitempty"`

	// API server configuration
	APIServer APIServerConfig `json:"apiServer,omitempty"`

	// ETCD configuration
	ETCD ETCDConfig `json:"etcd,omitempty"`

	// Control plane configuration
	ControlPlane ControlPlaneConfig `json:"controlPlane,omitempty"`

	// DNS configuration
	DNS DNSConfig `json:"dns,omitempty"`

	// Feature gates
	FeatureGates map[string]bool `json:"featureGates,omitempty"`

	// Additional API server arguments
	APIServerExtraArgs map[string]string `json:"apiServerExtraArgs,omitempty"`

	// Additional controller manager arguments
	ControllerManagerExtraArgs map[string]string `json:"controllerManagerExtraArgs,omitempty"`

	// Additional scheduler arguments
	SchedulerExtraArgs map[string]string `json:"schedulerExtraArgs,omitempty"`
}

type APIServerConfig struct {
	// Enable audit logging
	AuditLogging bool `json:"auditLogging,omitempty"`

	// Audit log retention days
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=365
	AuditLogRetentionDays int `json:"auditLogRetentionDays,omitempty"`

	// Enable encryption at rest
	EncryptionAtRest bool `json:"encryptionAtRest,omitempty"`

	// Enable admission plugins
	AdmissionPlugins []string `json:"admissionPlugins,omitempty"`

	// Disable admission plugins
	DisableAdmissionPlugins []string `json:"disableAdmissionPlugins,omitempty"`
}

type ETCDConfig struct {
	// ETCD version
	Version string `json:"version,omitempty"`

	// Enable backup
	BackupEnabled bool `json:"backupEnabled,omitempty"`

	// Backup schedule (cron format)
	BackupSchedule string `json:"backupSchedule,omitempty"`

	// Backup retention period in days
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=365
	BackupRetentionDays int `json:"backupRetentionDays,omitempty"`

	// Enable encryption
	Encryption bool `json:"encryption,omitempty"`
}

type ControlPlaneConfig struct {
	// Number of control plane nodes
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	Replicas int `json:"replicas,omitempty"`

	// Control plane machine class
	MachineClass string `json:"machineClass,omitempty"`

	// Control plane disk size in GB
	// +kubebuilder:validation:Minimum=20
	// +kubebuilder:validation:Maximum=1000
	DiskSizeGB int `json:"diskSizeGB,omitempty"`
}

type DNSConfig struct {
	// DNS provider (coredns, kube-dns)
	Provider string `json:"provider,omitempty"`

	// DNS domain
	Domain string `json:"domain,omitempty"`

	// Upstream DNS servers
	UpstreamServers []string `json:"upstreamServers,omitempty"`
}

type NodePoolConfig struct {
	// Node pool name
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Machine class
	// +kubebuilder:validation:Required
	MachineClass string `json:"machineClass"`

	// Minimum number of nodes
	// +kubebuilder:validation:Minimum=0
	MinNodes int `json:"minNodes"`

	// Maximum number of nodes
	// +kubebuilder:validation:Minimum=1
	MaxNodes int `json:"maxNodes"`

	// Desired number of nodes
	// +kubebuilder:validation:Minimum=0
	DesiredNodes int `json:"desiredNodes"`

	// Node labels
	Labels map[string]string `json:"labels,omitempty"`

	// Node taints
	Taints []NodeTaint `json:"taints,omitempty"`

	// Zones for this node pool
	Zones []string `json:"zones,omitempty"`

	// Auto-scaling configuration
	AutoScaling *NodePoolAutoScaling `json:"autoScaling,omitempty"`

	// Node configuration
	NodeConfig NodeConfig `json:"nodeConfig,omitempty"`
}

type NodeTaint struct {
	Key    string `json:"key"`
	Value  string `json:"value,omitempty"`
	Effect string `json:"effect"`
}

type NodePoolAutoScaling struct {
	// Enable auto-scaling
	Enabled bool `json:"enabled"`

	// Scale down delay after node add
	ScaleDownDelayAfterAdd string `json:"scaleDownDelayAfterAdd,omitempty"`

	// Scale down delay after delete
	ScaleDownDelayAfterDelete string `json:"scaleDownDelayAfterDelete,omitempty"`

	// Scale down delay after failure
	ScaleDownDelayAfterFailure string `json:"scaleDownDelayAfterFailure,omitempty"`

	// Scale down unneeded time
	ScaleDownUnneededTime string `json:"scaleDownUnneededTime,omitempty"`
}

type NodeConfig struct {
	// Operating system configuration
	OS NodeOSConfig `json:"os,omitempty"`

	// Disk configuration
	Disk NodeDiskConfig `json:"disk,omitempty"`

	// Network configuration
	Network NodeNetworkConfig `json:"network,omitempty"`

	// Security configuration
	Security NodeSecurityConfig `json:"security,omitempty"`

	// Kubelet configuration
	Kubelet KubeletConfig `json:"kubelet,omitempty"`
}

type NodeOSConfig struct {
	// Image ID or family
	Image string `json:"image,omitempty"`

	// OS type (linux, windows)
	Type string `json:"type,omitempty"`

	// User data script
	UserData string `json:"userData,omitempty"`
}

type NodeDiskConfig struct {
	// Root disk size in GB
	// +kubebuilder:validation:Minimum=20
	// +kubebuilder:validation:Maximum=2000
	SizeGB int `json:"sizeGB,omitempty"`

	// Disk type
	Type string `json:"type,omitempty"`

	// Enable encryption
	Encrypted bool `json:"encrypted,omitempty"`
}

type NodeNetworkConfig struct {
	// Subnet IDs
	SubnetIDs []string `json:"subnetIDs,omitempty"`

	// Security group IDs
	SecurityGroupIDs []string `json:"securityGroupIDs,omitempty"`

	// Assign public IP
	AssignPublicIP bool `json:"assignPublicIP,omitempty"`
}

type NodeSecurityConfig struct {
	// SSH key pairs
	SSHKeyPairs []string `json:"sshKeyPairs,omitempty"`

	// IAM instance profile
	IAMInstanceProfile string `json:"iamInstanceProfile,omitempty"`

	// Security groups
	SecurityGroups []string `json:"securityGroups,omitempty"`
}

type KubeletConfig struct {
	// Additional kubelet arguments
	ExtraArgs map[string]string `json:"extraArgs,omitempty"`

	// Max pods per node
	// +kubebuilder:validation:Minimum=10
	// +kubebuilder:validation:Maximum=250
	MaxPods int `json:"maxPods,omitempty"`
}

type KubernetesNetworkConfig struct {
	// CNI plugin (calico, flannel, weave, cilium, antrea)
	CNIPlugin string `json:"cniPlugin,omitempty"`

	// Pod CIDR
	PodCIDR string `json:"podCIDR,omitempty"`

	// Service CIDR
	ServiceCIDR string `json:"serviceCIDR,omitempty"`

	// DNS service IP
	DNSServiceIP string `json:"dnsServiceIP,omitempty"`

	// Network policy support
	NetworkPolicy bool `json:"networkPolicy,omitempty"`

	// Load balancer configuration
	LoadBalancer LoadBalancerConfig `json:"loadBalancer,omitempty"`
}

type LoadBalancerConfig struct {
	// Load balancer type (classic, network, application)
	Type string `json:"type,omitempty"`

	// Load balancer class
	Class string `json:"class,omitempty"`

	// Additional annotations
	Annotations map[string]string `json:"annotations,omitempty"`
}

type KubernetesAddonsConfig struct {
	// Ingress controller configuration
	IngressController *IngressControllerConfig `json:"ingressController,omitempty"`

	// Storage configuration
	Storage *StorageConfig `json:"storage,omitempty"`

	// Monitoring configuration
	Monitoring *MonitoringConfig `json:"monitoring,omitempty"`

	// Service mesh configuration
	ServiceMesh *ServiceMeshConfig `json:"serviceMesh,omitempty"`

	// Additional addons
	Additional []AddonConfig `json:"additional,omitempty"`
}

type IngressControllerConfig struct {
	// Ingress controller type (nginx, traefik, istio, haproxy)
	Type string `json:"type,omitempty"`

	// Enable ingress controller
	Enabled bool `json:"enabled"`

	// Configuration parameters
	Config map[string]string `json:"config,omitempty"`
}

type StorageConfig struct {
	// Default storage class
	DefaultClass string `json:"defaultClass,omitempty"`

	// Storage classes
	Classes []StorageClassConfig `json:"classes,omitempty"`
}

type StorageClassConfig struct {
	Name        string            `json:"name"`
	Provisioner string            `json:"provisioner"`
	Parameters  map[string]string `json:"parameters,omitempty"`
	Default     bool              `json:"default,omitempty"`
}

type MonitoringConfig struct {
	// Enable Prometheus
	Prometheus bool `json:"prometheus,omitempty"`

	// Enable Grafana
	Grafana bool `json:"grafana,omitempty"`

	// Enable AlertManager
	AlertManager bool `json:"alertManager,omitempty"`

	// Configuration parameters
	Config map[string]string `json:"config,omitempty"`
}

type ServiceMeshConfig struct {
	// Service mesh type (istio, linkerd, consul)
	Type string `json:"type,omitempty"`

	// Enable service mesh
	Enabled bool `json:"enabled"`

	// Configuration parameters
	Config map[string]string `json:"config,omitempty"`
}

type AddonConfig struct {
	Name    string            `json:"name"`
	Version string            `json:"version,omitempty"`
	Enabled bool              `json:"enabled"`
	Config  map[string]string `json:"config,omitempty"`
}

type KubernetesSecurityConfig struct {
	// Enable Pod Security Standards
	PodSecurityStandards bool `json:"podSecurityStandards,omitempty"`

	// Pod Security Standard level (privileged, baseline, restricted)
	PodSecurityLevel string `json:"podSecurityLevel,omitempty"`

	// Enable Network Policies
	NetworkPolicies bool `json:"networkPolicies,omitempty"`

	// Enable RBAC
	RBAC bool `json:"rbac,omitempty"`

	// Enable image scanning
	ImageScanning bool `json:"imageScanning,omitempty"`

	// Runtime security configuration
	RuntimeSecurity *RuntimeSecurityConfig `json:"runtimeSecurity,omitempty"`
}

type RuntimeSecurityConfig struct {
	// Enable runtime protection
	Enabled bool `json:"enabled"`

	// Runtime security provider (falco, twistlock, aqua)
	Provider string `json:"provider,omitempty"`

	// Configuration parameters
	Config map[string]string `json:"config,omitempty"`
}

type KubernetesObservabilityConfig struct {
	// Logging configuration
	Logging *LoggingConfig `json:"logging,omitempty"`

	// Metrics configuration
	Metrics *MetricsConfig `json:"metrics,omitempty"`

	// Tracing configuration
	Tracing *TracingConfig `json:"tracing,omitempty"`
}

type LoggingConfig struct {
	// Enable centralized logging
	Enabled bool `json:"enabled"`

	// Logging backend (elasticsearch, loki, splunk)
	Backend string `json:"backend,omitempty"`

	// Log retention period in days
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=365
	RetentionDays int `json:"retentionDays,omitempty"`

	// Configuration parameters
	Config map[string]string `json:"config,omitempty"`
}

type MetricsConfig struct {
	// Enable metrics collection
	Enabled bool `json:"enabled"`

	// Metrics backend (prometheus, datadog, newrelic)
	Backend string `json:"backend,omitempty"`

	// Metrics retention period in days
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=365
	RetentionDays int `json:"retentionDays,omitempty"`

	// Configuration parameters
	Config map[string]string `json:"config,omitempty"`
}

type TracingConfig struct {
	// Enable distributed tracing
	Enabled bool `json:"enabled"`

	// Tracing backend (jaeger, zipkin, datadog)
	Backend string `json:"backend,omitempty"`

	// Sampling rate (0.0 to 1.0) as string for cross-language compatibility
	// +kubebuilder:validation:Pattern=`^(0(\.[0-9]+)?|1(\.0+)?)$`
	SamplingRate string `json:"samplingRate,omitempty"`

	// Configuration parameters
	Config map[string]string `json:"config,omitempty"`
}

type KubernetesBackupConfig struct {
	// Enable cluster backup
	Enabled bool `json:"enabled"`

	// Backup provider (velero, kasten, portworx)
	Provider string `json:"provider,omitempty"`

	// Backup schedule (cron format)
	Schedule string `json:"schedule,omitempty"`

	// Backup retention period in days
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=365
	RetentionDays int `json:"retentionDays,omitempty"`

	// Storage location for backups
	StorageLocation string `json:"storageLocation,omitempty"`

	// Configuration parameters
	Config map[string]string `json:"config,omitempty"`
}

type MachineProviderReference struct {
	// Name of the machine provider
	Name string `json:"name"`

	// Namespace of the machine provider
	Namespace string `json:"namespace,omitempty"`
}

type KubernetesAuthConfig struct {
	// Authentication providers
	Providers []AuthProvider `json:"providers,omitempty"`

	// OIDC configuration
	OIDC *OIDCConfig `json:"oidc,omitempty"`

	// LDAP configuration
	LDAP *LDAPConfig `json:"ldap,omitempty"`

	// Service account configuration
	ServiceAccount *ServiceAccountAuthConfig `json:"serviceAccount,omitempty"`
}

type AuthProvider struct {
	// Provider name
	Name string `json:"name"`

	// Provider type (oidc, ldap, saml, github, google)
	Type string `json:"type"`

	// Configuration parameters
	Config map[string]string `json:"config,omitempty"`
}

type OIDCConfig struct {
	// OIDC issuer URL
	IssuerURL string `json:"issuerURL"`

	// OIDC client ID
	ClientID string `json:"clientID"`

	// OIDC client secret reference
	ClientSecretRef *CredentialsReference `json:"clientSecretRef,omitempty"`

	// Username claim
	UsernameClaim string `json:"usernameClaim,omitempty"`

	// Groups claim
	GroupsClaim string `json:"groupsClaim,omitempty"`
}

type LDAPConfig struct {
	// LDAP server URL
	ServerURL string `json:"serverURL"`

	// Bind DN
	BindDN string `json:"bindDN"`

	// Bind password reference
	BindPasswordRef *CredentialsReference `json:"bindPasswordRef,omitempty"`

	// User search base
	UserSearchBase string `json:"userSearchBase"`

	// Group search base
	GroupSearchBase string `json:"groupSearchBase"`
}

type ServiceAccountAuthConfig struct {
	// Default service account name
	DefaultName string `json:"defaultName,omitempty"`

	// Service account signing key
	SigningKeyRef *CredentialsReference `json:"signingKeyRef,omitempty"`
}

// KubernetesProviderStatus defines the observed state of KubernetesProvider
type KubernetesProviderStatus struct {
	// Current phase of the provider (Pending, Ready, Failed, Updating)
	Phase string `json:"phase,omitempty"`

	// Human-readable status message
	Message string `json:"message,omitempty"`

	// Kubernetes cluster information
	Cluster KubernetesProviderClusterStatus `json:"cluster,omitempty"`

	// Node pools status
	NodePools []NodePoolStatus `json:"nodePools,omitempty"`

	// Total node count
	NodeCount int `json:"nodeCount,omitempty"`

	// Ready node count
	ReadyNodeCount int `json:"readyNodeCount,omitempty"`

	// Cluster version information
	Version KubernetesVersionStatus `json:"version,omitempty"`

	// Addon status
	Addons []AddonStatus `json:"addons,omitempty"`

	// Capacity and resource usage
	Capacity KubernetesCapacityStatus `json:"capacity,omitempty"`

	// Health status
	Health KubernetesHealthStatus `json:"health,omitempty"`

	// Security status
	Security KubernetesSecurityStatus `json:"security,omitempty"`

	// Endpoint information
	Endpoints KubernetesEndpointStatus `json:"endpoints,omitempty"`

	// Last update time
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`

	// Conditions represent the latest available observations
	Conditions []KubernetesProviderCondition `json:"conditions,omitempty"`
}

type KubernetesProviderClusterStatus struct {
	// Cluster ID
	ID string `json:"id,omitempty"`

	// Cluster state (creating, active, updating, deleting, failed)
	State string `json:"state,omitempty"`

	// Cluster endpoint
	Endpoint string `json:"endpoint,omitempty"`

	// Certificate authority data
	CertificateAuthorityData string `json:"certificateAuthorityData,omitempty"`

	// OIDC issuer URL
	OIDCIssuerURL string `json:"oidcIssuerURL,omitempty"`

	// Platform version
	PlatformVersion string `json:"platformVersion,omitempty"`

	// Creation timestamp
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`
}

type NodePoolStatus struct {
	// Node pool name
	Name string `json:"name"`

	// Current state (creating, active, updating, deleting, failed)
	State string `json:"state"`

	// Current node count
	CurrentNodes int `json:"currentNodes"`

	// Ready node count
	ReadyNodes int `json:"readyNodes"`

	// Desired node count
	DesiredNodes int `json:"desiredNodes"`

	// Machine class
	MachineClass string `json:"machineClass"`

	// Auto-scaling status
	AutoScaling *NodePoolAutoScalingStatus `json:"autoScaling,omitempty"`

	// Health status
	Health string `json:"health,omitempty"`

	// Last update time
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`
}

type NodePoolAutoScalingStatus struct {
	// Is auto-scaling enabled
	Enabled bool `json:"enabled"`

	// Current minimum nodes
	MinNodes int `json:"minNodes"`

	// Current maximum nodes
	MaxNodes int `json:"maxNodes"`

	// Last scaling action
	LastScalingAction string `json:"lastScalingAction,omitempty"`

	// Last scaling time
	LastScalingTime *metav1.Time `json:"lastScalingTime,omitempty"`
}

type KubernetesVersionStatus struct {
	// Current Kubernetes version
	Current string `json:"current,omitempty"`

	// Available upgrade versions
	AvailableUpgrades []string `json:"availableUpgrades,omitempty"`

	// Platform version
	Platform string `json:"platform,omitempty"`

	// Control plane version
	ControlPlane string `json:"controlPlane,omitempty"`

	// Node version
	Node string `json:"node,omitempty"`
}

type AddonStatus struct {
	// Addon name
	Name string `json:"name"`

	// Addon version
	Version string `json:"version,omitempty"`

	// Status (active, inactive, failed, updating)
	Status string `json:"status"`

	// Health status
	Health string `json:"health,omitempty"`

	// Configuration status
	ConfigurationStatus string `json:"configurationStatus,omitempty"`

	// Last update time
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`
}

type KubernetesCapacityStatus struct {
	// Total CPU capacity (cores)
	CPU string `json:"cpu,omitempty"`

	// Total memory capacity
	Memory string `json:"memory,omitempty"`

	// Total storage capacity
	Storage string `json:"storage,omitempty"`

	// Total pods capacity
	Pods string `json:"pods,omitempty"`

	// CPU usage
	CPUUsage string `json:"cpuUsage,omitempty"`

	// Memory usage
	MemoryUsage string `json:"memoryUsage,omitempty"`

	// Storage usage
	StorageUsage string `json:"storageUsage,omitempty"`

	// Pods usage
	PodsUsage string `json:"podsUsage,omitempty"`
}

type KubernetesHealthStatus struct {
	// Overall health status (Healthy, Degraded, Unhealthy)
	Overall string `json:"overall,omitempty"`

	// API server health
	APIServer string `json:"apiServer,omitempty"`

	// ETCD health
	ETCD string `json:"etcd,omitempty"`

	// Controller manager health
	ControllerManager string `json:"controllerManager,omitempty"`

	// Scheduler health
	Scheduler string `json:"scheduler,omitempty"`

	// CoreDNS health
	CoreDNS string `json:"coreDNS,omitempty"`

	// Node health summary
	Nodes string `json:"nodes,omitempty"`

	// Last health check
	LastCheck *metav1.Time `json:"lastCheck,omitempty"`
}

type KubernetesSecurityStatus struct {
	// Pod Security Standards status
	PodSecurityStandards string `json:"podSecurityStandards,omitempty"`

	// Network policies status
	NetworkPolicies string `json:"networkPolicies,omitempty"`

	// RBAC status
	RBAC string `json:"rbac,omitempty"`

	// Image scanning status
	ImageScanning string `json:"imageScanning,omitempty"`

	// Runtime security status
	RuntimeSecurity string `json:"runtimeSecurity,omitempty"`

	// Security scan results
	SecurityScanResults map[string]string `json:"securityScanResults,omitempty"`

	// Last security scan
	LastSecurityScan *metav1.Time `json:"lastSecurityScan,omitempty"`
}

type KubernetesEndpointStatus struct {
	// API server endpoint
	APIServer string `json:"apiServer,omitempty"`

	// Ingress endpoints
	Ingress []string `json:"ingress,omitempty"`

	// Load balancer endpoints
	LoadBalancer []string `json:"loadBalancer,omitempty"`

	// Monitoring endpoints
	Monitoring map[string]string `json:"monitoring,omitempty"`
}

type KubernetesProviderCondition struct {
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

// Common Kubernetes provider phases
const (
	KubernetesProviderPhasePending  = "Pending"
	KubernetesProviderPhaseReady    = "Ready"
	KubernetesProviderPhaseFailed   = "Failed"
	KubernetesProviderPhaseUpdating = "Updating"
)

// Common Kubernetes provider condition types
const (
	KubernetesProviderConditionReady          = "Ready"
	KubernetesProviderConditionHealthy        = "Healthy"
	KubernetesProviderConditionAPIServerReady = "APIServerReady"
	KubernetesProviderConditionNodesReady     = "NodesReady"
	KubernetesProviderConditionAddonsReady    = "AddonsReady"
	KubernetesProviderConditionNetworkReady   = "NetworkReady"
	KubernetesProviderConditionSecurityReady  = "SecurityReady"
)

// Cluster states
const (
	KubernetesClusterStateCreating = "creating"
	KubernetesClusterStateActive   = "active"
	KubernetesClusterStateUpdating = "updating"
	KubernetesClusterStateDeleting = "deleting"
	KubernetesClusterStateFailed   = "failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// KubernetesProviderList contains a list of KubernetesProvider
type KubernetesProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubernetesProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubernetesProvider{}, &KubernetesProviderList{})
}

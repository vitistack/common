package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Machine is the Schema for the Machines API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=machines,scope=Namespaced,shortName=m
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="IpAddresses",type=string,JSONPath=`.status.ipAddresses`
// +kubebuilder:printcolumn:name="Provider",type=string,JSONPath=`.spec.provider`
// +kubebuilder:printcolumn:name="Instance Type",type=string,JSONPath=`.spec.instanceType`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Machine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineSpec   `json:"spec,omitempty"`
	Status MachineStatus `json:"status,omitempty"`
}

type MachineSpec struct {
	// The name of the machine
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name,omitempty"`

	// The instance type/size of the machine (e.g., t3.medium, Standard_B2s, n1-standard-2)
	// +kubebuilder:validation:MinLength=1
	InstanceType string `json:"instanceType,omitempty"`

	// The provider-specific machine type override
	MachineType string `json:"machineType,omitempty"`

	// CPU configuration
	CPU MachineCPU `json:"cpu,omitempty"`

	// Memory configuration in bytes
	// +kubebuilder:validation:Minimum=0
	Memory int64 `json:"memory,omitempty"`

	// Disk configuration
	Disks []MachineSpecDisk `json:"disks,omitempty"`

	// Network configuration
	Network MachineNetwork `json:"network,omitempty"`

	// Operating system configuration
	OS MachineOS `json:"os,omitempty"`

	// Machine Provider
	// +kubebuilder:validation:Required
	Provider MachineProviderType `json:"provider,omitempty"`

	// Cloud provider configuration
	ProviderConfig CloudProviderConfig `json:"providerConfig,omitempty"`

	// SSH key configuration
	SSHKeys []string `json:"sshKeys,omitempty"`

	// User data script to run on first boot
	UserData string `json:"userData,omitempty"`

	// Tags/labels to apply to the machine
	Tags map[string]string `json:"tags,omitempty"`

	// Security groups or firewall rules
	SecurityGroups []string `json:"securityGroups,omitempty"`

	// Whether to enable monitoring
	Monitoring bool `json:"monitoring,omitempty"`

	// Backup configuration
	Backup MachineBackup `json:"backup,omitempty"`
}

type MachineCPU struct {
	// Number of CPU cores
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=256
	Cores int `json:"cores,omitempty"`
	// Number of threads per core
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=8
	ThreadsPerCore int `json:"threadsPerCore,omitempty"`
	// Number of CPU sockets
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=16
	Sockets int `json:"sockets,omitempty"`
}

type MachineSpecDisk struct {
	// Name of the disk
	Name string `json:"name,omitempty"`
	// Size of the disk in GB
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65536
	SizeGB int64 `json:"sizeGB,omitempty"`
	// Type of the disk (e.g., gp2, gp3, pd-ssd, Premium_LRS)
	Type string `json:"type,omitempty"`
	// Whether this is the boot disk
	Boot bool `json:"boot,omitempty"`
	// Device name (e.g., /dev/sda, /dev/nvme0n1)
	Device string `json:"device,omitempty"`
	// Encryption settings
	Encrypted bool `json:"encrypted,omitempty"`
	// IOPS for the disk (if supported by provider)
	// +kubebuilder:validation:Minimum=100
	// +kubebuilder:validation:Maximum=64000
	IOPS int `json:"iops,omitempty"`
	// Throughput in MB/s (if supported by provider)
	// +kubebuilder:validation:Minimum=125
	// +kubebuilder:validation:Maximum=4000
	Throughput int `json:"throughput,omitempty"`
}

type MachineNetwork struct {
	// VPC/Virtual Network ID
	VPC string `json:"vpc,omitempty"`
	// Subnet ID
	Subnet string `json:"subnet,omitempty"`
	// Whether to assign a public IP
	AssignPublicIP bool `json:"assignPublicIP,omitempty"`
	// Static private IP address
	PrivateIP string `json:"privateIP,omitempty"`
	// Static public IP address or Elastic IP
	PublicIP string `json:"publicIP,omitempty"`
	// Network interfaces
	Interfaces []NetworkInterface `json:"interfaces,omitempty"`
}

type NetworkInterface struct {
	// Name of the network interface
	Name string `json:"name,omitempty"`
	// Subnet for this interface
	Subnet string `json:"subnet,omitempty"`
	// Security groups for this interface
	SecurityGroups []string `json:"securityGroups,omitempty"`
	// Whether this is the primary interface
	Primary bool `json:"primary,omitempty"`
}

type MachineOS struct {
	// Operating system family (linux, windows)
	Family string `json:"family,omitempty"`
	// Distribution (ubuntu, centos, rhel, windows-server, debian, alpine)
	Distribution string `json:"distribution,omitempty"`
	// Version of the OS
	Version string `json:"version,omitempty"`
	// Architecture (amd64, arm64)
	// +kubebuilder:validation:Enum=amd64;arm64;x86_64
	Architecture string `json:"architecture,omitempty"`
	// Image ID/AMI/Template ID
	ImageID string `json:"imageID,omitempty"`
	// ISO URI for custom installations
	ISOUri string `json:"isoUri,omitempty"`
	// Image family or marketplace image
	ImageFamily string `json:"imageFamily,omitempty"`
}

type CloudProviderConfig struct {
	// Provider name (aws, azure, gcp, vsphere, openstack)
	Name string `json:"name,omitempty"`
	// Region where the machine should be created
	Region string `json:"region,omitempty"`
	// Availability zone
	Zone string `json:"zone,omitempty"`
	// Provider-specific configuration
	Config map[string]string `json:"config,omitempty"`
	// Credentials reference
	CredentialsRef *CredentialsReference `json:"credentialsRef,omitempty"`
}

type CredentialsReference struct {
	// Name of the secret containing credentials
	SecretName string `json:"secretName,omitempty"`
	// Namespace of the secret (defaults to machine namespace)
	Namespace string `json:"namespace,omitempty"`
}

type MachineBackup struct {
	// Whether to enable automated backups
	Enabled bool `json:"enabled,omitempty"`
	// Backup schedule (cron format)
	Schedule string `json:"schedule,omitempty"`
	// Retention period in days
	RetentionDays int `json:"retentionDays,omitempty"`
}

type MachineStatus struct {
	// Current phase of the machine (Pending, Creating, Running, Stopping, Stopped, Terminating, Terminated, Failed)
	Phase string `json:"phase,omitempty"`

	// Detailed status message
	Message string `json:"message,omitempty"`

	// The unique identifier assigned by the provider
	ProviderID string `json:"providerID,omitempty"`

	// Internal machine identifier
	MachineID string `json:"machineID,omitempty"`

	// The current state of the machine
	State string `json:"state,omitempty"`

	// The last time the machine status was updated
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`

	// The provider that created this machine
	Provider MachineProviderType `json:"provider,omitempty"`

	// The region where the machine is located
	Region string `json:"region,omitempty"`

	// The zone where the machine is located
	Zone string `json:"zone,omitempty"`

	// The IP addresses of the machine
	IPAddresses []string `json:"ipAddresses,omitempty"`

	// IPv6 addresses of the machine
	IPv6Addresses []string `json:"ipv6Addresses,omitempty"`

	// Public IP addresses
	PublicIPAddresses []string `json:"publicIPAddresses,omitempty"`

	// Private IP addresses
	PrivateIPAddresses []string `json:"privateIPAddresses,omitempty"`

	// The machine's hostname
	Hostname string `json:"hostname,omitempty"`

	// The machine's CPU architecture
	Architecture string `json:"architecture,omitempty"`

	// The machine's operating system
	OperatingSystem string `json:"operatingSystem,omitempty"`

	// The machine's operating system version
	OperatingSystemVersion string `json:"operatingSystemVersion,omitempty"`

	// The machine's kernel version
	KernelVersion string `json:"kernelVersion,omitempty"`

	// Actual CPU count
	CPUs int `json:"cpus,omitempty"`

	// Actual memory in bytes
	Memory int64 `json:"memory,omitempty"`

	// Actual disk information
	Disks []MachineStatusDisk `json:"disks,omitempty"`

	// Network interface information
	NetworkInterfaces []NetworkInterfaceStatus `json:"networkInterfaces,omitempty"`

	// Conditions represent the latest available observations of the machine's state
	Conditions []MachineCondition `json:"conditions,omitempty"`

	// Boot time of the machine
	BootTime *metav1.Time `json:"bootTime,omitempty"`

	// Creation time of the machine
	CreationTime *metav1.Time `json:"creationTime,omitempty"`

	// Failure reason if the machine failed to be created
	FailureReason *string `json:"failureReason,omitempty"`

	// Failure message if the machine failed to be created
	FailureMessage *string `json:"failureMessage,omitempty"`
}

type MachineDisk struct {
	// The disk's name
	Name string `json:"name"`
	// The disk's size in bytes
	Size int64 `json:"size"`
	// The disk's type (e.g., SSD, HDD)
	Type string `json:"type"`
	// The disk's mount point
	MountPoint string `json:"mountPoint"`
	// The disk's filesystem type (e.g., ext4, xfs)
	FilesystemType string `json:"filesystemType"`
	// The disk's UUID
	UUID string `json:"uuid"`
	// The disk's label
	Label string `json:"label"`
	// The disk's serial number
	SerialNumber string `json:"serialNumber"`
}

type MachineStatusDisk struct {
	// The disk's name
	Name string `json:"name,omitempty"`
	// The disk's size in bytes
	Size int64 `json:"size,omitempty"`
	// The disk's type (e.g., SSD, HDD, gp2, gp3)
	Type string `json:"type,omitempty"`
	// The disk's mount point
	MountPoint string `json:"mountPoint,omitempty"`
	// PVC name
	PVCName string `json:"pvcName,omitempty"`
	// Volume mode, filesystem or block
	VolumeMode string `json:"volumeMode,omitempty"`
	// Access modes, readwriteonce, readwritemany, readonlymany
	AccessModes []string `json:"accessModes,omitempty"`
	// The disk's filesystem type (e.g., ext4, xfs)
	FilesystemType string `json:"filesystemType,omitempty"`
	// The disk's UUID
	UUID string `json:"uuid,omitempty"`
	// The disk's label
	Label string `json:"label,omitempty"`
	// The disk's serial number
	SerialNumber string `json:"serialNumber,omitempty"`
	// Device path (e.g., /dev/sda)
	Device string `json:"device,omitempty"`
	// Used space in bytes
	UsedBytes int64 `json:"usedBytes,omitempty"`
	// Available space in bytes
	AvailableBytes int64 `json:"availableBytes,omitempty"`
	// Usage percentage as string (e.g., "75.5%")
	UsagePercent string `json:"usagePercent,omitempty"`
}

type NetworkInterfaceStatus struct {
	// Name of the network interface
	Name string `json:"name,omitempty"`
	// MAC address
	MACAddress string `json:"macAddress,omitempty"`
	// IP addresses assigned to this interface
	IPAddresses []string `json:"ipAddresses,omitempty"`
	// IPv6 addresses assigned to this interface
	IPv6Addresses []string `json:"ipv6Addresses,omitempty"`
	// Interface state (up, down)
	State string `json:"state,omitempty"`
	// MTU size
	MTU int `json:"mtu,omitempty"`
	// Interface type (ethernet, wifi, etc.)
	Type string `json:"type,omitempty"`
}

type MachineCondition struct {
	// Type of condition
	Type string `json:"type,omitempty"`
	// Status of the condition (True, False, Unknown)
	Status string `json:"status,omitempty"`
	// Last time the condition transitioned from one status to another
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition
	Message string `json:"message,omitempty"`
}

// Common machine phases
const (
	MachinePhasePending     = "Pending"
	MachinePhaseCreating    = "Creating"
	MachinePhaseRunning     = "Running"
	MachinePhaseStopping    = "Stopping"
	MachinePhaseStopped     = "Stopped"
	MachinePhaseTerminating = "Terminating"
	MachinePhaseTerminated  = "Terminated"
	MachinePhaseFailed      = "Failed"
)

// Common machine condition types
const (
	MachineConditionReady               = "Ready"
	MachineConditionNetworkReady        = "NetworkReady"
	MachineConditionBootstrapReady      = "BootstrapReady"
	MachineConditionInfrastructureReady = "InfrastructureReady"
	MachineConditionDrainReady          = "DrainReady"
	MachineConditionBackupReady         = "BackupReady"
)

// Common condition statuses
const (
	ConditionTrue    = "True"
	ConditionFalse   = "False"
	ConditionUnknown = "Unknown"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// MachineList contains a list of Machine
type MachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Machine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Machine{}, &MachineList{})
}

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EtcdBackup is the Schema for the EtcdBackup API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=etcdbackups,scope=Namespaced,shortName=eb
// +kubebuilder:printcolumn:name="Cluster",type=string,JSONPath=`.spec.clusterName`,description="Target cluster name"
// +kubebuilder:printcolumn:name="Storage",type=string,JSONPath=`.spec.storageLocation.type`,description="Storage type"
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`,description="Current phase"
// +kubebuilder:printcolumn:name="Last Backup",type=date,JSONPath=`.status.lastBackupTime`,description="Last successful backup"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type EtcdBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EtcdBackupSpec   `json:"spec,omitempty"`
	Status EtcdBackupStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// EtcdBackupList contains a list of EtcdBackup
type EtcdBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EtcdBackup `json:"items"`
}

// EtcdBackupSpec defines the desired state of an etcd backup
type EtcdBackupSpec struct {
	// ClusterName is the name of the Kubernetes cluster to backup
	// +kubebuilder:validation:Required
	ClusterName string `json:"clusterName"`

	// Schedule is the cron schedule for automated backups (e.g., "0 */6 * * *" for every 6 hours)
	// +kubebuilder:validation:Optional
	Schedule string `json:"schedule,omitempty"`

	// Retention specifies how many backups to retain
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=7
	Retention int `json:"retention,omitempty"`

	// StorageLocation specifies where to store the backup
	// +kubebuilder:validation:Required
	StorageLocation EtcdBackupStorageLocation `json:"storageLocation"`
}

// EtcdBackupStorageLocation defines the storage destination for backups
type EtcdBackupStorageLocation struct {
	// Type is the storage type (e.g., "s3", "gcs", "azure", "local")
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=s3;gcs;azure;local
	Type string `json:"type"`

	// Bucket is the bucket name for cloud storage
	// +kubebuilder:validation:Optional
	Bucket string `json:"bucket,omitempty"`

	// Path is the path/prefix within the storage location
	// +kubebuilder:validation:Optional
	Path string `json:"path,omitempty"`

	// SecretRef references a secret containing storage credentials
	// +kubebuilder:validation:Optional
	SecretRef string `json:"secretRef,omitempty"`
}

// EtcdBackupStatus defines the observed state of an etcd backup
type EtcdBackupStatus struct {
	// Phase represents the current phase of the backup (Pending, Running, Completed, Failed)
	// +kubebuilder:validation:Enum=Pending;Running;Completed;Failed
	Phase string `json:"phase,omitempty"`

	// Message provides additional information about the current status
	Message string `json:"message,omitempty"`

	// LastBackupTime is the timestamp of the last successful backup
	LastBackupTime *metav1.Time `json:"lastBackupTime,omitempty"`

	// NextBackupTime is the scheduled time for the next backup (if scheduled)
	NextBackupTime *metav1.Time `json:"nextBackupTime,omitempty"`

	// BackupSize is the size of the last backup in bytes
	BackupSize string `json:"backupSize,omitempty"`

	// BackupCount is the current number of stored backups
	BackupCount int `json:"backupCount,omitempty"`

	// Conditions represent the latest available observations of the backup's state
	Conditions []EtcdBackupCondition `json:"conditions,omitempty"`
}

// EtcdBackupCondition describes the state of an etcd backup at a certain point
type EtcdBackupCondition struct {
	// Type is the type of the condition
	Type string `json:"type"`

	// Status is the status of the condition (True, False, Unknown)
	Status string `json:"status"`

	// LastTransitionTime is the last time the condition transitioned
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason is a brief reason for the condition's last transition
	Reason string `json:"reason,omitempty"`

	// Message is a human-readable message indicating details about the transition
	Message string `json:"message,omitempty"`
}

func init() {
	SchemeBuilder.Register(&EtcdBackup{}, &EtcdBackupList{})
}

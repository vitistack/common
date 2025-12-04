package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineClass is the Schema for the MachineClass API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=machineclasses,scope=Cluster,shortName=mc
// +kubebuilder:printcolumn:name="Name",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="Provider",type=string,JSONPath=`.spec.provider`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="Created",type=string,JSONPath=`.status.created`,description="Creation Timestamp"
type MachineClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineClassSpec   `json:"spec,omitempty"`
	Status MachineClassStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// MachineClassList contains a list of MachineClass
type MachineClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachineClass `json:"items"`
}

type MachineClassSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:default:=true
	Enabled bool `json:"enabled,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:default:=false
	Default bool `json:"default,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:default:="Standard"
	Category string `json:"category,omitempty"`

	// +kubebuilder:validation:Required
	Memory MachineClassMemorySpec `json:"memory,omitempty"`

	// +kubebuilder:validation:Required
	CPU MachineClassCPUSpec `json:"cpu,omitempty"`

	// +kubebuilder:validation:Required
	MachineProviders []MachineProviderType `json:"machineProviders,omitempty"`

	// +kubebuilder:validation:Optional
	GPU MachineClassGPUSpec `json:"gpu,omitempty"`

	// +kubebuilder:validation:Optional
	Description string `json:"description,omitempty"`

	// +kubebuilder:validation:Optional
	DisplayName string `json:"displayName,omitempty"`
}

type MachineClassCPUSpec struct {
	// +kubebuilder:validation:Required
	Cores uint `json:"cores,omitempty"`

	// +kubebuilder:validation:Optional
	Sockets uint `json:"sockets,omitempty"`

	// +kubebuilder:validation:Optional
	Threads uint `json:"threads,omitempty"`
}

type MachineClassMemorySpec struct {
	// +kubebuilder:validation:Required
	Quantity resource.Quantity `json:"quantity,omitempty"`

	// +kubebuilder:validation:Optional
	MinQuantity resource.Quantity `json:"minQuantity,omitempty"`

	// +kubebuilder:validation:Optional
	MaxQuantity resource.Quantity `json:"maxQuantity,omitempty"`
}

type MachineClassGPUSpec struct {
	// +kubebuilder:validation:Required
	Cores uint `json:"cores,omitempty"`

	// +kubebuilder:validation:Optional
	Manufacturer string `json:"manufacturer,omitempty"`
}

type MachineClassStatus struct {
	Name    string `json:"name,omitempty"`
	Phase   string `json:"phase,omitempty"`
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Created string `json:"created,omitempty"`
}

func init() {
	SchemeBuilder.Register(&MachineClass{}, &MachineClassList{})
}

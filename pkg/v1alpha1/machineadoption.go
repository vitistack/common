package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=machineadoptions,scope=Namespaced,shortName=madopt
// +kubebuilder:printcolumn:name="Provider",type=string,JSONPath=`.spec.providerName`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Machine",type=string,JSONPath=`.status.machineName`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type MachineAdoption struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineAdoptionSpec   `json:"spec,omitempty"`
	Status MachineAdoptionStatus `json:"status,omitempty"`
}

type MachineAdoptionSpec struct {
	// MachineProvider used to locate and manage the existing machine.
	// +kubebuilder:validation:MinLength=1
	ProviderName string `json:"providerName"`

	// Stable provider-specific identifier.
	// For vSphere, this is the VM instance UUID.
	// +kubebuilder:validation:MinLength=1
	ProviderMachineID string `json:"providerMachineID"`

	// Optional name for the generated Machine resource.
	// The provider derives it from the existing machine when omitted.
	MachineName string `json:"machineName,omitempty"`
}

type MachineAdoptionStatus struct {
	// +kubebuilder:validation:Enum=Pending;Processing;Completed;Failed
	Phase string `json:"phase,omitempty"`

	// Name of the generated Machine resource.
	MachineName string `json:"machineName,omitempty"`

	// Namespace of the generated Machine resource.
	MachineNamespace string `json:"machineNamespace,omitempty"`

	Message string `json:"message,omitempty"`

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions represent the latest adoption state.
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
type MachineAdoptionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []MachineAdoption `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MachineAdoption{}, &MachineAdoptionList{})
}

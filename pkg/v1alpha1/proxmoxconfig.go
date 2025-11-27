package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProxmoxConfig is the Schema for the ProxmoxConfig API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=proxmoxconfigs,scope=Cluster,shortName=pxc
// +kubebuilder:printcolumn:name="Name",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="Endpoint",type=string,JSONPath=`.spec.endpoint`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="Created",type=string,JSONPath=`.status.created`,description="Creation Timestamp"
type ProxmoxConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ProxmoxConfigSpec `json:"spec,omitempty"`

	Status ProxmoxConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ProxmoxConfigList contains a list of ProxmoxConfig
type ProxmoxConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProxmoxConfig `json:"items"`
}

type ProxmoxConfigSpec struct {
	// +kubebuilder:validation:Required
	Endpoint string `json:"endpoint,omitempty"`

	// +kubebuilder:validation:Required
	Port string `json:"port,omitempty"`

	// +kubebuilder:validation:Required
	Username string `json:"username,omitempty"`

	// +kubebuilder:validation:Required
	Token string `json:"token,omitempty"`

	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`
}

type ProxmoxConfigStatus struct {
	Phase   string `json:"phase,omitempty"`
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Created string `json:"created,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ProxmoxConfig{}, &ProxmoxConfigList{})
}

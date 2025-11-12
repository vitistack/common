package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkConfiguration is the Schema for the NetworkConfiguration API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=networkconfigurations,scope=Namespaced,shortName=nc
// +kubebuilder:printcolumn:name="Name",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="Created",type=string,JSONPath=`.status.created`,description="Creation Timestamp"
// +kubebuilder:printcolumn:name="Interfaces",type=string,JSONPath=`.status.networkInterfaces.count()`,description="Count Network Interfaces"
type NetworkConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworkConfigurationSpec   `json:"spec,omitempty"`
	Status NetworkConfigurationStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// NetworkConfigurationList contains a list of NetworkConfiguration
type NetworkConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkConfiguration `json:"items"`
}

type NetworkConfigurationSpec struct {
	// Name of the NetworkConfiguration
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=2
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:Pattern=`^[A-Za-z0-9_-]+$`
	Name string `json:"name"`

	// Description of the NetworkConfiguration
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MaxLength=256
	Description string `json:"description,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=2
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:Pattern=`^[A-Za-z0-9_-]+$`
	DatacenterIdentifier string `json:"datacenterIdentifier,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=2
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:Pattern=`^[A-Za-z0-9_-]+$`
	SupervisorIdentifier string `json:"supervisorIdentifier,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=2
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:Pattern=`^[A-Za-z0-9_-]+$`
	ClusterIdentifier string `json:"clusterIdentifier,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:Pattern=`^[A-Za-z0-9_-]+$`
	Provider string `json:"provider,omitempty"`

	NetworkInterfaces []NetworkConfigurationInterface `json:"networkInterfaces,omitempty"`
}

type NetworkConfigurationStatus struct {
	Conditions        []metav1.Condition              `json:"conditions,omitempty"`
	Phase             string                          `json:"phase,omitempty"`
	Status            string                          `json:"status,omitempty"`
	Message           string                          `json:"message,omitempty"`
	Created           metav1.Time                     `json:"created,omitempty"`
	NetworkInterfaces []NetworkConfigurationInterface `json:"networkInterfaces,omitempty"`
}

type NetworkConfigurationInterface struct {
	// Name
	Name string `json:"name,omitempty"`
	// Mac address
	MacAddress string `json:"macAddress,omitempty"`
	// IPv4 addresses
	IPv4Addresses []string `json:"ipv4Addresses,omitempty"`
	// IPv6 addresses
	IPv6Addresses []string `json:"ipv6Addresses,omitempty"`
	// Vlan
	Vlan string `json:"vlan,omitempty"`
	// Subnet
	IPv4Subnet string `json:"ipv4Subnet,omitempty"`
	// Subnet for ipv6
	IPv6Subnet string `json:"ipv6Subnet,omitempty"`
	// Gateway
	IPv4Gateway string `json:"ipv4Gateway,omitempty"`
	// Gateway for ipv6
	IPv6Gateway string `json:"ipv6Gateway,omitempty"`
	// DNS
	DNS []string `json:"dns,omitempty"`
	// DHCP reserved in dchp server(s)
	DHCPReserved bool `json:"dhcpReserved,omitempty"`
}

func init() {
	SchemeBuilder.Register(&NetworkConfiguration{}, &NetworkConfigurationList{})
}

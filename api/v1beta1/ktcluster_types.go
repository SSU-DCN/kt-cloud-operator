/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KTClusterSpec defines the desired state of KTCluster.
type KTClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	APIServerLoadBalancer             APIServerLoadBalancer `json:"apiServerLoadBalancer,omitempty"`
	ControlPlaneExternalNetworkEnable bool                  `json:"controlPlaneExternalNetworkEnable,omitempty"`
	IdentityRef                       IdentityRef           `json:"identityRef,omitempty"`
	ManagedSecurityGroups             ManagedSecurityGroups `json:"managedSecurityGroups,omitempty"`
	ManagedSubnets                    []ManagedSubnet       `json:"managedSubnets,omitempty"`
}

// APIServerLoadBalancer represents the API server load balancer settings
type APIServerLoadBalancer struct {
	Enabled bool `json:"enabled"`
}

// ExternalNetwork represents the external network configuration
// type ExternalNetwork struct {
// 	ID string `json:"id"`
// }

// IdentityRef holds the identity reference for OpenStack
type IdentityRef struct {
	CloudName string `json:"cloudName,omitempty"`
	Name      string `json:"name,omitempty"`
}

// ManagedSecurityGroups contains security group rules for nodes
type ManagedSecurityGroups struct {
	AllNodesSecurityGroupRules []SecurityGroupRule `json:"allNodesSecurityGroupRules,omitempty"`
}

// SecurityGroupRule represents individual security group rules
type SecurityGroupRule struct {
	Description         string   `json:"description,omitempty"`
	Direction           string   `json:"direction,omitempty"`
	EtherType           string   `json:"etherType,omitempty"`
	Name                string   `json:"name,omitempty"`
	PortRangeMin        int      `json:"portRangeMin,omitempty"`
	PortRangeMax        int      `json:"portRangeMax,omitempty"`
	Protocol            string   `json:"protocol,omitempty"`
	RemoteManagedGroups []string `json:"remoteManagedGroups,omitempty"`
}

// ManagedSubnet defines a subnet with CIDR and DNS settings
type ManagedSubnet struct {
	CIDR           string   `json:"cidr,omitempty"`
	DNSNameServers []string `json:"dnsNameservers,omitempty"`
}

// KTClusterStatus defines the observed state of KTCluster.
type KTClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// KTCluster is the Schema for the ktclusters API.
type KTCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KTClusterSpec   `json:"spec,omitempty"`
	Status KTClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KTClusterList contains a list of KTCluster.
type KTClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KTCluster `json:"items"`
}

func init() {
	objectTypes = append(objectTypes, &KTCluster{}, &KTClusterList{})
}

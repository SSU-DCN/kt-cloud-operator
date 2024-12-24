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

// KTNetworkFirewallSpec defines the desired state of KTNetworkFirewall.
type KTNetworkFirewallSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of KTNetworkFirewall. Edit ktnetworkfirewall_types.go to remove/update
	StartPort    string `json:"startport"`
	Protocol     int    `json:"protocol"`
	VirtualIPID  string `json:"virtualipid"`
	Action       int    `json:"action"`
	SrcNetworkID string `json:"srcnetworkid"`
	DstIP        string `json:"dstip"`
	EndPort      string `json:"endport"`
	DstNetworkID string `json:"dstnetworkid"`
}

// KTNetworkFirewallStatus defines the observed state of KTNetworkFirewall.
type KTNetworkFirewallStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// KTNetworkFirewall is the Schema for the ktnetworkfirewalls API.
type KTNetworkFirewall struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KTNetworkFirewallSpec   `json:"spec,omitempty"`
	Status KTNetworkFirewallStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KTNetworkFirewallList contains a list of KTNetworkFirewall.
type KTNetworkFirewallList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KTNetworkFirewall `json:"items"`
}

func init() {
	objectTypes = append(objectTypes, &KTNetworkFirewall{}, &KTNetworkFirewallList{})
}

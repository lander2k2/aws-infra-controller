/*
Copyright 2019 aws-infra-controller maintainers.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InventorySpec defines the desired state of Inventory
type InventorySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Region            string `json:"region"`
	VpcId             string `json:"vpcId"`
	RouteTableId      string `json:"routeTableId"`
	SubnetId          string `json:"subnetId"`
	InternetGatewayId string `json:"internetGatewayId"`
	SecurityGroupId   string `json:"securityGroupId"`
	BucketId          string `json:"bucketId"`
	IamPolicyId       string `json:"iamPolicyId"`
	IamRoleId         string `json:"iamRoleId"`
	InstanceProfileId string `json:"instanceProfileId"`
	InstanceId        string `json:"instanceId"`
}

// InventoryStatus defines the observed state of Inventory
type InventoryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Inventory is the Schema for the inventories API
// +k8s:openapi-gen=true
type Inventory struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InventorySpec   `json:"spec,omitempty"`
	Status InventoryStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InventoryList contains a list of Inventory
type InventoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Inventory `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Inventory{}, &InventoryList{})
}

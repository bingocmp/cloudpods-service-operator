// Copyright 2020 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GeneralServiceSpec defines the desired state of GeneralServiceSpec
type GeneralServiceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Params map[string]IntOrStringOrYamlStore `json:"params"`

	// +optional
	ReleaseId string `json:"releaseId"`

	// +optional
	ServiceId string `json:"serviceId"`

	ResourceSpecBase `json:",inline"`
}

// GeneralServiceStatus defines the observed state of GeneralService
type GeneralServiceStatus struct {
	// Important: Run "make" to regenerate code after modifying this file
	ResourceStatusBase `json:",inline"`
	// +optional
	ExternalInfo GeneralServiceInfo `json:"externalInfo,omitempty"`
}

type GeneralServiceInfo struct {
	ExternalInfoBase `json:",inline"`
	// +optional
	PrimaryKey string `json:"primaryKey"`
	// +optional
	ExternalId string `json:"externalId,omitempty"`
}

// +kubebuilder:object:root=true

// GeneralService is the Schema for the generalservices API
// +kubebuilder:subresource:status
type GeneralService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GeneralServiceSpec   `json:"spec,omitempty"`
	Status GeneralServiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GeneralServiceList contains a list of GeneralServiceSpec
type GeneralServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GeneralService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GeneralService{}, &GeneralServiceList{})
}

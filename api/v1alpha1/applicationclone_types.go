/*
Copyright 2023.

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

// ApplicationCloneSpec defines the desired state of ApplicationClone
type ApplicationCloneSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// From specifies the Application that would be cloned into the current namespace
	From From `json:"from"`

	// ComponentSources lists the Components that be built from source code
	ComponentSources []ComponentSource `json:"componentSources,omitempty"`
}

// ApplicationCloneStatus defines the observed state of ApplicationClone
type ApplicationCloneStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// List of Resources that were cloned
	Resources             []Resource `json:"resources"`
	Error                 string     `json:"error"`
	LastSuccessfulAttempt string     `json:"lastSuccessfulAttempt"`
	LastAttempt           string     `json:"lastAttempt"`
}

type Resource struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}

type From struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}
type ComponentSource struct {
	Name string `json:"name"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ApplicationClone is the Schema for the applicationclones API
type ApplicationClone struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationCloneSpec   `json:"spec,omitempty"`
	Status ApplicationCloneStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ApplicationCloneList contains a list of ApplicationClone
type ApplicationCloneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApplicationClone `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ApplicationClone{}, &ApplicationCloneList{})
}

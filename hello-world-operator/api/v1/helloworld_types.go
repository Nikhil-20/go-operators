/*
Copyright 2026.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HelloWorldSpec defines the desired state of HelloWorld
type HelloWorldSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// foo is an example field of HelloWorld. Edit helloworld_types.go to remove/update
	// +optional
	//Foo *string `json:"foo,omitempty"`

	// Message is the message to display
	// +kubebuilder:validation:Required
	Message string `json:"message,omitempty"`

	// Count is the number of times to display the message
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	Count int32 `json:"count,omitempty"`
}

// HelloWorldStatus defines the observed state of HelloWorld.
type HelloWorldStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the HelloWorld resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//

	// Phase represents the current phase
	Phase string `json:"phase,omitempty"`

	// ConfigMapCreated indicates if the ConfigMap was created
	ConfigMapCreated bool `json:"configMapCreated,omitempty"`

	// LastUpdated is when the status was last updated
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Message",type="string",JSONPath=".spec.message"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// HelloWorld is the Schema for the helloworlds API
type HelloWorld struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of HelloWorld
	// +required
	Spec HelloWorldSpec `json:"spec"`

	// status defines the observed state of HelloWorld
	// +optional
	Status HelloWorldStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HelloWorldList contains a list of HelloWorld
type HelloWorldList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []HelloWorld `json:"items"`
}

func registerfunc(s *runtime.Scheme) error {
	s.AddKnownTypes(SchemeGroupVersion, &HelloWorld{}, &HelloWorldList{})
	return nil
}

func init() {
	// SchemeBuilder.Register(func(s *runtime.Scheme) error {
	// 	s.AddKnownTypes(SchemeGroupVersion, &HelloWorld{}, &HelloWorldList{})
	// 	return nil
	// })

	SchemeBuilder.Register(registerfunc)
}

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Worker the worker object
type Worker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec WorkerSpec `json:"spec"`
}

// WorkerStatus custom status
type WorkerStatus struct {
	Name string
}

// WorkerSpec the worker spec
type WorkerSpec struct {
	Project     string `json:"project"`
	Concurrency int    `json:"concurrency"`
	CPU         int    `json:"cpu"`
	Memory      int    `json:"memory"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkerList is a list of Worker resources
type WorkerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Worker `json:"items"`
}

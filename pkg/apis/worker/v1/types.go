package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Worker the worker object
type Worker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkerSpec   `json:"spec"`
	Status WorkerStatus `json:"status"`
}

// WorkerStatus custom status
type WorkerStatus struct {
	Status            string
	TargetReplicas    int32             `json:"targetReplicas"`
	AvailableReplicas int32             `json:"availableReplicas"`
	UpdatedReplicas   int32             `json:"updatedReplicas"`
	Conditions        []WorkerCondition `json:"conditions"`
}

type WorkerCondition struct {
	Reason  string
	Message string
	Type    string
	//CreationTimeStamp time.Time
}

/**
  volumes:
  - name: data
    persistentVolumeClaim: data
  volumeMounts:
  - volume: data
    mountPath: /home/data
    subpath: sample
*/

// WorkerSpec the worker spec
type WorkerSpec struct {
	Project   string `json:"project"`
	Image     string `json:"image"`
	Replicas  int32  `json:"replicas"`
	Resources struct {
		Concurrency int  `json:"concurrency"`
		CPU         string  `json:"cpu"`
		Memory      string  `json:"memory"`
		GPU         bool `json:"gpu"`
	} `json:"resources"`
}

// GenerateDeploymentName generates the name for the deployment
// based on the worker
func (w *Worker) GenerateDeploymentName() string {
	return w.Name
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkerList is a list of Worker resources
type WorkerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Worker `json:"items"`
}

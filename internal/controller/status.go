package controller

import workerv1 "github.com/willeslau/k8s-controller/pkg/apis/worker/v1"

func newStatus() *workerv1.WorkerStatus {
	return nil
}

func newCreatedStatus(w *workerv1.Worker) *workerv1.WorkerStatus {
	condition := workerv1.WorkerCondition{
		Reason:  "NewWorkerCreated",
		Message: "New worker created",
		Type:    "Progressing",
	}
	conditions := make([]workerv1.WorkerCondition, 0)
	conditions = append(conditions, condition)
	return &workerv1.WorkerStatus{
		Status: "Created",
		TargetReplicas: w.Spec.Replicas,
		AvailableReplicas: int32(0),
		UpdatedReplicas: int32(0),
		Conditions: conditions,
	}
}

func newCompletedStatus(w *workerv1.Worker) *workerv1.WorkerStatus {
	condition := workerv1.WorkerCondition{
		Reason:  "WorkerAvailable",
		Message: "New worker is available",
		Type:    "Progressing",
	}
	conditions := make([]workerv1.WorkerCondition, 0)
	conditions = append(conditions, condition)
	return &workerv1.WorkerStatus{
		Status: "Created",
		TargetReplicas: w.Spec.Replicas,
		AvailableReplicas: int32(0),
		UpdatedReplicas: int32(0),
		Conditions: conditions,
	}
}

func newCondition() *workerv1.WorkerCondition {
	return nil
}

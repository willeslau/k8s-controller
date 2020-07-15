package controller

import (
	"fmt"
	workerv1 "github.com/willeslau/k8s-controller/pkg/apis/worker/v1"
	appsv1 "k8s.io/api/apps/v1"
)

const (
	PROGRESSING = "progressing"
	AVAILABILITY = "availability"
)

const (
	REQUIRED_REPLICAS_AVAILABLE = "RequiredReplicasAvailable"
	REQUIRED_REPLICAS_UNAVAILABLE = "RequiredReplicasUnavailable"

	NEW_WORKER_CREATED = "NewWorkerCreated"
	NEW_WORKER_AVAILABLE = "NewWorkerAvailable"
	NEW_WORKER_UPDATING = "NewWorkerUpdating"
)

func newCreatedStatus(w *workerv1.Worker) *workerv1.WorkerStatus {
	conditions := make([]workerv1.WorkerCondition, 0)
	return &workerv1.WorkerStatus{
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

// findStatusCondition
func findStatusCondition(conditions *[]workerv1.WorkerCondition, conditionType string) *workerv1.WorkerCondition {
	for _, c := range *conditions {
		if c.Type == conditionType {
			return &c
		}
	}
	return nil
}

// isSameStatusCondition compares if two status conditions are the same
func isSameStatusCondition(refCondition *workerv1.WorkerCondition, tarCondition *workerv1.WorkerCondition) bool {
	if refCondition == nil && tarCondition == nil { return true }
	if (refCondition == nil && tarCondition != nil) ||
		(refCondition != nil && tarCondition == nil) { return false }

	return refCondition.Type == tarCondition.Type &&
		refCondition.Message == tarCondition.Message &&
		refCondition.Reason == tarCondition.Reason
}

// isSameStatus compare the two input statuses
func isSameStatus(refStatus *workerv1.WorkerStatus, tarStatus *workerv1.WorkerStatus) bool {
	if refStatus.UpdatedReplicas != tarStatus.UpdatedReplicas ||
		refStatus.AvailableReplicas != tarStatus.AvailableReplicas ||
		refStatus.TargetReplicas != tarStatus.TargetReplicas {
		return false
	}

	// compare PROGRESSING
	refProgressing := findStatusCondition(&refStatus.Conditions, PROGRESSING)
	tarProgressing := findStatusCondition(&tarStatus.Conditions, PROGRESSING)
	if !isSameStatusCondition(refProgressing, tarProgressing) { return false }

	// compare AVAILABILITY
	refAvailability := findStatusCondition(&refStatus.Conditions, AVAILABILITY)
	tarAvailability := findStatusCondition(&tarStatus.Conditions, AVAILABILITY)
	return isSameStatusCondition(refAvailability, tarAvailability)
}

// updateReplicaCounts updates the replica counts in the status of the worker
func updateReplicaCounts(status *workerv1.WorkerStatus, worker *workerv1.Worker, currentDeployment *appsv1.Deployment) {
	status.TargetReplicas = worker.Spec.Replicas
	status.UpdatedReplicas = currentDeployment.Status.UpdatedReplicas
	status.AvailableReplicas = currentDeployment.Status.AvailableReplicas
}

func assignCondition(status *workerv1.WorkerStatus, condition workerv1.WorkerCondition) *workerv1.WorkerStatus {
	for i := range status.Conditions {
		if status.Conditions[i].Type == condition.Type {
			status.Conditions[i] = condition
			return status
		}
	}
	status.Conditions = append(status.Conditions, condition)
	return status
}

// updateProgressingCondition updates the progress portion of the conditions
func updateProgressingCondition(status *workerv1.WorkerStatus, worker *workerv1.Worker, currentDeployment *appsv1.Deployment) {
	currentReplicas := currentDeployment.Status.ReadyReplicas
	updatedReplicas := currentDeployment.Status.UpdatedReplicas
	requiredReplicas := worker.Spec.Replicas

	condition := workerv1.WorkerCondition{Type: PROGRESSING}
	if requiredReplicas <= currentReplicas {
		condition.Reason = NEW_WORKER_AVAILABLE
		condition.Message = "New worker is available"
	} else if updatedReplicas > 0 {
		condition.Reason = NEW_WORKER_UPDATING
		condition.Message = "New worker is updating"
	} else if updatedReplicas == 0 {
		condition.Reason = NEW_WORKER_CREATED
		condition.Message = "New worker is created"
	}

	status = assignCondition(status, condition)
}

// updateAvailabilityCondition updates the availability portion of the conditions
func updateAvailabilityCondition(status *workerv1.WorkerStatus, worker *workerv1.Worker, currentDeployment *appsv1.Deployment) {
	currentReplicas := currentDeployment.Status.ReadyReplicas
	requiredReplicas := worker.Spec.Replicas

	condition := workerv1.WorkerCondition{Type: AVAILABILITY}
	if currentReplicas >= requiredReplicas {
		condition.Reason = REQUIRED_REPLICAS_AVAILABLE
		condition.Message = fmt.Sprintf("Worker has reached the required replicas %d", requiredReplicas)
	} else {
		condition.Reason = REQUIRED_REPLICAS_UNAVAILABLE
		condition.Message = fmt.Sprintf("Worker has reached the %d replicas, but not the required %d", currentReplicas, requiredReplicas)
	}

	status = assignCondition(status, condition)
}
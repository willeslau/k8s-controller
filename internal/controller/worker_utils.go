package controller

import (
	"context"
	workerv1 "github.com/willeslau/k8s-controller/pkg/apis/worker/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


func (c *Controller) createNewWorker(w *workerv1.Worker) (*appsv1.Deployment, error) {
	newD, err := c.deploymentRobin.createDeploymentFromWorker(w)
	if err != nil {
		return nil, err
	}
	w.Status = *newCreatedStatus(w)
	w, err = c.syncWorkerStatus(w)
	if err != nil {
		return nil, err
	}
	return newD, nil
}

func (c *Controller) completeUpdate(w *workerv1.Worker) (*workerv1.Worker, error) {
	w.Status = *newCompletedStatus(w)
	return c.syncWorkerStatus(w)
}

func filter(d *appsv1.Deployment) bool {
	// not deleted
	return d.DeletionTimestamp == nil
}

func (c *Controller) updateWorker(w *workerv1.Worker, dList []*appsv1.Deployment) ([]*appsv1.Deployment, error) {
	latestD := findLatestDeployment(dList, filter)
	var err error

	switch {
	case isNewWorker(w, latestD):
		latestD, err = c.deploymentRobin.createDeploymentFromWorker(w)
		if err != nil { return nil, err }
	case isWorkerUpdated(w, latestD):
		latestD, err = c.deploymentRobin.updateDeploymentOfWorker(w)
		if err != nil { return nil, err }
	}

	status, err := calculateWorkerStatus(w, latestD, dList)
	if err != nil { return nil, err }

	if isSameStatus(&w.Status, status) { return dList, nil }

	w.Status = *status
	w, err = c.syncWorkerStatus(w)
	if err != nil { return nil, err }

	return dList, nil
}

func calculateWorkerStatus(worker *workerv1.Worker, currentDeployment *appsv1.Deployment, dList []*appsv1.Deployment) (*workerv1.WorkerStatus, error) {
	d := findMatchDeployment(currentDeployment, dList)

	var status *workerv1.WorkerStatus
	if d == nil {
		status = newCreatedStatus(worker)
	} else {
		status = worker.Status.DeepCopy()
		updateReplicaCounts(status, worker, currentDeployment)
	}

	updateProgressingCondition(status, worker, currentDeployment)
	updateAvailabilityCondition(status, worker, currentDeployment)

	return status, nil
}

// isNewWorker checks if the worker passed in is new
func isNewWorker(worker *workerv1.Worker, currentDeployment *appsv1.Deployment) bool {
	return currentDeployment == nil && worker.DeletionTimestamp == nil
}

// isWorkerUpdated checks if the worker is updated
func isWorkerUpdated(worker *workerv1.Worker, currentDeployment *appsv1.Deployment) bool {
	deployment := generateDeploymentFromWorker(worker)
	res := !isSameDeploymentSpecs(deployment, currentDeployment)
	return res
}

// syncWorkerStatus syncs the stauts of the worker with the api server
func (c *Controller) syncWorkerStatus(w *workerv1.Worker) (*workerv1.Worker, error) {
	return c.workerclientset.WillesxmV1().Workers(w.Namespace).UpdateStatus(context.TODO(), w, metav1.UpdateOptions{})
}
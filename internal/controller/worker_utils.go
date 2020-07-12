package controller

import (
	"context"
	"fmt"
	workerv1 "github.com/willeslau/k8s-controller/pkg/apis/worker/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) updateWorkerStatus(w *workerv1.Worker) (*appsv1.Deployment, error) {
	newD, err := c.deploymentRobin.createDeploymentFromWorker(w)
	if err != nil {
		return nil, err
	}
	return newD, nil
}

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

func (c *Controller) updateWorker(w *workerv1.Worker, dList []*appsv1.Deployment) ([]*appsv1.Deployment, error) {
	latestD := findLatestDeployment(dList)

	if latestD == nil {
		newD, err := c.createNewWorker(w)
		if err != nil { return nil, err }
		dList = append(dList, newD)
		return dList, nil
	}

	switch compareWorkerAndDeployment(w, latestD).(type) {
	case *WorkerUpdated:
		fmt.Println("")
	case *WorkerUpdateCompleted:
		_, err := c.completeUpdate(w)
		if err != nil { return nil, err }
	case *WorkerUpdating:
		fmt.Println("progressing")
	//case "failed":
	//	fmt.Println("failed")
	default:
		fmt.Println("TODO")
	}
	return dList, nil
}

func compareWorkerAndDeployment(worker *workerv1.Worker, oldDeployment *appsv1.Deployment) WorkerUpdateEvent {
	newDeployment := generateDeploymentFromWorker(worker)
	equal := compareDeployments(newDeployment, oldDeployment)
	if !equal { return &WorkerUpdated{newDeployment} }

	if worker.Spec.Replicas == oldDeployment.Status.ReadyReplicas {
		return &WorkerUpdateCompleted{oldDeployment}
	}

	// TODO: add progressing status
	// TODO: add failure deployment
	return &WorkerUpdating{oldDeployment}
}

// syncWorkerStatus syncs the stauts of the worker with the api server
func (c *Controller) syncWorkerStatus(w *workerv1.Worker) (*workerv1.Worker, error) {
	return c.workerclientset.WillesxmV1().Workers(w.Namespace).UpdateStatus(context.TODO(), w, metav1.UpdateOptions{})
}
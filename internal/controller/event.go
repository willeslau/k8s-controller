package controller

import appsv1 "k8s.io/api/apps/v1"

type WorkerUpdateEvent interface {
	Deployment() *appsv1.Deployment
}

// WorkerUpdated is the event that a new worker is created
type WorkerCreationRequested struct {
	deployment *appsv1.Deployment
}

// Deployment returns the recorded deployment
func (w *WorkerCreationRequested) Deployment() *appsv1.Deployment {
	return w.deployment
}

// WorkerUpdated is the event that the worker yaml is updated
type WorkerUpdated struct {
	deployment *appsv1.Deployment
}

// Deployment returns the recorded deployment
func (w *WorkerUpdated) Deployment() *appsv1.Deployment {
	return w.deployment
}

type WorkerUpdateCompleted struct {
	deployment *appsv1.Deployment
}

// Deployment returns the recorded deployment
func (w *WorkerUpdateCompleted) Deployment() *appsv1.Deployment {
	return w.deployment
}

type WorkerUpdateFailed struct {

}

type WorkerUpdateTimeout struct {

}

type WorkerUpdating struct {
	deployment *appsv1.Deployment
}

// Deployment returns the recorded deployment
func (w *WorkerUpdating) Deployment() *appsv1.Deployment {
	return w.deployment
}

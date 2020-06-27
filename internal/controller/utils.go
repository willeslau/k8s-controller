package controller

import (
	workerv1 "github.com/willeslau/k8s-controller/pkg/apis/worker/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getDeploymentForWorker(worker *workerv1.Worker) *appsv1.Deployment {
	return nil
}

// TODO: complete this function
func compareDeployment(deployment *appsv1.Deployment, expectedDeployment *appsv1.Deployment) (bool, error) {
	return true, nil
}

func generateDeploymentFromWorker(worker *workerv1.Worker) *appsv1.Deployment {
	// create all the labels
	labels := worker.Labels
	labels["createdBy"] = worker.Name

	// define all the variables
	replicas := int32(1)
	name := worker.GenerateDeploymentName()
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: worker.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(worker, workerv1.SchemeGroupVersion.WithKind("Worker")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: worker.Spec.Image,
						},
					},
				},
			},
		},
	}
}

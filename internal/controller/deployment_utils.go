package controller

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"sort"

	workerv1 "github.com/willeslau/k8s-controller/pkg/apis/worker/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const REPLICAS = int32(1)

func createSelectorsFromLabels(ls map[string]string) (labels.Selector, error) {
	selector := labels.NewSelector()
	for key, val := range ls {
		r, err := labels.NewRequirement(key, selection.Equals, []string{val})
		if err != nil {
			return nil, err
		}
		selector = selector.Add(*r)
	}
	return selector, nil
}

func getDeploymentForWorker(worker *workerv1.Worker) *appsv1.Deployment {
	return nil
}

// TODO: complete this function
func compareDeployment(deployment *appsv1.Deployment, expectedDeployment *appsv1.Deployment) (bool, error) {
	return true, nil
}

func findLatestDeployment(dList []*appsv1.Deployment) *appsv1.Deployment {
	if dList == nil || len(dList) == 0 {
		return nil
	}

	sort.Slice(dList, func(i, j int) bool {
		if dList[i].CreationTimestamp.Equal(&dList[j].CreationTimestamp) {
			return dList[i].Name < dList[j].Name
		}
		return dList[j].CreationTimestamp.Before(&dList[j].CreationTimestamp)
	})

	return dList[0]
}

// DeploymentRobin is a collection of helper functions that require k8s client
type DeploymentRobin struct {
	kubeclientset kubernetes.Interface
}

func NewDeploymentRobin(kubeclientset kubernetes.Interface) *DeploymentRobin {
	return &DeploymentRobin{kubeclientset: kubeclientset}
}

// CreateDeploymentFromWorker will create the deployment from the worker
func (d *DeploymentRobin) CreateDeploymentFromWorker(w *workerv1.Worker) (*appsv1.Deployment, error) {
	deployment, err := d.kubeclientset.
		AppsV1().
		Deployments(w.Namespace).
		Create(context.TODO(), generateDeploymentFromWorker(w), metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		// TODO: add handler
		klog.Error("deployment alr exists {}", deployment.Name)
	}

	if err != nil {
		return nil, err
	}

	return deployment, nil
}

func generateDeploymentFromWorker(worker *workerv1.Worker) *appsv1.Deployment {
	// create all the labels
	wLabels := worker.Labels
	wLabels["createdBy"] = worker.Name

	// define all the variables
	replicas := REPLICAS
	//TODO: add hashing
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
				MatchLabels: wLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: wLabels,
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

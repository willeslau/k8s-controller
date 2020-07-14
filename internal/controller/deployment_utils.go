package controller

import (
	"context"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"hash/fnv"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"sort"

	workerv1 "github.com/willeslau/k8s-controller/pkg/apis/worker/v1"
	workerlister "github.com/willeslau/k8s-controller/pkg/client/listers/worker/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

// hashDeployment only hashes selective fields of the deployment
func hashDeployment(d *appsv1.Deployment) string {
	hasher := fnv.New32()

	spec := make(map[string]interface{})
	spec["replicas"] = *d.Spec.Replicas
	spec["volumes"] = d.Spec.Template.Spec.Volumes
	spec["containers"] = d.Spec.Template.Spec.Containers

	container := make(map[string]interface{})
	container["spec"] = spec
	container["label"] = d.Labels

	printer := spew.ConfigState{
		Indent: " ",
		SortKeys: true,
		DisableMethods: true,
		SpewKeys: true,
	}
	printer.Fprintf(hasher, "%#v", container)
	return rand.SafeEncodeString(fmt.Sprint(hasher.Sum32()))
}

func getDeploymentForWorker(worker *workerv1.Worker) *appsv1.Deployment {
	return nil
}

// orderByCreationTimeStampAndName sorts the input by desc creation timestamp and then name
func orderByCreationTimeStampAndName(dList []*appsv1.Deployment) []*appsv1.Deployment {
	sort.Slice(dList, func(i, j int) bool {
		if dList[i].CreationTimestamp.Equal(&dList[j].CreationTimestamp) {
			return dList[i].Name < dList[j].Name
		}
		return dList[j].CreationTimestamp.Before(&dList[j].CreationTimestamp)
	})
	return dList
}

func findLatestDeployment(dList []*appsv1.Deployment) *appsv1.Deployment {
	if dList == nil || len(dList) == 0 {
		return nil
	}
	dList = orderByCreationTimeStampAndName(dList)
	return dList[0]
}

// DeploymentRobin is a collection of helper functions that require k8s client
type DeploymentRobin struct {
	kubeclientset kubernetes.Interface
	wLister       workerlister.WorkerLister
}

func NewDeploymentRobin(kubeclientset kubernetes.Interface, wLister workerlister.WorkerLister) *DeploymentRobin {
	return &DeploymentRobin{kubeclientset: kubeclientset, wLister: wLister}
}

// updateDeploymentOfWorker updates the deployment of the worker with the deployment passed in
func (d *DeploymentRobin) updateDeploymentOfWorker(worker *workerv1.Worker, currentDeployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	deployment, err := d.kubeclientset.
		AppsV1().
		Deployments(worker.Namespace).
		Update(context.TODO(), currentDeployment, metav1.UpdateOptions{})

	if err != nil {
		return nil, err
	}

	return deployment, nil
}

// createDeploymentFromWorker will create the deployment from the worker
func (d *DeploymentRobin) createDeploymentFromWorker(w *workerv1.Worker) (*appsv1.Deployment, error) {
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

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (d *DeploymentRobin) resolveControllerRef(namespace string, ownerRef *metav1.OwnerReference) *workerv1.Worker {
	if ownerRef.Kind != ControllerKind { return nil }
	w, err := d.wLister.Workers(namespace).Get(ownerRef.Name)
	if err != nil { return nil }
	if w.UID != ownerRef.UID { return nil }
	return w
}

func compareDeploymentSpecs(newDeployment *appsv1.Deployment, oldDeployment *appsv1.Deployment) bool {
	newHash := hashDeployment(newDeployment)
	oldHash := hashDeployment(oldDeployment)
	return newHash == oldHash
}

func generateLabelsForDeployment(worker *workerv1.Worker) map[string]string {
	// create all the labels
	wLabels := worker.ObjectMeta.Labels
	wLabels["createdBy"] = worker.Name
	return wLabels
}

func generateDeploymentFromWorker(worker *workerv1.Worker) *appsv1.Deployment {
	wLabels := generateLabelsForDeployment(worker)

	// define all the variables
	//TODO: add hashing
	name := worker.GenerateDeploymentName()
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: worker.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(worker, workerv1.SchemeGroupVersion.WithKind("Worker")),
			},
			Labels: wLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &worker.Spec.Replicas,
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
							TerminationMessagePath: "/dev/termination-log",
							TerminationMessagePolicy: "File",
							ImagePullPolicy: "Always",
						},
					},
				},
			},
		},
	}
}

// findMatchDeployment finds the ref deployment in a list of deployments
func findMatchDeployment(refDeployment *appsv1.Deployment, dList []*appsv1.Deployment) *appsv1.Deployment {
	if dList == nil { return nil }

	for i := range dList {
		d := dList[i]
		if d.UID == refDeployment.UID { return d }
	}

	return nil
}

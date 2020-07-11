package controller

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	workerv1 "github.com/willeslau/k8s-controller/pkg/apis/worker/v1"
	clientset "github.com/willeslau/k8s-controller/pkg/client/clientset/versioned"
	workerscheme "github.com/willeslau/k8s-controller/pkg/client/clientset/versioned/scheme"
	informers "github.com/willeslau/k8s-controller/pkg/client/informers/externalversions/worker/v1"
	listers "github.com/willeslau/k8s-controller/pkg/client/listers/worker/v1"
)

const controllerAgentName = "worker-controller"

const (
	// SuccessSynced is what again? Donno for now
	SuccessSynced = "Synced"

	// MessageResourceSynced is the msg when worker resources are synced
	MessageResourceSynced = "Worker synced successfully"
)

// Controller is the controller implementation for Worker resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	// workerclientset is a clientset for our own API group
	workerclientset clientset.Interface

	workersLister listers.WorkerLister
	workersSynced cache.InformerSynced

	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface

	recorder record.EventRecorder

	deploymentRobin DeploymentRobin
}

// NewController returns a new worker controller
func NewController(
	kubeclientset kubernetes.Interface,
	workerclientset clientset.Interface,
	workerInformer informers.WorkerInformer,
	deploymentInformer appsinformers.DeploymentInformer) *Controller {

	utilruntime.Must(workerscheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	dRobin := DeploymentRobin{kubeclientset: kubeclientset}

	ctl := &Controller{
		kubeclientset:     kubeclientset,
		workerclientset:   workerclientset,
		workersLister:     workerInformer.Lister(),
		workersSynced:     workerInformer.Informer().HasSynced,
		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Workers"),
		recorder:          recorder,
		deploymentRobin: dRobin,
	}

	klog.Info("Setting up event handlers")
	// Set up an event handler for when Worker resources change
	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		// Test code only
		AddFunc: func(obj interface{}) {
		},
		UpdateFunc: func(old, new interface{}) {
		},
		DeleteFunc: func(obj interface{}) {
		},
	})

	workerInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctl.enqueueWorker,
		UpdateFunc: func(old, new interface{}) {
			oldWorker := old.(*workerv1.Worker)
			newWorker := new.(*workerv1.Worker)
			if oldWorker.ResourceVersion == newWorker.ResourceVersion {
				// same version, there is no need to do anything
				return
			}
			ctl.enqueueWorker(new)
		},
		DeleteFunc: ctl.enqueueWorkerForDelete,
	})

	return ctl
}

func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		klog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	klog.V(4).Infof("Processing object: %s", object.GetName())
	ownerRef := metav1.GetControllerOf(object)
	if ownerRef != nil {
		// If this object is not owned by a Foo, we should not do anything more
		// with it.
		if ownerRef.Kind != "Worker" {
			return
		}

		foo, err := c.workersLister.Workers(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			klog.V(4).Infof("ignoring orphaned object '%s' of foo '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueWorker(foo)
		return
	}
}

// addDeploymentController is a test function for deployment
func (c *Controller) addDeploymentController(obj interface{}) {
	d := obj.(*appsv1.Deployment)
	deploymentSelector, _ := metav1.LabelSelectorAsSelector(d.Spec.Selector)
	fmt.Println(deploymentSelector)
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Info("Start cache syncing")
	if ok := cache.WaitForCacheSync(stopCh, c.workersSynced); !ok {
		return fmt.Errorf("Failed to wait for caches to sync")
	}

	klog.Info("Launch worker go routines: ", threadiness)
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Workers are activated")
	<-stopCh
	klog.Info("Workers have stopped")
	klog.Info("Exit controller")

	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {

	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)

		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// the processing is done in here
		if err := c.syncHandler(key); err != nil {
			// Add the back to the queue for processing as
			// there are errors occurred
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}

		// Finally, if there are no errors, then we tell queue not to enqueue
		// from now on.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) getDeploymentsForWorker(w *workerv1.Worker) ([]*appsv1.Deployment, error) {
	dSelectors, err := createSelectorsFromLabels(w.ObjectMeta.Labels)
	if err != nil {
		return nil, err
	}
	dList, err := c.deploymentsLister.Deployments(w.Namespace).List(dSelectors)
	if errors.IsNotFound(err) {
		return make([]*appsv1.Deployment, 0), nil
	}
	if err != nil {
		return nil, err
	}
	if dList == nil {
		return make([]*appsv1.Deployment, 0), nil
	}
	return dList, nil
}

func (c *Controller) findDeploymentOfWorker(worker *workerv1.Worker) (*appsv1.Deployment, error) {
	name := worker.GenerateDeploymentName()

	// TODO: replace the todo context with a proper context
	d, err := c.kubeclientset.AppsV1().Deployments(worker.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return d, nil
}

// the actual processing logic lies here
func (c *Controller) syncHandler(key string) error {
	startTime := time.Now()
	klog.V(4).Infof("Started syncing Worker %q (%v)", key, startTime)
	defer func() {
		klog.V(4).Infof("Finished syncing Worker %q (%v)", key, startTime)
	}()

	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// get the actual object from the cache
	worker, err := c.workersLister.Workers(namespace).Get(name)
	// worker is deleted actually, perform the corresponding actions
	if errors.IsNotFound(err) {
		klog.V(4).Infof("Worker is deleted: %s/%s ...", namespace, name)
		return nil
	}
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("failed to list worker by: %s/%s", namespace, name))
		return err
	}

	// Deep copy Worker, do not touch the original object
	w := worker.DeepCopy()

	// list all the deployments linked to the worker
	dList, err := c.getDeploymentsForWorker(w)
	if err != nil {
		return err
	}
	klog.V(4).Infof("Found %d deployments for worker %p", len(dList), key)

	isUpdated, dList, err := c.updateWorker(w, dList)
	if err != nil {
		return err
	}

	// There are no updates, return straight away
	if !isUpdated { return nil }

	err = c.updateStatus(w, dList)
	if err != nil { return err }

	c.recorder.Event(worker, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) updateStatus(w *workerv1.Worker, dList []*appsv1.Deployment) error {
	return nil
}

func (c *Controller) updateWorker(w *workerv1.Worker, dList []*appsv1.Deployment) (bool, []*appsv1.Deployment, error) {
	latestD := findLatestDeployment(dList)
	isUpdated := false

	if latestD == nil {
		newD, err := c.deploymentRobin.CreateDeploymentFromWorker(w)
		if err != nil { return false, nil, err }
		dList = append(dList, newD)
		isUpdated = true
	}

	return isUpdated, dList, nil
}

// enqueueWorker takes a Worker resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Worker.
func (c *Controller) enqueueWorker(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}

	c.workqueue.AddRateLimited(key)
}

func (c *Controller) enqueueWorkerForDelete(obj interface{}) {
	var key string
	var err error
	key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

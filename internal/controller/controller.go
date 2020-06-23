package controller

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
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
	SuccessSynced = "Synced"

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

	workqueue workqueue.RateLimitingInterface

	recorder record.EventRecorder
}

// NewController returns a new worker controller
func NewController(
	kubeclientset kubernetes.Interface,
	workerclientset clientset.Interface,
	workerInformer informers.WorkerInformer) *Controller {

	utilruntime.Must(workerscheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:   kubeclientset,
		workerclientset: workerclientset,
		workersLister:   workerInformer.Lister(),
		workersSynced:   workerInformer.Informer().HasSynced,
		workqueue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Workers"),
		recorder:        recorder,
	}

	klog.Info("Setting up event handlers")
	// Set up an event handler for when Worker resources change
	workerInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueWorker,
		UpdateFunc: func(old, new interface{}) {
			oldWorker := old.(*workerv1.Worker)
			newWorker := new.(*workerv1.Worker)
			if oldWorker.ResourceVersion == newWorker.ResourceVersion {
				// same version, there is no need to do anything
				return
			}
			klog.V(4).Info(oldWorker)
			controller.enqueueWorker(new)
		},
		DeleteFunc: controller.enqueueWorkerForDelete,
	})

	return controller
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

// the actual processing logic lies here
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// get the actual object from the cache
	worker, err := c.workersLister.Workers(namespace).Get(name)
	if err != nil {
		// worker is deleted actually, perform the corresponding actions
		if errors.IsNotFound(err) {
			klog.Infof("Worker is deleted: %s/%s ...", namespace, name)
			return nil
		}

		utilruntime.HandleError(fmt.Errorf("failed to list worker by: %s/%s", namespace, name))
		return err
	}

	klog.Infof("expected state: ", worker)

	c.recorder.Event(worker, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
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

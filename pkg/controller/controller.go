package controller

import (
	"fmt"
	"time"

	"dink/pkg/k8s"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"k8s.io/klog/v2"
)

type Controller struct {
	Name          string
	ClusterClient kubernetes.Interface
	ClientSet     k8s.Interface
	Informer      cache.SharedIndexInformer
	EventHandler  Handler
	Queue         workqueue.RateLimitingInterface
}

// Result contains the result of a Reconciler invocation.
type Result struct {
	// Requeue tells the Controller to requeue the reconcile key.
	// If it is set to false, the controller will not requeue even if error occurred.
	// If it is nil, the controller will retry at most 3 times on errors.
	Requeue *bool

	// RequeueAfter if greater than 0, tells the Controller to requeue the reconcile key after the Duration.
	// Implies that Requeue is true, there is no need to set Requeue to true at the same time as RequeueAfter.
	RequeueAfter time.Duration
}

// Interface ...
type Handler interface {
	// Reconcile compares the actual state with the desired, and attempts to
	// converge the two.
	Reconcile(obj interface{}) (Result, error)

	// AddFinalizer adds a finalizer to the object if it not exists,
	// If a finalizer added, this func needs to update the object to the Kubernetes.
	AddFinalizer(obj interface{}) (added bool, err error)

	// HandleFinalizer needs to do things:
	// - execute the finalizer, like deleting any external resources associated with the obj
	// - remove the coorspending finalizer key from the obj
	// - update the object to the Kubernetes
	//
	// Ensure that this func must be idempotent and safe to invoke
	// multiple types for same object.
	HandleFinalizer(obj interface{}) error
}

// Run ...
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) {
	defer c.Queue.ShutDown()

	klog.Infof("Start controller %s threadiness %d.", c.Name, threadiness)

	go c.Informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("timeout to sync caches"))
	}

	klog.Infof("Controller %s synced and ready", c.Name)

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.work, time.Second, stopCh)
	}

	<-stopCh
	klog.Infof("Shutting down %s controller", c.Name)
}

// HasSynced ...
func (c *Controller) HasSynced() bool {
	return c.Informer.HasSynced()
}

func (c *Controller) work() {
	for c.nextWork() {
	}
}

func (c *Controller) nextWork() bool {
	key, shutdown := c.Queue.Get()
	if shutdown {
		return false
	}

	defer c.Queue.Done(key)
	res, err := c.doWork(key.(string))
	if res.Requeue != nil {
		requeue := *res.Requeue
		if !requeue {
			klog.Warningf("process %s failed (requeue=false, gave up)", key)
			c.Queue.Forget(key)
		} else if res.RequeueAfter > 0 {
			klog.Warningf("process %s failed (requeue=true, will retry after %s)", key, res.RequeueAfter)
			c.Queue.Forget(key)
			c.Queue.AddAfter(key, res.RequeueAfter)
		} else {
			klog.Warningf("process %s failed (requeue=true, will retry)", key)
			c.Queue.AddRateLimited(key)
		}
	} else if err != nil {
		if c.Queue.NumRequeues(key) < 3 {
			klog.Errorf("process %s failed (will retry): %v", key, err)
			c.Queue.AddRateLimited(key)
		} else {
			klog.Errorf("process %s failed (gave up): %v", key, err)
			c.Queue.Forget(key)
			utilruntime.HandleError(err)
		}
	} else {
		c.Queue.Forget(key)
	}

	return true
}

func (c *Controller) doWork(key string) (res Result, err error) {
	obj, exists, err := c.Informer.GetIndexer().GetByKey(key)
	if err != nil {
		return res, fmt.Errorf("error fetching object with key %s from store: %v", key, err)
	}

	if obj == nil || !exists {
		klog.Warningf("Object is nil or not exist, obj %v exist %v ", obj, exists)
		return res, nil
	}

	object, ok := obj.(metav1.Object)
	if !ok {
		klog.Warningf("Expect it is a Kubernetes resource object, got unknown type resource, %v", obj)
		return res, fmt.Errorf("unknown resource type")
	}

	// The object deletion timestamp is not zero value that indicates the resource is being deleted
	if !object.GetDeletionTimestamp().IsZero() {
		return res, c.EventHandler.HandleFinalizer(object)
	}

	// Add finalizer if needed
	added, err := c.EventHandler.AddFinalizer(object)
	if err != nil || added {
		return res, err
	}

	return c.EventHandler.Reconcile(obj)
}

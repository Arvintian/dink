package controllers

import (
	"dink/pkg/k8s"
	"reflect"

	"dink/pkg/controller"
	"dink/pkg/controller/handlers"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

func NewPodController(clusterClient kubernetes.Interface, client k8s.Interface) *controller.Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	factory := informers.NewSharedInformerFactoryWithOptions(
		clusterClient,
		controller.Config.ResyncPeriodSeconds,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = controller.PodSelector()
		}),
	)

	informer := factory.Core().V1().Pods().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				return
			}
			queue.Add(key)
		},
		UpdateFunc: func(old, new interface{}) {
			if reflect.DeepEqual(old, new) {
				return
			}
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err != nil {
				return
			}
			queue.Add(key)
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				return
			}
			queue.Add(key)
		},
	})

	return &controller.Controller{
		Name:          "Pod Controller",
		ClientSet:     client,
		ClusterClient: clusterClient,
		Informer:      informer,
		Queue:         queue,
		EventHandler:  handlers.NewPodHandler(client, clusterClient),
	}
}

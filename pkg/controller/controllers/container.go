package controllers

import (
	"dink/pkg/k8s"
	"reflect"

	"dink/pkg/controller"
	"dink/pkg/controller/handlers"

	"dink/pkg/generated/informers/externalversions"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

func NewContainerController(client k8s.Interface) *controller.Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	factory := externalversions.NewSharedInformerFactoryWithOptions(
		client,
		controller.Config.ResyncPeriodSeconds,
		externalversions.WithTweakListOptions(func(options *metav1.ListOptions) {
		}),
	)

	informer := factory.Dink().V1beta1().Containers().Informer()
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
		Name:          "Container Controller",
		ClientSet:     client,
		ClusterClient: client,
		Informer:      informer,
		Queue:         queue,
		EventHandler:  handlers.NewPodHandler(client, client),
	}
}

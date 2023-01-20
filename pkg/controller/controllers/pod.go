package controllers

import (
	"context"
	"dink/pkg/k8s"
	"reflect"

	dingv1beta1 "dink/pkg/apis/dink/v1beta1"
	"dink/pkg/controller"
	"dink/pkg/controller/handlers"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

func NewPodController(ctx context.Context, client k8s.Interface) *controller.Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	factory := informers.NewSharedInformerFactoryWithOptions(
		client,
		controller.Config.ResyncPeriods,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = dingv1beta1.PodSelector()
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
		ClusterClient: client,
		Informer:      informer,
		Queue:         queue,
		EventHandler:  handlers.NewPodHandler(ctx, client, client),
	}
}

package handlers

import (
	"dink/pkg/k8s"
	"fmt"

	dinkv1beta1 "dink/pkg/apis/dink/v1beta1"
	"dink/pkg/controller"

	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type ContainerHandler struct {
	ClusterClient kubernetes.Interface
	Client        k8s.Interface
}

var _ controller.Handler = (*ContainerHandler)(nil)

func NewContainerHandler(client k8s.Interface, clusterClient kubernetes.Interface) *ContainerHandler {
	return &ContainerHandler{
		Client:        client,
		ClusterClient: clusterClient,
	}
}

func (h *ContainerHandler) Reconcile(obj interface{}) (res controller.Result, err error) {
	originContainer, ok := obj.(*dinkv1beta1.Container)
	if !ok {
		return res, fmt.Errorf("unknown resource type")
	}

	container := originContainer.DeepCopy()
	klog.Infof("Reconcile Container %s/%s", container.Namespace, container.Name)
	return res, nil
}

func (h *ContainerHandler) AddFinalizer(obj interface{}) (bool, error) {
	return false, nil
}

func (h *ContainerHandler) HandleFinalizer(obj interface{}) error {
	return nil
}

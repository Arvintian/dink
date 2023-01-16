package handlers

import (
	"dink/pkg/k8s"
	"fmt"

	"dink/pkg/controller"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type PodHandler struct {
	ClusterClient kubernetes.Interface
	Client        k8s.Interface
}

var _ controller.Handler = (*PodHandler)(nil)

func NewPodHandler(client k8s.Interface, clusterClient kubernetes.Interface) *PodHandler {
	return &PodHandler{
		Client:        client,
		ClusterClient: clusterClient,
	}
}

func (h *PodHandler) Reconcile(obj interface{}) (res controller.Result, err error) {
	originPod, ok := obj.(*corev1.Pod)
	if !ok {
		return res, fmt.Errorf("unknown resource type")
	}

	pod := originPod.DeepCopy()
	klog.Infof("Reconcile Pod %s/%s", pod.Namespace, pod.Name)
	return res, nil
}

func (h *PodHandler) AddFinalizer(obj interface{}) (bool, error) {
	return false, nil
}

func (h *PodHandler) HandleFinalizer(obj interface{}) error {
	return nil
}

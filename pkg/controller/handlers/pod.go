package handlers

import (
	"context"
	"dink/pkg/k8s"
	"fmt"

	"dink/pkg/controller"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type PodHandler struct {
	ClusterClient kubernetes.Interface
	Client        k8s.Interface
	Context       context.Context
}

var _ controller.Handler = (*PodHandler)(nil)

const (
	finalizerPod string = "pod.dink.io/finalizer"
)

func NewPodHandler(ctx context.Context, clusterClient kubernetes.Interface, client k8s.Interface) *PodHandler {
	return &PodHandler{
		Client:        client,
		ClusterClient: client,
		Context:       ctx,
	}
}

func (h *PodHandler) Reconcile(obj interface{}) (res controller.Result, err error) {
	originPod, ok := obj.(*corev1.Pod)
	if !ok {
		return res, fmt.Errorf("unknown resource type")
	}
	pod := originPod.DeepCopy()

	container, err := h.Client.DinkV1beta1().Containers(pod.Namespace).Get(h.Context, pod.OwnerReferences[0].Name, metav1.GetOptions{})
	if err != nil {
		return res, err
	}

	container.Status.State = string(pod.Status.Phase)
	container.Status.PodStatus = &pod.Status
	_, err = h.Client.DinkV1beta1().Containers(container.Namespace).UpdateStatus(h.Context, container, metav1.UpdateOptions{})
	if err == nil {
		klog.Infof("update container %s/%s pod status %s", container.Namespace, container.Name, pod.Status.Phase)
	}
	return res, err
}

func (h *PodHandler) AddFinalizer(obj interface{}) (bool, error) {
	originPod, ok := obj.(*corev1.Pod)
	if !ok {
		return false, fmt.Errorf("unknown resource type")
	}

	if sets.NewString(originPod.Finalizers...).Has(finalizerPod) {
		return false, nil
	}

	pod := originPod.DeepCopy()
	pod.Finalizers = append(pod.Finalizers, finalizerPod)
	_, err := h.ClusterClient.CoreV1().Pods(pod.Namespace).Update(h.Context, pod, metav1.UpdateOptions{})
	return true, err
}

func (h *PodHandler) HandleFinalizer(obj interface{}) error {
	originPod, ok := obj.(*corev1.Pod)
	if !ok {
		return fmt.Errorf("unknown resource type")
	}
	if !sets.NewString(originPod.Finalizers...).Has(finalizerPod) {
		return nil
	}
	pod := originPod.DeepCopy()

	container, err := h.Client.DinkV1beta1().Containers(pod.Namespace).Get(h.Context, pod.OwnerReferences[0].Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	prevState := container.Status.State
	if container.Status.State == "Running" {
		container.Status.State = "Stopped"
	}
	container.Status.PodStatus = nil
	if _, err = h.Client.DinkV1beta1().Containers(container.Namespace).UpdateStatus(h.Context, container, metav1.UpdateOptions{}); err != nil {
		return err
	}
	if container.Status.State != prevState {
		klog.Infof("update container %s/%s status %s", container.Namespace, container.Name, container.Status.State)
	}

	pod.Finalizers = sets.NewString(pod.Finalizers...).Delete(finalizerPod).UnsortedList()
	_, err = h.Client.CoreV1().Pods(pod.Namespace).Update(h.Context, pod, metav1.UpdateOptions{})
	return err
}

package handlers

import (
	"context"
	"dink/pkg/k8s"
	"dink/pkg/utils"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	dinkv1beta1 "dink/pkg/apis/dink/v1beta1"
	"dink/pkg/apis/dink/v1beta1/template"
	"dink/pkg/controller"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type ContainerHandler struct {
	ClusterClient kubernetes.Interface
	Client        k8s.Interface
	Context       context.Context
}

var _ controller.Handler = (*ContainerHandler)(nil)

const (
	finalizerContainer string = "container.dink.io/finalizer"
)

func NewContainerHandler(ctx context.Context, client k8s.Interface, clusterClient kubernetes.Interface) *ContainerHandler {
	return &ContainerHandler{
		Client:        client,
		ClusterClient: clusterClient,
		Context:       ctx,
	}
}

func (h *ContainerHandler) Reconcile(obj interface{}) (res controller.Result, err error) {
	originContainer, ok := obj.(*dinkv1beta1.Container)
	if !ok {
		return res, fmt.Errorf("unknown resource type")
	}
	container := originContainer.DeepCopy()

	if container.Status.ContainerID == "" {
		id, err := h.createContainer(container)
		if err != nil {
			container.Status.State = dinkv1beta1.StateInitError
			klog.Errorf("create container %s/%s error", container.Namespace, container.Name)
			if _, err := h.Client.DinkV1beta1().Containers(container.Namespace).UpdateStatus(h.Context, container, metav1.UpdateOptions{}); err != nil {
				klog.Errorf("update container %s/%s status error %v", container.Namespace, container.Name, err)
			}
			return res, err
		}
		container.Status.State = dinkv1beta1.StateCreated
		container.Status.ContainerID = id
		if _, err := h.Client.DinkV1beta1().Containers(container.Namespace).UpdateStatus(h.Context, container, metav1.UpdateOptions{}); err != nil {
			klog.Errorf("update container %s/%s status error %v", container.Namespace, container.Name, err)
		}
		klog.Infof("create container %s/%s success", container.Namespace, container.Name)
		return res, err
	}

	if dinkv1beta1.IsFinalState(container.Status.State) && shouldRestart(container.Spec.RestartPolicy, container.Status.State) {
		// recreate container pod after 1s
		time.AfterFunc(time.Second, func() {
			co, err := h.Client.DinkV1beta1().Containers(container.Namespace).Get(h.Context, container.Name, metav1.GetOptions{})
			if err != nil {
				klog.Errorf("restart container %s/%s error %v", container.Namespace, container.Name, err)
				return
			}
			agentPod := template.CreatePodSepc(container, template.Config{
				Root:       controller.Config.Root,
				RunRoot:    controller.Config.RunRoot,
				RuncRoot:   controller.Config.RuncRoot,
				DockerData: controller.Config.DockerData,
				AgentImage: controller.Config.AgentImage,
				NFSServer:  controller.Config.NFSServer,
				NFSPath:    controller.Config.NFSPath,
			})
			if _, err := h.Client.CoreV1().Pods(co.Namespace).Create(h.Context, agentPod, metav1.CreateOptions{}); err != nil {
				klog.Errorf("restart container %s/%s error %v", container.Namespace, container.Name, err)
				return
			}
			klog.Infof("restart container %s/%s success", container.Namespace, container.Name)
		})
	}

	return res, nil
}

func (h *ContainerHandler) createContainer(c *dinkv1beta1.Container) (string, error) {
	dockerCli, err := client.NewClientWithOpts(client.WithHost(controller.Config.DockerHost), client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}
	envs := []string{}
	for _, item := range c.Spec.Template.Env {
		envs = append(envs, fmt.Sprintf("%s=%s", item.Name, item.Value))
	}
	config := &container.Config{
		Image:      c.Spec.Template.Image,
		Hostname:   c.Spec.HostName,
		Env:        envs,
		WorkingDir: c.Spec.Template.WorkingDir,
		Entrypoint: c.Spec.Template.Command,
		Cmd:        c.Spec.Template.Args,
		Tty:        c.Spec.Template.TTY,
	}

	// image pull
	filter := filters.NewArgs()
	filter.Add("reference", c.Spec.Template.Image)
	images, err := dockerCli.ImageList(h.Context, types.ImageListOptions{
		Filters: filter,
	})
	if err != nil {
		return "", err
	}
	if len(images) < 1 {
		out, err := dockerCli.ImagePull(h.Context, c.Spec.Template.Image, types.ImagePullOptions{})
		if err != nil {
			return "", err
		}
		defer out.Close()
		io.Copy(os.Stdout, out)
	}

	// create container
	createRsp, err := dockerCli.ContainerCreate(h.Context, config, nil, nil, nil, fmt.Sprintf("%s-%s", c.Namespace, c.Name))
	if err != nil {
		return "", err
	}
	inspectRsp, err := dockerCli.ContainerInspect(h.Context, createRsp.ID)
	if err != nil {
		return "", err
	}

	containerHome := filepath.Join(controller.Config.Root, "containers", createRsp.ID)
	if err := utils.CreateDir(containerHome, 0755); err != nil {
		return "", err
	}

	bts, err := json.Marshal(inspectRsp)
	if err != nil {
		return "", err
	}
	if err := utils.WriteBytesToFile(bts, filepath.Join(containerHome, "docker.json")); err != nil {
		return "", err
	}
	return createRsp.ID, nil
}

func (h *ContainerHandler) AddFinalizer(obj interface{}) (bool, error) {
	originContainer, ok := obj.(*dinkv1beta1.Container)
	if !ok {
		return false, fmt.Errorf("unknown resource type")
	}
	if sets.NewString(originContainer.Finalizers...).Has(finalizerContainer) {
		return false, nil
	}

	container := originContainer.DeepCopy()
	container.Finalizers = append(container.Finalizers, finalizerContainer)
	_, err := h.Client.DinkV1beta1().Containers(container.Namespace).Update(h.Context, container, metav1.UpdateOptions{})
	return true, err
}

func (h *ContainerHandler) HandleFinalizer(obj interface{}) error {
	originContainer, ok := obj.(*dinkv1beta1.Container)
	if !ok {
		return fmt.Errorf("unknown resource type")
	}
	if !sets.NewString(originContainer.Finalizers...).Has(finalizerContainer) {
		return nil
	}
	container := originContainer.DeepCopy()

	if err := h.deleteContainer(container); err != nil {
		return err
	}
	klog.Infof("delete container %s/%s success", container.Namespace, container.Name)

	container.Finalizers = sets.NewString(container.Finalizers...).Delete(finalizerContainer).UnsortedList()
	_, err := h.Client.DinkV1beta1().Containers(container.Namespace).Update(h.Context, container, metav1.UpdateOptions{})
	return err
}

func (h *ContainerHandler) deleteContainer(c *dinkv1beta1.Container) error {
	if c.Status.ContainerID == "" {
		return nil
	}

	dockerCli, err := client.NewClientWithOpts(client.WithHost(controller.Config.DockerHost), client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	if err := dockerCli.ContainerRemove(h.Context, c.Status.ContainerID, types.ContainerRemoveOptions{}); err != nil {
		return err
	}

	return os.RemoveAll(filepath.Join(controller.Config.Root, "containers", c.Status.ContainerID))
}

func shouldRestart(policy, state string) bool {
	if !dinkv1beta1.IsFinalState(state) {
		return false
	}
	if policy == dinkv1beta1.RestartPolicyNever {
		return false
	}
	if policy == dinkv1beta1.RestartPolicyAlways {
		return true
	}
	if policy == dinkv1beta1.RestartPolicyUnlessStopped && state != dinkv1beta1.StateStopped {
		return true
	}
	if policy == dinkv1beta1.RestartPolicyOnFailure && state == dinkv1beta1.StateFailed {
		return true
	}
	return false
}

package handlers

import (
	"dink/pkg/controller"
	"fmt"
	"net/http"

	dinkv1beta1 "dink/pkg/apis/dink/v1beta1"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func StartContainer(c *gin.Context) {
	namespace, name := c.Param("namespace"), c.Param("name")
	client := Config.Client

	container, err := client.DinkV1beta1().Containers(namespace).Get(c, name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	if errors.IsNotFound(err) {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	if container.Status.ContainerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "container not created or init error",
		})
		return
	}

	if container.Status.PodStatus != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "container not stopped",
		})
	}

	privileged := true
	agentPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", name, container.Status.ContainerID[:12]),
			Labels: map[string]string{
				controller.LabelPodCreatedBy: controller.DinkCreator,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: dinkv1beta1.APIVersion,
					Kind:       dinkv1beta1.ContainerKind,
					Name:       container.Name,
					UID:        container.UID,
				},
			},
		},
		Spec: corev1.PodSpec{
			DNSPolicy: corev1.DNSClusterFirst,
			Hostname:  container.Spec.HostName,
			Containers: []corev1.Container{
				{
					Name:  "dink-agent",
					Image: Config.AgentImage,
					Command: []string{
						"/app/agent",
						"--root",
						Config.Root,
						"--run-root",
						Config.RunRoot,
						"--runc-root",
						Config.RuncRoot,
						"--docker-data",
						Config.DockerData,
						"start",
						"--id",
						container.Status.ContainerID,
					},
					Lifecycle: &corev1.Lifecycle{
						PreStop: &corev1.Handler{
							Exec: &corev1.ExecAction{
								Command: []string{
									"runc",
									"--root",
									Config.RuncRoot,
									"kill",
									container.Status.ContainerID,
									"SIGTERM",
								},
							},
						},
					},
					LivenessProbe:  container.Spec.Template.LivenessProbe,
					ReadinessProbe: container.Spec.Template.ReadinessProbe,
					Resources:      container.Spec.Template.Resources,
					SecurityContext: &corev1.SecurityContext{
						Privileged: &privileged,
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "dink-root",
							ReadOnly:  false,
							MountPath: Config.Root,
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "dink-root",
					VolumeSource: corev1.VolumeSource{
						NFS: &corev1.NFSVolumeSource{
							Server:   Config.NFSServer,
							Path:     Config.NFSPath,
							ReadOnly: false,
						},
					},
				},
			},
		},
	}

	if _, err := client.CoreV1().Pods(namespace).Create(c, agentPod, metav1.CreateOptions{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	container.Status.State = "Starting"
	if _, err := client.DinkV1beta1().Containers(container.Namespace).UpdateStatus(c, container, metav1.UpdateOptions{}); err != nil {
		klog.Error(err)
	}

	c.JSON(http.StatusOK, gin.H{})
}

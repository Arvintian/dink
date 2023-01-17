package handlers

import (
	"dink/pkg/controller"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	privileged := true
	agentPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", name, container.Status.ContainerID[:12]),
			Labels: map[string]string{
				controller.LabelPodCreatedBy: controller.DinkCreator,
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
						"--runc-root",
						Config.RuncRoot,
						"--docker-data",
						Config.DockerData,
						"start",
						"--id",
						container.Status.ContainerID,
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

	c.JSON(http.StatusOK, gin.H{})
}

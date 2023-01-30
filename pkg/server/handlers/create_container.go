package handlers

import (
	"net/http"

	dinkv1beta1 "dink/pkg/apis/dink/v1beta1"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateContainer(c *gin.Context) {
	namespace, name := c.Param("namespace"), c.Param("name")
	client := Config.Client

	payload := defaultContainerConfig()
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	container := dinkv1beta1.Container{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: dinkv1beta1.ContainerSpec{
			RestartPolicy: payload.RestartPolicy,
			HostName:      payload.HostName,
			Template: corev1.Container{
				Image:      payload.Image,
				Env:        payload.Env,
				WorkingDir: payload.WorkingDir,
				Command:    payload.Entrypoint,
				Args:       payload.Cmd,
				Stdin:      payload.Stdin,
				TTY:        payload.TTY,
				Resources:  payload.Resources,
				SecurityContext: &corev1.SecurityContext{
					RunAsUser:  payload.UID,
					RunAsGroup: payload.GID,
				},
			},
		},
	}

	if _, err := client.DinkV1beta1().Containers(namespace).Create(c, &container, metav1.CreateOptions{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

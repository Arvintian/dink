package handlers

import (
	"net/http"

	"dink/pkg/apis/dink/v1beta1/template"

	"github.com/gin-gonic/gin"
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

	if container.Status.ContainerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "container not created or init error",
		})
		return
	}

	if container.Status.PodStatus != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "container is running",
		})
		return
	}

	agentPod := template.CreatePodSepc(container, template.Config{
		Root:       Config.Root,
		RunRoot:    Config.RunRoot,
		RuncRoot:   Config.RuncRoot,
		DockerData: Config.DockerData,
		AgentImage: Config.AgentImage,
		NFSServer:  Config.NFSServer,
		NFSPath:    Config.NFSPath,
		NFSOptions: Config.NFSOptions,
	})

	if _, err := client.CoreV1().Pods(namespace).Create(c, agentPod, metav1.CreateOptions{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

package handlers

import (
	"net/http"

	dinkv1beta1 "dink/pkg/apis/dink/v1beta1"
	"dink/pkg/apis/dink/v1beta1/template"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func StopContainer(c *gin.Context) {
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

	if dinkv1beta1.IsFinalState(container.Status.State) || container.Status.PodStatus == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "container not is running",
		})
		return
	}

	thePod := template.GetPodName(container)
	if err := client.CoreV1().Pods(namespace).Delete(c, thePod, metav1.DeleteOptions{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

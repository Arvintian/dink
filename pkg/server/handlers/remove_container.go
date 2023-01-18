package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RemoveContainer(c *gin.Context) {
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

	if container.Status.PodStatus != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "container is running",
		})
		return
	}

	if err := client.DinkV1beta1().Containers(namespace).Delete(c, name, metav1.DeleteOptions{}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

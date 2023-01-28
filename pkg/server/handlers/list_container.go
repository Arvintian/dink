package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListContainer(c *gin.Context) {
	namespace := c.Param("namespace")
	client := Config.Client

	containers, err := client.DinkV1beta1().Containers(namespace).List(c, metav1.ListOptions{})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"containers": containers.Items,
	})
}

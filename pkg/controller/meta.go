package controller

import (
	"fmt"
)

const (
	LabelPodCreatedBy = "pod.kubernetes.io/created-by"
	DinkCreator       = "dink"
)

func PodSelector() string {
	return fmt.Sprintf("%s=%s", LabelPodCreatedBy, DinkCreator)
}

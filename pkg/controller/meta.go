package controller

import (
	"fmt"
)

const (
	LabelPodCreatedBy  = "pod.kubernetes.io/created-by"
	AnnotationSpecHash = "spec.dink.io/hash"
	DinkCreator        = "dink"
)

func PodSelector() string {
	return fmt.Sprintf("%s=%s", LabelPodCreatedBy, DinkCreator)
}

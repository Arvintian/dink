package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Container is a specification for a Container resource
type Container struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ContainerSpec   `json:"spec"`
	Status ContainerStatus `json:"status"`
}

// ContainerSpec is the spec for a Container resource
type ContainerSpec struct {
	RestartPolicy string           `json:"restartPolicy"`
	Template      corev1.Container `json:"template"`
}

// ContainerStatus is the status for a Container resource
type ContainerStatus struct {
	State int32 `json:"state"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ContainerList is a list of Container resources
type ContainerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Container `json:"items"`
}

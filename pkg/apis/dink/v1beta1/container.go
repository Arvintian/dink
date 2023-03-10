package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ContainerKind       = "Container"
	ContainerPluralName = "containers"
)

const (
	StateInit      = "Init"
	StateInitError = "InitError"
	StateCreated   = "Created"
	StatePending   = "Pending"
	StateRunning   = "Running"
	StateFailed    = "Failed"
	StateSucceeded = "Succeeded"
	StateUnknown   = "Unknown"
	StateStopped   = "Stopped"
)

const (
	RestartPolicyAlways        = "Always"
	RestartPolicyOnFailure     = "OnFailure"
	RestartPolicyUnlessStopped = "UnlessStopped"
	RestartPolicyNever         = "Never"
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
	HostName      string            `json:"hostname"`
	RestartPolicy string            `json:"restartPolicy"`
	NodeSelector  map[string]string `json:"nodeSelector"`
	Template      corev1.Container  `json:"template"`
	Volumes       []corev1.Volume   `json:"volumes"`
}

// ContainerStatus is the status for a Container resource
type ContainerStatus struct {
	State       string            `json:"state"`
	ContainerID string            `json:"containerID"`
	PodStatus   *corev1.PodStatus `json:"podStatus"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ContainerList is a list of Container resources
type ContainerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Container `json:"items"`
}

func IsFinalState(state string) bool {
	finalState := []string{StateSucceeded, StateFailed, StateStopped}
	for _, final := range finalState {
		if state == final {
			return true
		}
	}
	return false
}

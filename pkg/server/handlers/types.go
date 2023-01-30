package handlers

import (
	dinkv1beta1 "dink/pkg/apis/dink/v1beta1"

	corev1 "k8s.io/api/core/v1"
)

type ContainerConfig struct {
	Image         string                      `json:"image"`
	HostName      string                      `json:"hostName"`
	RestartPolicy string                      `json:"restartPolicy"`
	Env           []corev1.EnvVar             `json:"env"`
	WorkingDir    string                      `json:"workingDir"`
	Entrypoint    []string                    `json:"entrypoint"`
	Cmd           []string                    `json:"cmd"`
	Stdin         bool                        `json:"stdin"`
	TTY           bool                        `json:"tty"`
	UID           *int64                      `json:"uid"`
	GID           *int64                      `json:"gid"`
	Resources     corev1.ResourceRequirements `json:"resources"`
}

func defaultContainerConfig() ContainerConfig {
	return ContainerConfig{
		RestartPolicy: dinkv1beta1.RestartPolicyNever,
		Env:           []corev1.EnvVar{},
		Entrypoint:    []string{},
		Cmd:           []string{},
	}
}

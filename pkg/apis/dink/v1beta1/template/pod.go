package template

import (
	"dink/pkg/controller"
	"fmt"

	dinkv1beta1 "dink/pkg/apis/dink/v1beta1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Config struct {
	Root       string
	RunRoot    string
	RuncRoot   string
	DockerData string
	AgentImage string
	NFSServer  string
	NFSPath    string
}

func GetPodName(container *dinkv1beta1.Container) string {
	return fmt.Sprintf("%s-%s", container.Name, container.Status.ContainerID[:12])
}

func CreatePodSepc(container *dinkv1beta1.Container, cfg Config) *corev1.Pod {
	privileged := true
	labels := map[string]string{
		controller.LabelPodCreatedBy: controller.DinkCreator,
	}
	for k, v := range container.Labels {
		labels[k] = v
	}
	volumes := []corev1.Volume{
		{
			Name: "dink-root",
			VolumeSource: corev1.VolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Server:   cfg.NFSServer,
					Path:     cfg.NFSPath,
					ReadOnly: false,
				},
			},
		},
	}
	volumes = append(volumes, container.Spec.Volumes...)
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "dink-root",
			ReadOnly:  false,
			MountPath: cfg.Root,
		},
	}
	volumeMounts = append(volumeMounts, container.Spec.Template.VolumeMounts...)
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   GetPodName(container),
			Labels: labels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: dinkv1beta1.APIVersion,
					Kind:       dinkv1beta1.ContainerKind,
					Name:       container.Name,
					UID:        container.UID,
				},
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			DNSPolicy:     corev1.DNSClusterFirst,
			Hostname:      container.Spec.HostName,
			NodeSelector:  container.Spec.NodeSelector,
			Containers: []corev1.Container{
				{
					Name:  "dink-agent",
					Image: cfg.AgentImage,
					Command: []string{
						"/app/agent",
						"--root",
						cfg.Root,
						"--run-root",
						cfg.RunRoot,
						"--runc-root",
						cfg.RuncRoot,
						"--docker-data",
						cfg.DockerData,
						"start",
						"--id",
						container.Status.ContainerID,
					},
					Stdin: container.Spec.Template.Stdin,
					TTY:   container.Spec.Template.TTY,
					Lifecycle: &corev1.Lifecycle{
						PreStop: &corev1.Handler{
							Exec: &corev1.ExecAction{
								Command: []string{
									"runc",
									"--root",
									cfg.RuncRoot,
									"kill",
									container.Status.ContainerID,
									"SIGTERM",
								},
							},
						},
					},
					LivenessProbe:  container.Spec.Template.LivenessProbe,
					ReadinessProbe: container.Spec.Template.ReadinessProbe,
					Resources:      container.Spec.Template.Resources,
					SecurityContext: &corev1.SecurityContext{
						Privileged: &privileged,
					},
					VolumeMounts: volumeMounts,
				},
			},
			Volumes: volumes,
		},
	}

}

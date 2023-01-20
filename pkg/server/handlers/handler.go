package handlers

import (
	"dink/pkg/apis/dink/v1beta1/template"
	"dink/pkg/k8s"
)

type ServerConfig struct {
	Client k8s.Interface
	template.Config
}

var Config ServerConfig

func init() {
	Config.Root = "/var/lib/dink"
	Config.RunRoot = "/run/dink"
	Config.RuncRoot = "/run/dink/runc"
	Config.DockerData = "/var/lib/dink/docker"
}

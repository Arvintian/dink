package handlers

import (
	"dink/pkg/k8s"
)

type ServerConfig struct {
	Root       string
	RuncRoot   string
	DockerData string
	AgentImage string
	NFSServer  string
	NFSPath    string
	Client     k8s.Interface
}

var Config ServerConfig

func init() {
	Config.Root = "/var/lib/dink"
	Config.RuncRoot = "/run/dink/runc"
	Config.DockerData = "/var/lib/dink/docker"
}

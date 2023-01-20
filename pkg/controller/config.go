package controller

import (
	"dink/pkg/apis/dink/v1beta1/template"
	"time"
)

type ControllerConfig struct {
	ResyncPeriods time.Duration
	DockerHost    string
	template.Config
}

var Config ControllerConfig

func init() {
	Config.ResyncPeriods = 15 * 60 * time.Second
	Config.DockerHost = "tcp://127.0.0.1:2375"
	Config.Root = "/var/lib/dink"
	Config.RunRoot = "/run/dink"
	Config.RuncRoot = "/run/dink/runc"
	Config.DockerData = "/var/lib/dink/docker"
}

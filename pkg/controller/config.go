package controller

import "time"

type ControllerConfig struct {
	ResyncPeriods time.Duration
	Root          string
	RunRoot       string
	RuncRoot      string
	DockerData    string
	DockerHost    string
	AgentImage    string
	NFSServer     string
	NFSPath       string
}

var Config ControllerConfig

func init() {
	Config.ResyncPeriods = 15 * 60 * time.Second
	Config.Root = "/var/lib/dink"
	Config.RunRoot = "/run/dink"
	Config.RuncRoot = "/run/dink/runc"
	Config.DockerData = "/var/lib/dink/docker"
	Config.DockerHost = "tcp://127.0.0.1:2375"
}

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

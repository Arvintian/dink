package controller

import "time"

type ControllerConfig struct {
	// ResyncPeriodSeconds defines resync period in seconds for controllers
	ResyncPeriodSeconds time.Duration `json:"resync_period_seconds"`
}

var Config ControllerConfig

package goru

import "gottb.io/goru/config"

type Hook func(config *config.Config) error

var startHooks = []Hook{}
var stopHooks = []Hook{}

func StartWith(h Hook) {
	startHooks = append(startHooks, h)
}

func StopWith(h Hook) {
	stopHooks = append(stopHooks, h)
}

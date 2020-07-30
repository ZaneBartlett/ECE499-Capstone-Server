package components

import (
	"tech/mixer/config"
)

// MixerComponentIf -
type MixerComponentIf interface {
	Action(action string, data []byte) (response []byte, err error)
	Start() error
	Stop() error
}

// MixerComponent - A Component of the Mixer device
type MixerComponent struct {
	Name          string
	ConfigService *config.CfgService
}

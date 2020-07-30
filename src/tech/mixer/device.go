package mixer

import (
	"fmt"
	"os/exec"
	"tech/app/comms"
	"tech/app/components"
	"tech/app/logger"
	"tech/mixer/config"
)

const (
	// DatabaseName - Database filepath
	DatabaseName = "/data/config.db"
)

// Mixer - Time Code Processor struct
type Mixer struct {

	// List of all instantiated Objects
	ComponentList map[string]components.MixerComponentIf
	cfgService    *config.CfgService
	UserAuth      *comms.UserAuth
	MixerControl  *components.MixerControl
	Factory       *Factory
}

// NewMixer - Instantiates the device's Mixer object
func NewMixer() *Mixer {

	mixer := Mixer{}

	cfgService, err := config.NewCfgService(DatabaseName)
	if err != nil {
		logger.Log("Failed to create config service, error is %v", err)
	}
	mixer.cfgService = cfgService

	mixer.ComponentList = make(map[string]components.MixerComponentIf)

	userAuth := comms.NewUserAuth(mixer.cfgService)
	mixer.ComponentList[userAuth.Name] = userAuth
	mixer.UserAuth = userAuth

	mixerControl := components.NewMixerControl(mixer.cfgService)
	mixer.ComponentList[mixerControl.Name] = mixerControl
	mixer.MixerControl = mixerControl

	factory := NewFactory(mixer.cfgService)
	mixer.ComponentList[factory.Name] = factory
	mixer.Factory = factory

	// Create and Initialize database tables if needed
	mixer.cfgService.Initialize()

	return &mixer
}

func (mixer *Mixer) Start() error {

	var err error

	for _, component := range mixer.ComponentList {
		err = component.Start()
		if err != nil {
			logger.Log("Failed to start component, err is '%v'", err)
			return err
		}
	}

	return nil
}

// Reset - Power cycle the device via a shellscript. Script must be in the same
// directory as Host.go
func (mixer *Mixer) Reset() {

	_, err := exec.Command("reboot").Output()
	if err != nil {
		logger.Log("Unable to reset device, %v", err)
	}

}

func (mixer *Mixer) PowerOff() {

	_, err := exec.Command("shutdown", "-h", "now").Output()
	if err != nil {
		logger.Log("Unable to reset device, %v", err)
	}

}

// Action - Iterates through the clock's available objects and executes an 'action'
func (mixer *Mixer) Action(target string, action string, data []byte) (response []byte, err error) {

	if action == "Reboot" {

		mixer.Reset()
		logger.Log("Power cycling device, connection will be lost")

		return
	}

	if action == "PowerOff" {

		mixer.PowerOff()
		logger.Log("Powering off device, connection will be lost")

		return
	}

	for key, val := range mixer.ComponentList {

		if target == key {
			response, err := val.Action(action, data)
			return response, err
		}
	}

	response = nil
	err = fmt.Errorf("Failed to find target: %s", target)
	return
}

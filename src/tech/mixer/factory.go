package mixer

import (
	"fmt"
	"os/exec"
	"strings"
	"tech/app/components"
	"tech/app/logger"
	"tech/mixer/config"
)

// Factory -
type Factory struct {
	components.MixerComponent

	Name string
}

// NewFactory -
func NewFactory(cfg *config.CfgService) *Factory {
	factory := &Factory{}
	factory.Name = "factory"
	factory.ConfigService = cfg

	cfg.Register(factory.Name, factory.createNetworkTable)

	return factory
}

func (fact *Factory) Start() error {
	initNetworkData, err := fact.ConfigService.Get(fact.Name)
	if err != nil {
		return err
	}

	netMap, err := config.JsonToMap(initNetworkData)
	if err != nil {
		return err
	}
	enableDhcp, _ := config.JSONbool(netMap["enableDhcp"])
	if !enableDhcp {
		_, err = fact.setNetworkInfo(netMap, "db")
	}

	return err
}

func (fact *Factory) Stop() error {
	return nil
}

// Action -
func (fact *Factory) Action(action string, data []byte) (response []byte, err error) {
	var mapData map[string]interface{}
	mapData, err = config.JsonToMap(data)
	if err != nil {
		return
	}

	switch action {
	case "GetNetwork":
		response, err = fact.ConfigService.Get(fact.Name)

	case "SetNetwork":
		response, err = fact.setNetworkInfo(mapData, "ui")

	default:
		logger.Log("unrecognised action received in factory")

	}

	return
}

func (fact *Factory) createNetworkTable(cfg *config.CfgService) (err error) {
	networkSchema := []string{
		"enableDhcp INTEGER",
		"ipAddress1 INTEGER",
		"ipAddress2 INTEGER",
		"ipAddress3 INTEGER",
		"ipAddress4 INTEGER",
		"ipMask1 INTEGER",
		"ipMask2 INTEGER",
		"ipMask3 INTEGER",
		"ipMask4 INTEGER",
		"gateway1 INTEGER",
		"gateway2 INTEGER",
		"gateway3 INTEGER",
		"gateway4 INTEGER",
		"nfcMode  INTEGER"}

	networkDefaults := fact.getNetworkDefaults()

	err = cfg.CreateTable(fact.Name, networkSchema)
	if err != nil {
		return
	}

	err = cfg.InitTable(fact.Name, networkDefaults)
	if err != nil {
		return
	}

	return nil
}

func (fact *Factory) SetNetworkInfo(data map[string]interface{}) ([]byte, error) {
	return fact.setNetworkInfo(data, "ui")
}

func (fact *Factory) setNetworkInfo(mapData map[string]interface{}, source string) ([]byte, error) {

	var currentData map[string]string

	ipAddressString := ""
	ipMaskString := ""
	gatewayString := ""

	//apply to db
	if source == "ui" {
		fact.SetNetworkTable(mapData)

		currentData = fact.getNetworkDefaults()

		ipAddress := mapData["ipAddress"].(map[string]interface{})
		ipMask := mapData["ipMask"].(map[string]interface{})
		gateway := mapData["gateway"].(map[string]interface{})

		ipAddressString = fmt.Sprintf("%d", int64(ipAddress["ipAddress1"].(float64))) + "." + fmt.Sprintf("%d", int64(ipAddress["ipAddress2"].(float64))) + "." + fmt.Sprintf("%d", int64(ipAddress["ipAddress3"].(float64))) + "." + fmt.Sprintf("%d", int64(ipAddress["ipAddress4"].(float64)))
		ipMaskString = fmt.Sprintf("%d", int64(ipMask["ipMask1"].(float64))) + "." + fmt.Sprintf("%d", int64(ipMask["ipMask2"].(float64))) + "." + fmt.Sprintf("%d", int64(ipMask["ipMask3"].(float64))) + "." + fmt.Sprintf("%d", int64(ipMask["ipMask4"].(float64)))
		gatewayString = fmt.Sprintf("%d", int64(gateway["gateway1"].(float64))) + "." + fmt.Sprintf("%d", int64(gateway["gateway2"].(float64))) + "." + fmt.Sprintf("%d", int64(gateway["gateway3"].(float64))) + "." + fmt.Sprintf("%d", int64(gateway["gateway4"].(float64)))

	} else if source == "db" {
		fact.SetNetworkTableFromDb(mapData)

		currentData = fact.getNetworkDefaults()

		ipAddressString = currentData["ipAddress1"] + "." + currentData["ipAddress2"] + "." + currentData["ipAddress3"] + "." + currentData["ipAddress4"]
		ipMaskString = currentData["ipMask1"] + "." + currentData["ipMask2"] + "." + currentData["ipMask3"] + "." + currentData["ipMask4"]
		gatewayString = currentData["gateway1"] + "." + currentData["gateway2"] + "." + currentData["gateway3"] + "." + currentData["gateway4"]
	}

	//apply to device

	logger.Log("ipAddress: " + ipAddressString)
	logger.Log("ipMask: " + ipMaskString)
	logger.Log("gateway: " + gatewayString)

	out, err := exec.Command("ifconfig", "eth0", ipAddressString, "netmask", ipMaskString).Output()
	if err != nil {
		logger.Log("ifconfig error: %v", err)
	}

	currentGateway := currentData["gateway1"] + "." + currentData["gateway2"] + "." + currentData["gateway3"] + "." + currentData["gateway4"]
	out, err = exec.Command("route", "del", "default", "gw", currentGateway).Output()
	out, err = exec.Command("route", "add", "default", "gw", gatewayString).Output()
	if err != nil {
		logger.Log("gateway error: %v", err)
	}
	logger.LogDebug(string(out))

	return fact.ConfigService.Get(fact.Name)
}

func (fact *Factory) SetNetworkTable(mapData map[string]interface{}) {
	fact.ConfigService.SetValue(fact.Name, "enableDhcp", mapData["enableDhcp"])
	fact.ConfigService.SetValue(fact.Name, "ipAddress1", (mapData["ipAddress"].(map[string]interface{}))["ipAddress1"])
	fact.ConfigService.SetValue(fact.Name, "ipAddress2", (mapData["ipAddress"].(map[string]interface{}))["ipAddress2"])
	fact.ConfigService.SetValue(fact.Name, "ipAddress3", (mapData["ipAddress"].(map[string]interface{}))["ipAddress3"])
	fact.ConfigService.SetValue(fact.Name, "ipAddress4", (mapData["ipAddress"].(map[string]interface{}))["ipAddress4"])
	fact.ConfigService.SetValue(fact.Name, "ipMask1", (mapData["ipMask"].(map[string]interface{}))["ipMask1"])
	fact.ConfigService.SetValue(fact.Name, "ipMask2", (mapData["ipMask"].(map[string]interface{}))["ipMask2"])
	fact.ConfigService.SetValue(fact.Name, "ipMask3", (mapData["ipMask"].(map[string]interface{}))["ipMask3"])
	fact.ConfigService.SetValue(fact.Name, "ipMask4", (mapData["ipMask"].(map[string]interface{}))["ipMask4"])
	fact.ConfigService.SetValue(fact.Name, "gateway1", (mapData["gateway"].(map[string]interface{}))["gateway1"])
	fact.ConfigService.SetValue(fact.Name, "gateway2", (mapData["gateway"].(map[string]interface{}))["gateway2"])
	fact.ConfigService.SetValue(fact.Name, "gateway3", (mapData["gateway"].(map[string]interface{}))["gateway3"])
	fact.ConfigService.SetValue(fact.Name, "gateway4", (mapData["gateway"].(map[string]interface{}))["gateway4"])
	fact.ConfigService.SetValue(fact.Name, "nfcMode", mapData["nfcMode"])
}

func (fact *Factory) SetNetworkTableFromDb(mapData map[string]interface{}) {
	fact.ConfigService.SetValue(fact.Name, "enableDhcp", mapData["enableDhcp"])
	fact.ConfigService.SetValue(fact.Name, "ipAddress1", mapData["ipAddress1"])
	fact.ConfigService.SetValue(fact.Name, "ipAddress2", mapData["ipAddress2"])
	fact.ConfigService.SetValue(fact.Name, "ipAddress3", mapData["ipAddress3"])
	fact.ConfigService.SetValue(fact.Name, "ipAddress4", mapData["ipAddress4"])
	fact.ConfigService.SetValue(fact.Name, "ipMask1", mapData["ipMask1"])
	fact.ConfigService.SetValue(fact.Name, "ipMask2", mapData["ipMask2"])
	fact.ConfigService.SetValue(fact.Name, "ipMask3", mapData["ipMask3"])
	fact.ConfigService.SetValue(fact.Name, "ipMask4", mapData["ipMask4"])
	fact.ConfigService.SetValue(fact.Name, "gateway1", mapData["gateway1"])
	fact.ConfigService.SetValue(fact.Name, "gateway2", mapData["gateway2"])
	fact.ConfigService.SetValue(fact.Name, "gateway3", mapData["gateway3"])
	fact.ConfigService.SetValue(fact.Name, "gateway4", mapData["gateway4"])
	fact.ConfigService.SetValue(fact.Name, "nfcMode", mapData["nfcMode"])
}

func (fact *Factory) getNetworkDefaults() map[string]string {

	networkDefaults := map[string]string{
		"enableDhcp": "1",
		"ipAddress1": "10",
		"ipAddress2": "0",
		"ipAddress3": "0",
		"ipAddress4": "90",
		"ipMask1":    "255",
		"ipMask2":    "255",
		"ipMask3":    "255",
		"ipMask4":    "255",
		"gateway1":   "10",
		"gateway2":   "0",
		"gateway3":   "0",
		"gateway4":   "0",
		"nfcMode":    "0"}

	ipAddress := ""
	ipMask := ""
	gateway := ""

	cmd := exec.Command("ifconfig", "eth0")
	grep := exec.Command("grep", "inet addr")
	grep.Stdin, _ = cmd.StdoutPipe()

	cmd.Start()

	defaultData, err := grep.Output()
	if err != nil {
		logger.LogDebug("ip error: %v", err)
	}

	key := ""
	for _, char := range string(defaultData) {
		if char == ' ' && key != "inet" {
			key = ""
		} else if key == "inet addr:" {
			ipAddress += string(char)
		} else if key == "Mask:" {
			ipMask += string(char)
		} else {
			key += string(char)
		}
	}
	logger.Log("ipAddress: " + ipAddress)
	logger.Log("ipMask: " + ipMask)

	cmd = exec.Command("route", "-n")
	grep = exec.Command("grep", "UG")
	grep.Stdin, _ = cmd.StdoutPipe()

	cmd.Start()

	gatewayData, err := grep.Output()
	if err != nil {
		logger.LogDebug("gateway error: %v", err)
	}
	key = ""
	spaceCount := 0
	for _, char := range string(gatewayData) {
		if spaceCount == 9 && char != ' ' {
			gateway += string(char)
		} else if char == ' ' {
			spaceCount++
		}
		if spaceCount > 9 {
			break
		}
	}
	logger.Log("gateway: " + gateway)

	ipAddressSlice := strings.SplitN(ipAddress, ".", 4)
	ipMaskSlice := strings.SplitN(ipMask, ".", 4)
	gatewaySlice := strings.SplitN(gateway, ".", 4)

	index := 1
	for _, ipA := range ipAddressSlice {
		if index == 1 {
			networkDefaults["ipAddress1"] = string(ipA)
		} else if index == 2 {
			networkDefaults["ipAddress2"] = string(ipA)
		} else if index == 3 {
			networkDefaults["ipAddress3"] = string(ipA)
		} else if index == 4 {
			networkDefaults["ipAddress4"] = string(ipA)
		}
		index++
	}

	index = 1
	for _, ipM := range ipMaskSlice {
		if index == 1 {
			networkDefaults["ipMask1"] = string(ipM)
		} else if index == 2 {
			networkDefaults["ipMask2"] = string(ipM)
		} else if index == 3 {
			networkDefaults["ipMask3"] = string(ipM)
		} else if index == 4 {
			networkDefaults["ipMask4"] = string(ipM)
		}
		index++
	}

	index = 1
	for _, gate := range gatewaySlice {
		if index == 1 {
			networkDefaults["gateway1"] = string(gate)
		} else if index == 2 {
			networkDefaults["gateway2"] = string(gate)
		} else if index == 3 {
			networkDefaults["gateway3"] = string(gate)
		} else if index == 4 {
			networkDefaults["gateway4"] = string(gate)
		}
		index++
	}

	return networkDefaults
}

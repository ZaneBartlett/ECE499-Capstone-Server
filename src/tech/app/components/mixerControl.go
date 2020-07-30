package components

import (
	"encoding/json"
	"os/exec"
	"strconv"
	"tech/app/logger"
	"tech/mixer/config"
)

const (
	mixerControlName = "mixerControl"
)

// MixerControl -
type MixerControl struct {
	MixerComponent

	NfcMode         bool
	UserStatusCode  int
	MixerStatusCode int
	NfcStatusCode   int
}

// NewMixerControl -
func NewMixerControl(cfg *config.CfgService) *MixerControl {

	mxr := &MixerControl{}
	mxr.Name = mixerControlName
	mxr.UserStatusCode = 0
	mxr.MixerStatusCode = 0
	mxr.ConfigService = cfg

	cfg.Register(mxr.Name, mxr.createTable)

	return mxr
}

// Action -
func (mxr *MixerControl) Action(action string, data []byte) (response []byte, err error) {

	var mapData map[string]interface{}
	mapData, err = config.JsonToMap(data)
	if err != nil {
		logger.Log("Failed to unmarshall data on '%s'", mxr.Name)
		return
	}

	switch action {
	case "GetDrinkOptions":
		response, err = mxr.getDrinkOptions()

	case "SetDrinkOptions":
		response, err = mxr.setDrinkOptions(data)

	case "InitMixing":
		response, err = mxr.initMixing(mapData)

	case "GetStatus":
		response, err = mxr.getStatus()

	case "ReadNfc":
		response, err = mxr.readNFC()

	default:
		logger.Log("Unrecognized action received in '%s': '%s'", mxr.Name, action)
	}

	return
}

// Start -
func (mxr *MixerControl) Start() error {
	return nil
}

// Stop -
func (mxr *MixerControl) Stop() error {
	return nil
}

// createTable -
func (mxr *MixerControl) createTable(cfg *config.CfgService) (err error) {
	controlSchema := []string{
		"drink0 TEXT",
		"drink1 TEXT",
		"drink2 TEXT",
		"drink3 TEXT",
		"drink4 TEXT",
		"drink5 TEXT",
		"mix INTEGER"}

	controlDefault := map[string]string{
		"drink0": "'Crown Royal'",
		"drink1": "'Agave Syrup'",
		"drink2": "'Coca-Cola'",
		"drink3": "'Tofino Gin'",
		"drink4": "'Soda Water'",
		"drink5": "'Lemon Juice'",
		"mix":    "1"}

	err = cfg.CreateTable(mxr.Name, controlSchema)
	if err != nil {
		return err
	}

	err = cfg.InitTable(mxr.Name, controlDefault)
	if err != nil {
		return err
	}

	return nil
}

func (mxr *MixerControl) getDrinkOptions() ([]byte, error) {

	drinks, err := mxr.ConfigService.Get(mxr.Name)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(drinks, "", "\t")
}

func (mxr *MixerControl) setDrinkOptions(data []byte) ([]byte, error) {

	drinks, err := mxr.ConfigService.Set(mxr.Name, data)
	if err != nil {
		return nil, err
	}

	return drinks, err
}

func (mxr *MixerControl) getStatus() ([]byte, error) {
	controlDefault := map[string]interface{}{
		"userStatus":  mxr.UserStatusCode,
		"mixerStatus": mxr.MixerStatusCode,
		"nfcStatus":   mxr.NfcStatusCode}

	return json.MarshalIndent(controlDefault, "", "\t")
}

func (mxr *MixerControl) readNFC() ([]byte, error) {
	mxr.UserStatusCode = 3
	mxr.NfcStatusCode = 1
	networkData, err := mxr.ConfigService.Get("factory")
	if err != nil {
		mxr.NfcMode = false
	} else {
		networkMap, _ := config.JsonToMap(networkData)
		mxr.NfcMode, _ = config.JSONbool(networkMap["nfcMode"])
	}

	out, err := exec.Command("python3", "./scripts/read_nfc.py", strconv.FormatBool(mxr.NfcMode)).Output()
	mxr.NfcStatusCode = 2
	if err != nil {
		logger.Log("nfc read error error: %v", err)
	}
	logger.LogDebug(string(out))

	mxr.NfcStatusCode = 0
	mxr.UserStatusCode = 0
	return networkData, nil
}

func (mxr *MixerControl) initMixing(data map[string]interface{}) ([]byte, error) {

	mxr.MixerStatusCode = 1
	mxr.UserStatusCode = 1
	pourAmt0 := int(data["pourAmt0"].(float64))
	pourAmt1 := int(data["pourAmt1"].(float64))
	pourAmt2 := int(data["pourAmt2"].(float64))
	pourAmt3 := int(data["pourAmt3"].(float64))
	pourAmt4 := int(data["pourAmt4"].(float64))
	pourAmt5 := int(data["pourAmt5"].(float64))
	mix, _ := config.JSONbool(data["mix"])

	if pourAmt0 != 0 {
		mxr.motorScriptCall("0", strconv.Itoa(pourAmt0))
	}
	if pourAmt1 != 0 {
		mxr.motorScriptCall("1", strconv.Itoa(pourAmt1))
	}
	if pourAmt2 != 0 {
		mxr.motorScriptCall("2", strconv.Itoa(pourAmt2))
	}
	if pourAmt3 != 0 {
		mxr.motorScriptCall("3", strconv.Itoa(pourAmt3))
	}
	if pourAmt4 != 0 {
		mxr.motorScriptCall("4", strconv.Itoa(pourAmt4))
	}
	if pourAmt5 != 0 {
		mxr.motorScriptCall("5", strconv.Itoa(pourAmt5))
	}
	if !mix {
		mxr.motorScriptCall("mix", "0")
	}
	if mxr.MixerStatusCode == 1 {
		mxr.MixerStatusCode = 0
		mxr.MixerStatusCode = 0
	}
	return nil, nil
}

func (mxr *MixerControl) motorScriptCall(target string, amount string) {
	out, err := exec.Command("python3", "./scripts/motor_control.py", target, amount).Output()
	if err != nil {
		logger.Log("motor control error: %v", err)
		mxr.MixerStatusCode = 2
		mxr.UserStatusCode = 2
	}
	logger.LogDebug(string(out))
}

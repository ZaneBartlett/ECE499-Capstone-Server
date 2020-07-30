package comms

import (
	"encoding/json"
	"fmt"
	"tech/app/components"
	"tech/app/logger"
	"tech/mixer/config"
)

const (
	userAuthName = "userAuth"
)

// UserAuth -
type UserAuth struct {
	components.MixerComponent

	username string
	password string
	isAdmin  bool
	loggedIn bool
}

// NewUserAuth -
func NewUserAuth(cfg *config.CfgService) *UserAuth {

	user := &UserAuth{}
	user.Name = userAuthName
	user.ConfigService = cfg

	cfg.Register(user.Name, user.createUserTable)

	return user
}

// Action -
func (usr *UserAuth) Action(action string, data []byte) (response []byte, err error) {

	var mapData map[string]interface{}
	mapData, err = config.JsonToMap(data)
	if err != nil {
		logger.Log("Failed to unmarshall data on '%s'", usr.Name)
		return
	}

	switch action {
	case "Login":
		response, err = usr.login(mapData["username"].(string), mapData["password"].(string))

	case "UpdatePassword":
		response, err = usr.passwordChange(mapData["username"].(string), mapData["currentPassword"].(string), mapData["newPassword"].(string))

	case "Logout":
		response, err = usr.logout(mapData["username"].(string))

	case "GetPaymentInfo":
		response, err = usr.GetPaymentInfo(mapData["username"].(string))

	case "SetPaymentInfo":
		response, err = usr.SetPaymentInfo(mapData)

	default:
		logger.Log("Unrecognized action received in '%s': '%s'", usr.Name, action)
	}

	return
}

// Start -
func (usr *UserAuth) Start() error {
	return nil
}

// Stop -
func (usr *UserAuth) Stop() error {
	return nil
}

// createUserTable - Generates the tech (rowId=0) and user (rowId=1) rows in userInfo table
func (usr *UserAuth) createUserTable(cfg *config.CfgService) (err error) {
	userSchema := []string{
		"username TEXT",
		"password TEXT",
		"isAdmin INTEGER",
		"loggedIn INTEGER",
		"ccNumber INTEGER",
		"ccExpiryMonth INTEGER",
		"ccExpiryYear INTEGER",
		"cvv INTEGER",
		"cardName TEXT"}

	userDefaultAdmin := map[string]string{
		"username":      "'admin'",
		"password":      "'admin'",
		"isAdmin":       "1",
		"loggedIn":      "0",
		"ccNumber":      "",
		"ccExpiryMonth": "0",
		"ccExpiryYear":  "0",
		"cvv":           "0",
		"cardName":      ""}

	userDefault := map[string]string{
		"username":      "'user'",
		"password":      "'user'",
		"isAdmin":       "0",
		"loggedIn":      "0",
		"ccNumber":      "",
		"ccExpiryMonth": "0",
		"ccExpiryYear":  "0",
		"cvv":           "0",
		"cardName":      ""}

	err = cfg.CreateTable(usr.Name, userSchema)
	if err != nil {
		return err
	}

	err = cfg.InitUser(usr.Name, userDefaultAdmin, 0)
	if err != nil {
		return err
	}

	err = cfg.InitUser(usr.Name, userDefault, 1)
	if err != nil {
		return nil
	}

	return nil
}

func (usr *UserAuth) login(username string, password string) ([]byte, error) {

	user, err := usr.ConfigService.GetUser(usr.Name, username)
	if err != nil {
		return nil, err
	}

	// Check for invalid username and password
	if user["username"] != nil && user["password"] != nil {
		if user["password"].(string) != password {
			return nil, fmt.Errorf("invalid password")
		}
	} else {
		user["username"] = ""
		user["password"] = ""
	}

	if user["loggedIn"] != 1 && user["username"] != "" {
		user["loggedIn"] = 1
		usr.ConfigService.SetUserValue(usr.Name, username, "loggedIn", int64(1))
	}

	return json.MarshalIndent(user, "", "\t")
}

func (usr *UserAuth) logout(username string) ([]byte, error) {
	err := usr.ConfigService.SetUserValue(usr.Name, username, "loggedIn", int64(0))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (usr *UserAuth) passwordChange(username string, oldPassword string, newPassword string) ([]byte, error) {
	user, err := usr.ConfigService.GetUser(usr.Name, username)
	if err != nil || oldPassword != user["password"].(string) {
		return nil, err
	}

	err = usr.ConfigService.SetUserValue(usr.Name, username, "password", newPassword)
	if err != nil {
		return nil, err
	}

	user["password"] = newPassword

	return json.MarshalIndent(user, "", "\t")
}

func (usr *UserAuth) GetPaymentInfo(username string) ([]byte, error) {
	response, err := usr.ConfigService.GetUser(usr.Name, username)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(response, "", "\t")
}

func (usr *UserAuth) SetPaymentInfo(data map[string]interface{}) ([]byte, error) {
	err := usr.ConfigService.SetUserValue(usr.Name, data["username"].(string), "ccNumber", data["ccNumber"])
	err = usr.ConfigService.SetUserValue(usr.Name, data["username"].(string), "ccExpiryMonth", data["ccExpiryMonth"])
	err = usr.ConfigService.SetUserValue(usr.Name, data["username"].(string), "ccExpiryYear", data["ccExpiryYear"])
	err = usr.ConfigService.SetUserValue(usr.Name, data["username"].(string), "cvv", data["cvv"])
	err = usr.ConfigService.SetUserValue(usr.Name, data["username"].(string), "cardName", data["cardName"])
	if err != nil {
		return nil, err
	}
	return nil, nil
}

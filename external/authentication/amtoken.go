package authentication

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

type AccountManagementAuthenticationTokenResponse struct {
	Data []AccountManagementAuthenticationTokenStruct `json:"data"`
}

type AccountManagementAuthenticationTokenStruct struct {
	Type        string `json:"type"`
	AccountName string `json:"account_name"`
	AccountId   string `json:"account_id"`
	Expires     string `json:"expires"`
	Token       string `json:"token"`
}

func SaveAccountManagementAuthenticationToken(response AccountManagementAuthenticationTokenStruct) {
	accountManagementAuthenticationTokenPath := getAccountManagementAuthenticationTokenPath()

	jsonToken, err := json.Marshal(response)

	if err != nil {
		log.Warnf("Could not convert token to JSON  %v", err)
	} else {
		err := os.WriteFile(accountManagementAuthenticationTokenPath, jsonToken, 0600)

		if err != nil {
			log.Warnf("Could not save token %s, error: %v", accountManagementAuthenticationTokenPath, err)
		} else {
			log.Debugf("Saved token to %s", accountManagementAuthenticationTokenPath)
		}
	}
}

func GetAccountManagementAuthenticationToken() *AccountManagementAuthenticationTokenStruct {

	accountManagementAuthenticationTokenPath := getAccountManagementAuthenticationTokenPath()
	data, err := os.ReadFile(accountManagementAuthenticationTokenPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Warnf("Could not read %s, error %s", accountManagementAuthenticationTokenPath, err)
		} else {
			log.Tracef("No saved api token %s", accountManagementAuthenticationTokenPath)
		}
		data = []byte{}
	} else {
		accountManagementAuthenticationToken := AccountManagementAuthenticationTokenStruct{}
		err = json.Unmarshal(data, &accountManagementAuthenticationToken)
		if err != nil {
			log.Debugf("Could not unmarshall existing file %s, error %s", data, err)
		} else {
			return &accountManagementAuthenticationToken
		}
	}

	return nil
}

func ClearAccountManagementAuthenticationToken() error {
	err := os.Remove(getAccountManagementAuthenticationTokenPath())
	if os.IsNotExist(err) {
		return nil
	}
	return nil
}

func IsAccountManagementAuthenticationTokenSet() bool {
	_, err := os.Stat(getAccountManagementAuthenticationTokenPath())

	if os.IsNotExist(err) {
		return false
	}

	return true
}

func getAccountManagementAuthenticationTokenPath() string {
	return filepath.Clean(GetAuthenticationCacheDirectory() + "/account_management_authentication_token.json")
}

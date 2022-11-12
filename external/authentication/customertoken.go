package authentication

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

type CustomerTokenResponse struct {
	Data           CustomerTokenStruct                `json:"data"`
	AdditionalInfo CustomerTokenEpccCliAdditionalInfo `json:"additional_data"`
}

type CustomerTokenStruct struct {
	Type       string `json:"type"`
	Id         string `json:"id"`
	CustomerId string `json:"customer_id"`
	Expires    int64  `json:"expires"`
	Token      string `json:"token"`
}

type CustomerTokenEpccCliAdditionalInfo struct {
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email"`
}

func SaveCustomerToken(response CustomerTokenResponse) {
	custTokenPath := getCustomerTokenPath()

	jsonToken, err := json.Marshal(response)

	if err != nil {
		log.Warnf("Could not convert token to JSON  %v", err)
	} else {
		err := os.WriteFile(custTokenPath, jsonToken, 0600)

		if err != nil {
			log.Warnf("Could not save token %s, error: %v", custTokenPath, err)
		} else {
			log.Debugf("Saved token to %s", custTokenPath)
		}
	}
}

func GetCustomerToken() *CustomerTokenResponse {

	customerTokenPath := getCustomerTokenPath()
	data, err := os.ReadFile(customerTokenPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Warnf("Could not read %s, error %s", customerTokenPath, err)
		} else {
			log.Tracef("No saved api token %s", customerTokenPath)
		}
		data = []byte{}
	} else {
		customerTokenResponse := CustomerTokenResponse{}
		err = json.Unmarshal(data, &customerTokenResponse)
		if err != nil {
			log.Debugf("Could not unmarshall existing file %s, error %s", data, err)
		} else {
			return &customerTokenResponse
		}
	}

	return nil
}

func ClearCustomerToken() error {
	err := os.Remove(getCustomerTokenPath())
	if os.IsNotExist(err) {
		return nil
	}
	return nil
}

func IsCustomerTokenSet() bool {
	_, err := os.Stat(getCustomerTokenPath())

	if os.IsNotExist(err) {
		return false
	}

	return true
}

func getCustomerTokenPath() string {
	return filepath.Clean(GetAuthenticationCacheDirectory() + "/customer_token.json")
}

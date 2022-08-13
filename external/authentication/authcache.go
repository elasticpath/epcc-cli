package authentication

import (
	"encoding/json"
	"github.com/elasticpath/epcc-cli/external/profiles"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

func GetApiToken() *ApiTokenResponse {
	apiTokenPath := getApiTokenPath()
	data, err := os.ReadFile(apiTokenPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Warnf("Could not read %s, error %s", apiTokenPath, err)
		} else {
			log.Debugf("No saved api token %s, logging in again", apiTokenPath)
		}
		data = []byte{}
	} else {
		savedApiToken := ApiTokenResponse{}
		err = json.Unmarshal(data, &savedApiToken)
		if err != nil {
			log.Debugf("Could not unmarshall existing file %s, error %s", data, err)
		} else {
			return &savedApiToken
		}
	}

	return nil
}

func SaveApiToken(bearerToken *ApiTokenResponse) {
	jsonToken, err := json.Marshal(bearerToken)

	apiTokenPath := getApiTokenPath()
	if err != nil {
		log.Warnf("Could not convert token to JSON  %v", err)
	} else {
		err = os.WriteFile(apiTokenPath, jsonToken, 0600)

		if err != nil {
			log.Warnf("Could not save token %s, error: %v", apiTokenPath, err)
		} else {
			log.Debugf("Saved token to %s", apiTokenPath)
		}
	}
}

func ClearApiToken() error {
	err := os.Remove(getApiTokenPath())
	if os.IsNotExist(err) {
		return nil
	}
	return nil
}

func IsApiTokenSet() bool {
	_, err := os.Stat(getApiTokenPath())

	if os.IsNotExist(err) {
		return false
	}

	return true
}

func getApiTokenPath() string {
	apiTokenPath := filepath.Clean(GetAuthenticationCacheDirectory() + "/bearer.json")
	return apiTokenPath
}

func GetAuthenticationCacheDirectory() string {
	authenticationCacheDirectory := filepath.Clean(filepath.FromSlash(profiles.GetProfileDataDirectory() + "/auth_cache/"))
	//built in check if dir exists
	if err := os.MkdirAll(authenticationCacheDirectory, 0700); err != nil {
		log.Errorf("could not make directory")
	}

	return authenticationCacheDirectory
}

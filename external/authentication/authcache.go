package authentication

import (
	"github.com/elasticpath/epcc-cli/external/profiles"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

func getApiTokenPath() string {
	apiTokenPath := filepath.Clean(getAuthenticationCacheDirectory() + "/bearer.json")
	return apiTokenPath
}

func ClearApiToken() error {
	err := os.Remove(getApiTokenPath())
	if os.IsNotExist(err) {
		return nil
	}
	return nil
}

func getAuthenticationCacheDirectory() string {
	authenticationCacheDirectory := filepath.Clean(filepath.FromSlash(profiles.GetProfileDataDirectory() + "/api_authentication/"))
	//built in check if dir exists
	if err := os.MkdirAll(authenticationCacheDirectory, 0700); err != nil {
		log.Errorf("could not make directory")
	}

	return authenticationCacheDirectory
}

package profiles

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/elasticpath/epcc-cli/config"
	log "github.com/sirupsen/logrus"
	"net/url"
	"os"
)

func GetProfileName() string {
	if config.Envs.EPCC_PROFILE != "" {
		log.Tracef("Using EPCC_PROFILE value for profile %s", config.Envs.EPCC_PROFILE)
		return config.Envs.EPCC_PROFILE
	} else {
		u, err := url.Parse(config.Envs.EPCC_API_BASE_URL)
		profileName := ""
		if err != nil {
			result := newSHA256([]byte(config.Envs.EPCC_CLIENT_ID + ":" + config.Envs.EPCC_API_BASE_URL))
			profileName = hex.EncodeToString(result)
		} else {
			profileName = fmt.Sprintf("%s-%s", u.Host, config.Envs.EPCC_CLIENT_ID)
		}

		log.Tracef("Using auto generated profile name %s", profileName)

		return profileName
	}
}

func GetProfileDirectory() string {
	profileDir := GetProfileDataBaseURL() + GetProfileName()

	log.Tracef("Creating profile directory %s", profileDir)
	if err := os.MkdirAll(profileDir, 0700); err != nil {
		panic(fmt.Sprintf("Could not create home directory %v", err))
	}

	return profileDir
}

func GetProfileDataBaseURL() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("Could not get user hoem directory home directory %v", err))
	}

	return homeDir + "/.epcc/profiles_data/"
}

// NewSHA256 ...
func newSHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

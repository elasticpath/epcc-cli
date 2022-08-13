package authentication

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func IsAutoLoginEnabled() bool {
	_, err := os.Stat(getDisableAutoLoginFile())

	if err != nil {
		if !os.IsNotExist(err) {
			log.Warnf("Got error while reading file, %v", err)
		}
		return true
	}

	return false
}
func getDisableAutoLoginFile() string {
	return GetAuthenticationCacheDirectory() + "/disable_auto_login"
}

func DisableAutoLogin() error {
	return os.WriteFile(getDisableAutoLoginFile(), []byte{}, 0700)
}

func EnableAutoLogin() error {
	err := os.Remove(getDisableAutoLoginFile())

	if os.IsNotExist(err) {
		return nil
	}

	return err
}

package version

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/profiles"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

func CheckVersionChangeAndLogWarning() {
	lastVersionFilename := filepath.Clean(profiles.GetProfileDataDirectory() + "/last_version.txt")

	lastVersionBytes, err := os.ReadFile(lastVersionFilename)
	if err != nil && !os.IsNotExist(err) {
		log.Warnf("Couldn't read file: %s: %v", lastVersionFilename, err)

		return
	}

	lastVersion := strings.Trim(string(lastVersionBytes), "\r\n ")

	currentVersion := strings.Trim(fmt.Sprintf("%s (Commit %s)", Version, Commit), "\r\n ")

	if lastVersion == currentVersion {
		return
	}

	err = os.WriteFile(lastVersionFilename, []byte(currentVersion), 0600)

	if err != nil {
		log.Warnf("Couldn't write file: %s: %v", lastVersionFilename, err)
	}

	if lastVersion == "" {
		return
	}

	log.Infof("Version of epcc has changed %s => %s", lastVersion, currentVersion)
}

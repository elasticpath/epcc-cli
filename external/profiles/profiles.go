package profiles

import (
	"github.com/elasticpath/epcc-cli/config"
	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
	"os"
	"path/filepath"
)

//profile name is set to config.Profile in InitConfig

var ProfileName = "default"

func GetProfileDirectory() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Errorf("could not get home directory")
		os.Exit(1)
	}
	profileDirectory := home + "/.epcc/"
	profileDirectory = filepath.FromSlash(profileDirectory)
	//built in check if dir exists
	if err = os.MkdirAll(profileDirectory, 0700); err != nil {
		log.Errorf("could not make directory")
	}

	return filepath.Clean(profileDirectory)
}

func GetProfileDataDirectory() string {
	profileDirectory := GetProfileDirectory()
	profileDataDirectory := filepath.Clean(filepath.FromSlash(profileDirectory + "/" + ProfileName + "/data"))
	//built in check if dir exists
	if err := os.MkdirAll(profileDataDirectory, 0700); err != nil {
		log.Errorf("could not make directory")
	}

	return profileDataDirectory
}

func GetConfigFilePath() string {
	configPath := GetProfileDirectory()
	configPath = filepath.Clean(filepath.FromSlash(configPath + "/../config"))
	if _, err := os.Stat(configPath); err != nil {
		log.Trace("could not find file at " + configPath)
		file, err := os.Create(configPath)
		defer file.Close()
		if err != nil {
			log.Errorf("could not create file at " + configPath)
		}
		log.Trace("creating config file at " + configPath)
	}

	return configPath
}

func GetProfile(name string) *config.Env {
	result := &config.Env{}
	configPath := GetConfigFilePath()
	cfg, err := ini.Load(configPath)
	if err != nil {
		log.Debug("could not load file at " + configPath)
		return result
	}

	if !cfg.HasSection(name) {
		log.Debug("could not find profile in file")
		return result
	}

	err = cfg.Section(name).MapTo(result)
	if err != nil {
		log.Debug("could not load file at " + configPath)
	}

	return result

}

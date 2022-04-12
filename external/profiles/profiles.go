package profiles

import (
	"github.com/elasticpath/epcc-cli/config"
	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
	"os"
	"path/filepath"
)

//profile name is set to config.Profile in InitConfig

func getProfileDirectory() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Errorf("could not get home directory")
		os.Exit(1)
	}
	configDir := home + "/.epcc/profiles_data"
	configDir = filepath.FromSlash(configDir)
	//built in check if dir exists
	if err = os.MkdirAll(configDir, 0700); err != nil {
		log.Errorf("could not make directory")
	}

	return configDir
}

func GetProfilePath() string {
	configPath := getProfileDirectory()
	configPath = filepath.FromSlash(configPath + "/config")
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
	result := config.Env{}
	configPath := GetProfilePath()
	cfg, err := ini.Load(configPath)
	if err != nil {
		log.Errorf("could not load file at " + configPath)
		os.Exit(1)
	}
	if !cfg.HasSection(name) {
		log.Errorf("could not find profile in file")
		os.Exit(1)
	}
	cfg.Section(name).MapTo(&result)
	return &result

}

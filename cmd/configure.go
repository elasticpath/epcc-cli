package cmd

import (
	"bufio"
	"github.com/elasticpath/epcc-cli/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
	"os"
	"path/filepath"
	"strings"
)

var configure = &cobra.Command{
	Use:   "configure",
	Short: "Creates a profile by prompting for input over the command line.",
	Long:  "Will first prompt for a name then a series of variable specific for the user being created",
	Run: func(cmd *cobra.Command, args []string) {

			configPath := GetProfilePath()
			cfg, err := ini.Load(configPath)
			if err != nil{
				log.Errorf("error loading to file " +configPath)
				os.Exit(1)
			}
			newProfile := config.Env{}
			reader := bufio.NewReader(os.Stdin)
			println("Create new Profile")
			print("Profile Name:")
			text := readInput(reader)
			print("Base URL:")
			newProfile.EPCC_API_BASE_URL = readInput(reader)
			print("Client ID:")
			newProfile.EPCC_CLIENT_ID = readInput(reader)
			print("Client Secret:")
			newProfile.EPCC_CLIENT_SECRET = readInput(reader)
			print("Beta Features:")
			newProfile.EPCC_BETA_API_FEATURES= readInput(reader)

			section, err :=cfg.NewSection(text)
			section.ReflectFrom(&newProfile)
			cfg.SaveTo(configPath)
			if err!= nil{
				log.Errorf("error writing to file " +configPath)
				os.Exit(1)
			}
			config.Envs = &newProfile
			config.Profile = text

	},
}
func getProfileDirectory() string{
	home, err := os.UserHomeDir()
	if err != nil {
		log.Errorf("could not get home directory")
		os.Exit(1)
	}
	configDir := home +"/.epcc/profiles_data"
	configDir = filepath.FromSlash(configDir)
	//built in check if dir exists
	if err = os.MkdirAll(configDir, 0700); err != nil{
		log.Errorf("could not make directory")
	}

	return configDir
}
func GetProfilePath() string{
	configPath := getProfileDirectory()
	configPath = filepath.FromSlash(configPath + "/config")
	if _, err := os.Stat(configPath); err != nil  {
		log.Trace("could not find file at " + configPath)
		file , err := os.Create(configPath)
		defer file.Close()
		if err != nil{
			log.Errorf("could not create file at " + configPath)
		}
		log.Trace("creating config file at " + configPath)
	}

	return configPath
}


func readInput(reader *bufio.Reader) string{
response, _ := reader.ReadString('\n')
return strings.TrimSuffix(response, "\n")
}
func GetProfile(name string) *config.Env{
	result := config.Env{}
	configPath := GetProfilePath()
	cfg, err := ini.Load(configPath)
	if err != nil{
		log.Errorf("could not load file at " + configPath)
		os.Exit(1)
	}
	if !cfg.HasSection(name){
		log.Errorf("could not find profile in file")
		os.Exit(1)
	}
	cfg.Section(name).MapTo(&result)
	return &result

}

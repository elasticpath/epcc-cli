package cmd

import (
	"bufio"
	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/profiles"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
	"os"
	"strings"
)

var configure = &cobra.Command{
	Use:   "configure",
	Short: "Creates a profile by prompting for input over the command line.",
	Long:  "Will first prompt for a name then a series of variable specific for the user being created",
	Run: func(cmd *cobra.Command, args []string) {

		configPath := profiles.GetProfilePath()
		cfg, err := ini.Load(configPath)
		if err != nil {
			log.Errorf("error loading to file " + configPath)
			os.Exit(1)
		}
		newProfile := config.Env{}
		reader := bufio.NewReader(os.Stdin)
		println("Create new Profile")
		print("Profile Name:")
		text := readInput(reader)
		print("API Base URL [https://api.moltin.com]:")
		if input := readInput(reader); input != "" {
			newProfile.EPCC_API_BASE_URL = input
		} else {
			newProfile.EPCC_API_BASE_URL = "https://api.moltin.com"
		}
		newProfile.EPCC_API_BASE_URL = readInput(reader)
		print("Client ID [None]:")
		newProfile.EPCC_CLIENT_ID = readInput(reader)
		print("Client Secret [None]:")
		newProfile.EPCC_CLIENT_SECRET = readInput(reader)
		print("(https://documentation.elasticpath.com/commerce-cloud/docs/api/basics/api-contract.html#beta-apis) [None]:")
		newProfile.EPCC_BETA_API_FEATURES = readInput(reader)

		section, err := cfg.NewSection(text)
		section.ReflectFrom(&newProfile)
		cfg.SaveTo(configPath)
		if err != nil {
			log.Errorf("error writing to file " + configPath)
			os.Exit(1)
		}
		config.Envs = &newProfile
		config.Profile = text

	},
}

func readInput(reader *bufio.Reader) string {
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Errorf("error reading from stdin %s", err.Error())
		os.Exit(1)
	}
	response = strings.TrimSuffix(response, "\n")
	return response
}

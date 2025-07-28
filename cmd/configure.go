package cmd

import (
	"bufio"
	"fmt"
	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/profiles"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
	"os"
	"strconv"
	"strings"
)

var configure = &cobra.Command{
	Use:   "configure",
	Short: "Creates a profile by prompting for input over the command line.",
	Long:  "Will first prompt for a name then a series of variable specific for the user being created",
	Run: func(cmd *cobra.Command, args []string) {
		configPath := profiles.GetConfigFilePath()
		cfg, err := ini.Load(configPath)
		if err != nil {
			log.Errorf("error loading to file %", configPath)
			os.Exit(1)
		}
		newProfile := config.Env{}

		reader := bufio.NewReader(os.Stdin)
		println("Create new Profile")
		var text = ""
		profileName := "default"
		if defaultVal, ok := os.LookupEnv("EPCC_PROFILE"); ok {
			profileName = defaultVal
		}

		fmt.Printf("Profile Name[%s]:", profileName)
		text = readInput(reader)
		if text != "" {
			profileName = text
		}

		print("API Base URL [https://euwest.api.elasticpath.com]:")
		if input := readInput(reader); input != "" {
			newProfile.EPCC_API_BASE_URL = input
		} else {
			newProfile.EPCC_API_BASE_URL = "https://euwest.api.elasticpath.com"
		}
		print("Client ID [None]:")
		newProfile.EPCC_CLIENT_ID = readInput(reader)
		print("Client Secret [None]:")
		newProfile.EPCC_CLIENT_SECRET = readInput(reader)
		print("Beta Features Enabled (See: https://elasticpath.dev/guides/Getting-Started/api-contract#beta-apis) [None]:")
		newProfile.EPCC_BETA_API_FEATURES = readInput(reader)

		print("Rate Limit [10]:")

		if input := readInput(reader); input != "" {
			rateLimit, err := strconv.Atoi(input)

			if err != nil {
				log.Errorf("Invalid rate limit %s, error: %v", input, err)
				os.Exit(2)
			}
			newProfile.EPCC_RATE_LIMIT = uint16(rateLimit)
		} else {
			newProfile.EPCC_RATE_LIMIT = 10
		}

		section, err := cfg.NewSection(profileName)
		if err != nil {
			log.Errorf("error creating section, error: %v", err)
			os.Exit(3)
		}
		section.ReflectFrom(&newProfile)
		cfg.SaveTo(configPath)
		if err != nil {
			log.Errorf("error writing to file %s, error: %v", configPath, err)
			os.Exit(1)
		}
		config.SetEnv(&newProfile)
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

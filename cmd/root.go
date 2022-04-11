package cmd

import (
	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/logger"
	"github.com/elasticpath/epcc-cli/globals"
	log "github.com/sirupsen/logrus"
	"github.com/thediveo/enumflag"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/elasticpath/epcc-cli/external/json"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	cobra.OnInitialize(initConfig)

	if err := env.Parse(config.Envs); err != nil {
		panic("Could not parse environment variables")
	}

	rootCmd.AddCommand(
		cmCommand,
		docsCommand,
		testJson,
		get,
		create,
		delete,
		update,
		logs,
		resourceListCommand,
	)

	testJson.Flags().BoolVarP(&noWrapping, "no-wrapping", "", false, "if set, we won't wrap the output the json in a data tag")
	testJson.Flags().BoolVarP(&compliant, "compliant", "", false, "if set, we wrap most keys in an attributes tage automatically.")

	rootCmd.PersistentFlags().Var(
		enumflag.New(&logger.Loglevel, "log", logger.LoglevelIds, enumflag.EnumCaseInsensitive),
		"log",
		"sets logging level; can be 'trace', 'debug', 'info', 'warn', 'error', 'fatal', 'panic'")
	rootCmd.PersistentFlags().BoolVarP(&json.MonochromeOutput, "monochrome-output", "M", false, "By default, epcc will output using colors if the terminal supports this. Use this option to disable it.")
	rootCmd.PersistentFlags().StringSliceVarP(&globals.RawHeaders, "header", "H", []string{}, "Extra headers and values to include in the request when sending HTTP to a server. You may specify any number of extra headers.")
}

var rootCmd = &cobra.Command{
	Use:   os.Args[0],
	Short: "A command line interface for interacting with the Elastic Path Commerce Cloud API",
	Long: `The EPCC CLI tool provides a powerful command line interface for interacting with the Elastic Path Commerce Cloud API.

The EPCC CLI tool uses environment variables for configuration and in particular a tool like https://direnv.net/ which
auto populates your shell with environment variables when you switch directories. This allows you to store a context in a folder,
and come back to it at any time.

Environment Variables

- EPCC_API_BASE_URL - The API endpoint that we will hit
- EPCC_CLIENT_ID - The client id (available in Commerce Manager)
- EPCC_CLIENT_SECRET - The client secret (available in Commerce Manager)
- EPCC_BETA_API_FEATURES - Beta features in the API we want to enable.
`,
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		log.SetLevel(logger.Loglevel)
	},
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Errorf("Error occured while processing command %s", err)
		os.Exit(1)
	}
}

func initConfig() {
	// Don't forget to read config either from cfgFile or from home directory!
	cfgFile := ""
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Errorf("Error %s", err)
			os.Exit(1)
		}

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".epcc")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
		} else {
			log.Errorf("Can't read config %s", err)
			os.Exit(1)
		}

	}
}

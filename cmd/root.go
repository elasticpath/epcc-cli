package cmd

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/json"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var Envs = &config.Env{}

func init() {
	cobra.OnInitialize(initConfig)

	if err := env.Parse(Envs); err != nil {
		panic("Could not parse environment variables")
	}

	rootCmd.AddCommand(
		cmCommand,
		docsCommand,
		testJson,
		get,
	)

	testJson.Flags().BoolVarP(&noWrapping, "no-wrapping", "", false, "if set, we won't wrap the output the json in a data tag")
	rootCmd.PersistentFlags().BoolVarP(&json.MonochromeOutput, "monochrome-output", "M", false, "By default, epcc will output using colors if the terminal supports this. Use this option to disable it.")
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
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
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
			fmt.Println(err)
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
			fmt.Println("Can't read config:", err)
			os.Exit(1)
		}

	}
}

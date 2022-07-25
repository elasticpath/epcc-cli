package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/logger"
	"github.com/elasticpath/epcc-cli/external/profiles"
	"github.com/elasticpath/epcc-cli/external/version"
	"github.com/elasticpath/epcc-cli/globals"
	log "github.com/sirupsen/logrus"
	"github.com/thediveo/enumflag"
	"golang.org/x/time/rate"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/spf13/cobra"
)

var rateLimit uint16

func init() {
	cobra.OnInitialize(initConfig)

	if err := env.Parse(config.Envs); err != nil {
		panic("Could not parse environment variables")
	}

	RootCmd.AddCommand(
		cmCommand,
		docsCommand,
		testJson,
		get,
		create,
		update,
		delete,
		DeleteAll,
		Logs,
		resourceListCommand,
		aliasesCmd,
		configure,
		login,
		logout,
		ResetStore,
	)
	Logs.AddCommand(LogsList, LogsShow, LogsClear)

	testJson.Flags().BoolVarP(&noWrapping, "no-wrapping", "", false, "if set, we won't wrap the output the json in a data tag")
	testJson.Flags().BoolVarP(&compliant, "compliant", "", false, "if set, we wrap most keys in an attributes tage automatically.")

	RootCmd.PersistentFlags().Var(
		enumflag.New(&logger.Loglevel, "log", logger.LoglevelIds, enumflag.EnumCaseInsensitive),
		"log",
		"sets logging level; can be 'trace', 'debug', 'info', 'warn', 'error', 'fatal', 'panic'")

	RootCmd.PersistentFlags().BoolVarP(&json.MonochromeOutput, "monochrome-output", "M", false, "By default, epcc will output using colors if the terminal supports this. Use this option to disable it.")
	RootCmd.PersistentFlags().StringSliceVarP(&globals.RawHeaders, "header", "H", []string{}, "Extra headers and values to include in the request when sending HTTP to a server. You may specify any number of extra headers.")
	RootCmd.PersistentFlags().StringVarP(&profiles.ProfileName, "profile", "P", "default", "overrides the current EPCC_PROFILE var to run the command with the chosen profile.")
	RootCmd.PersistentFlags().Uint16VarP(&rateLimit, "rate-limit", "", 10, "Request limit per second")

	aliasesCmd.AddCommand(aliasListCmd, aliasClearCmd)

}

var persistentPreRunFuncs []func(cmd *cobra.Command, args []string) error

func AddRootPreRunFunc(f func(cmd *cobra.Command, args []string) error) {
	persistentPreRunFuncs = append(persistentPreRunFuncs, f)
}

var RootCmd = &cobra.Command{
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
- EPCC_CLI_HTTP_HEADER_[0,1,...] - An additional HTTP header to set with all requests, the format should be "HeaderName: value"
- EPCC_PROFILE - The name of the profile we will use (isolates namespace, credentials, etc...)

`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		log.SetLevel(logger.Loglevel)

		if config.Envs.EPCC_RATE_LIMIT != 0 {
			rateLimit = config.Envs.EPCC_RATE_LIMIT
		}
		log.Debugf("Rate limit set to %d request per second ", rateLimit)
		httpclient.Limit = rate.NewLimiter(rate.Limit(rateLimit), 1)

		for _, runFunc := range persistentPreRunFuncs {
			err := runFunc(cmd, args)
			if err != nil {
				return err
			}
		}

		return nil
	},

	SilenceUsage: true,
	Version:      fmt.Sprintf("%s (Commit %s)", version.Version, version.Commit),
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Errorf("Error occured while processing command %s", err)
		os.Exit(1)
	}
}

func initConfig() {
	envProfileName, ok := os.LookupEnv("EPCC_PROFILE")
	if ok {
		profiles.ProfileName = envProfileName
	}
	config.Envs = profiles.GetProfile(profiles.ProfileName)

	// Override profile configuration with environment variables
	if err := env.Parse(config.Envs); err != nil {
		panic("Could not parse environment variables")
	}
}

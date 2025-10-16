package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/authentication"
	"github.com/elasticpath/epcc-cli/external/clictx"
	"github.com/elasticpath/epcc-cli/external/headergroups"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/logger"
	"github.com/elasticpath/epcc-cli/external/misc"
	"github.com/elasticpath/epcc-cli/external/profiles"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/elasticpath/epcc-cli/external/shutdown"
	"github.com/elasticpath/epcc-cli/external/version"
	log "github.com/sirupsen/logrus"
	"github.com/thediveo/enumflag"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/spf13/cobra"
)

var rateLimit uint16

var requestTimeout float32

var statisticsFrequency uint16

var jqCompletionFunc = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{
		".data.",
		".data.attributes.",
		".data.attributes.email",
		".data.attributes.id",
		".data.attributes.name",
		".data.attributes.sku",
		".data.attributes.slug",
		".data.email",
		".data.id",
		".data.name",
		".data.sku",
		".data.slug",
		".data[].attributes",
		".data[].attributes.email",
		".data[].attributes.name",
		".data[].attributes.sku",
		".data[].attributes.slug",
		".data[].email",
		".data[].id",
		".data[].name",
		".data[].sku",
		".data[].slug",
	}, cobra.ShellCompDirectiveNoSpace
}

var profileNameFromCommandLine = ""

func InitializeCmd() {

	DumpTraces()

	os.Args = misc.AddImplicitDoubleDash(os.Args)
	if len(os.Args) > 1 && os.Args[1] == "__complete" {
		DisableLongOutput = true
		DisableExampleOutput = true
	}

	cobra.OnInitialize(initConfig)
	initConfig()

	e := &config.Env{}
	if err := env.Parse(e); err != nil {
		log.Fatalf("Could not parse environment variables %v", err)
	}

	applyLogLevelEarlyDetectionHack()
	log.Tracef("Root Command Building In Progress")

	resources.PublicInit()
	initRunbookCommands()
	log.Tracef("Runbooks initialized")
	RootCmd.AddCommand(
		cmCommand,
		docsCommand,
		testJson,
		Logs,
		resourceListCommand,
		aliasesCmd,
		configure,
		LoginCmd,
		logoutCmd,
		ResetStore,
		runbookGlobalCmd,
	)

	log.Tracef("Building Create Commands")
	NewCreateCommand(RootCmd)
	log.Tracef("Building Delete Commands")
	NewDeleteCommand(RootCmd)
	log.Tracef("Building Get Commands")
	NewGetCommand(RootCmd)

	log.Tracef("Building Update Commands")
	NewUpdateCommand(RootCmd)

	log.Tracef("Building Delete All Commands")
	NewDeleteAllCommand(RootCmd)

	log.Tracef("Building Resource Info Commands")
	NewResourceInfoCommand(RootCmd)

	Logs.AddCommand(LogsList, LogsShow, LogsClear, LogsCurlReplay)

	LogsCurlReplay.PersistentFlags().BoolVarP(&CurlInlineAuth, "inline-auth", "", false, "If set, we will replace the authorization header with a curl call and our current credentials")

	testJson.ResetFlags()
	testJson.Flags().BoolVarP(&noWrapping, "no-wrapping", "", false, "if set, we won't wrap the output the json in a data tag")
	testJson.Flags().BoolVarP(&compliant, "compliant", "", false, "if set, we wrap most keys in an attributes tags automatically.")

	addLogLevel(RootCmd)

	RootCmd.PersistentFlags().BoolVarP(&json.MonochromeOutput, "monochrome-output", "M", false, "By default, epcc will output using colors if the terminal supports this. Use this option to disable it.")
	RootCmd.PersistentFlags().StringSliceVarP(&httpclient.RawHeaders, "header", "H", []string{}, "Extra headers and values to include in the request when sending HTTP to a server. You may specify any number of extra headers.")
	RootCmd.PersistentFlags().StringVarP(&profileNameFromCommandLine, "profile", "P", "", "overrides the current EPCC_PROFILE var to run the command with the chosen profile.")
	RootCmd.PersistentFlags().Uint16VarP(&rateLimit, "rate-limit", "", 10, "Request limit per second")
	RootCmd.PersistentFlags().BoolVarP(&httpclient.Retry5xx, "retry-5xx", "", false, "Whether we should retry requests with HTTP 5xx response code")
	RootCmd.PersistentFlags().BoolVarP(&httpclient.Retry429, "retry-429", "", false, "Whether we should retry requests with HTTP 429 response code")
	RootCmd.PersistentFlags().BoolVarP(&httpclient.RetryConnectionErrors, "retry-connection-errors", "", false, "Whether we should retry requests with connection errors")
	RootCmd.PersistentFlags().UintVarP(&httpclient.RetryDelay, "retry-delay", "", 500, "When retrying how long should we delay")
	RootCmd.PersistentFlags().BoolVarP(&httpclient.RetryAllErrors, "retry-all-errors", "", false, "When enable retries on all errors (i.e., the same as --retry-5xx --retry-429 and --retry-connection-errors")

	RootCmd.PersistentFlags().BoolVarP(&httpclient.DontLog2xxs, "silence-2xx", "", false, "Whether we should silence HTTP 2xx response code logging")

	RootCmd.PersistentFlags().Float32VarP(&requestTimeout, "timeout", "", 60, "Request timeout in seconds (fractional values allowed)")
	RootCmd.PersistentFlags().Uint16VarP(&statisticsFrequency, "statistics-frequency", "", 15, "How often to print runtime statistics (0 turns them off)")

	ResetStore.ResetFlags()
	ResetStore.PersistentFlags().BoolVarP(&DeleteApplicationKeys, "delete-application-keys", "", false, "if set, we delete application keys as well")

	aliasesCmd.AddCommand(aliasListCmd, aliasClearCmd)

	LoginCmd.AddCommand(loginClientCredentials)
	LoginCmd.AddCommand(loginImplicit)
	LoginCmd.AddCommand(loginInfo)
	LoginCmd.AddCommand(loginDocs)
	LoginCmd.AddCommand(loginCustomer)
	LoginCmd.AddCommand(loginAccountManagement)
	LoginCmd.AddCommand(loginOidc)

	loginOidc.PersistentFlags().Uint16VarP(&OidcPort, "port", "p", 8080, "The port to listen on for the OIDC callback")
	logoutCmd.AddCommand(logoutBearer)
	logoutCmd.AddCommand(logoutCustomer)
	logoutCmd.AddCommand(logoutAccountManagement)
	logoutCmd.AddCommand(LogoutHeaders)

	NewHeadersCommand(RootCmd)
	log.Tracef("Root Command Constructed")
}

// If there is a log level argument, we will set it much earlier on a dummy command
// this helps if you need to enable tracing while the root command is being built.
func applyLogLevelEarlyDetectionHack() {
	for i, arg := range os.Args {
		if arg == "--log" && i+1 < len(os.Args) {
			newCmd := &cobra.Command{
				Use: "foo",
			}
			addLogLevel(newCmd)

			newCmd.SetArgs([]string{"--log", os.Args[i+1]})

			newCmd.RunE = func(command *cobra.Command, args []string) error {
				log.SetLevel(logger.Loglevel)
				return nil
			}

			err := newCmd.Execute()
			if err != nil {
				log.Warnf("Couldn't set log level early: %v", err)
			}
			return
		}
	}
}

func addLogLevel(cmd *cobra.Command) {
	cmd.PersistentFlags().Var(
		enumflag.New(&logger.Loglevel, "log", logger.LoglevelIds, enumflag.EnumCaseInsensitive),
		"log",
		"sets logging level; can be 'trace', 'debug', 'info', 'warn', 'error', 'fatal', 'panic'")
}

var persistentPreRunFuncs []func(cmd *cobra.Command, args []string) error

func AddRootPreRunFunc(f func(cmd *cobra.Command, args []string) error) {
	persistentPreRunFuncs = append(persistentPreRunFuncs, f)
}

var RootCmd = GetRootCommand()

func GetRootCommand() *cobra.Command {
	return &cobra.Command{
		Use:   os.Args[0],
		Short: "A command line interface for interacting with the Elastic Path Composable Commerce API",
		Long: `The EPCC CLI tool provides a powerful command line interface for interacting with the Elastic Path Composable Commerce API.

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
- EPCC_CLI_DISABLE_TLS_VERIFICATION - Disables TLS verification
- EPCC_RUNBOOK_DIRECTORY - Directory to scan for additional runbooks
- EPCC_CLI_DISABLE_TEMPLATE_EXECUTION - Disables template execution (recommended if input is untrusted).
- EPCC_CLI_DISABLE_RESOURCES - A comma seperated list of resources that will be hidden in command lists
`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			log.SetLevel(logger.Loglevel)

			e := config.GetEnv()
			if e.EPCC_RATE_LIMIT != 0 {
				rateLimit = e.EPCC_RATE_LIMIT
			}
			authentication.Initialize()

			log.Debugf("Rate limit set to %d request per second, printing statistics every %d seconds ", rateLimit, statisticsFrequency)

			httpclient.Initialize(rateLimit, requestTimeout, int(statisticsFrequency))

			for _, runFunc := range persistentPreRunFuncs {
				err := runFunc(cmd, args)
				if err != nil {
					return err
				}
			}

			version.CheckVersionChangeAndLogWarning()

			return nil
		},

		SilenceUsage: true,
		Version:      fmt.Sprintf("%s [commit: %s, built: %s]", version.Version, version.Commit, version.Date),
	}
}

func Execute() {
	sigs := make(chan os.Signal, 1)
	normalShutdown := make(chan bool, 1)
	shutdownHandlerDone := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		exit := false
		select {
		case sig := <-sigs:
			log.Warnf("Shutting down program due to signal [%v]", sig)
			shutdown.ShutdownFlag.Store(true)
			clictx.Cancel()
			exit = true
		case <-normalShutdown:
		}

		defer func() {
			shutdownHandlerDone <- true
		}()

		go func() {
			time.Sleep(2 * time.Second)
			log.Infof("Waiting for all outstanding operations to finish")
		}()

		shutdown.OutstandingOpCounter.Wait()

		httpclient.LogStats()
		aliases.FlushAliases()
		headergroups.FlushHeaderGroups()

		if exit {
			os.Exit(3)
		}

	}()

	err := RootCmd.Execute()
	normalShutdown <- true

	<-shutdownHandlerDone

	if err != nil {
		log.Errorf("Error occurred while processing command: %s", err)

		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func DumpTraces() {
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGQUIT)
		buf := make([]byte, 1<<20)
		for {
			<-sigs
			stacklen := runtime.Stack(buf, true)
			log.Printf("=== received SIGQUIT ===\n*** goroutine dump...\n%s\n*** end\n", buf[:stacklen])
		}
	}()
}

func initConfig() {

	envProfileName, ok := os.LookupEnv("EPCC_PROFILE")
	if ok {
		profiles.SetProfileName(envProfileName)
	}

	if profileNameFromCommandLine != "" {
		profiles.SetProfileName(profileNameFromCommandLine)
	}

	e := profiles.GetProfile(profiles.GetProfileName())

	// Override profile configuration with environment variables
	if err := env.Parse(e); err != nil {
		log.Fatalf("Could not process environment variables, error: %v", err)
		panic("Could not parse environment variables")
	}

	config.SetEnv(e)
}

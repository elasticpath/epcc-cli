package cmd

import (
	"context"
	"fmt"
	"github.com/buildkite/shellwords"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/elasticpath/epcc-cli/external/runbooks"
	_ "github.com/elasticpath/epcc-cli/external/runbooks"
	"github.com/elasticpath/epcc-cli/external/shutdown"
	"github.com/jolestar/go-commons-pool/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sync/semaphore"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var runbookGlobalCmd = &cobra.Command{
	Use:          "runbooks",
	Short:        "Built in runbooks for interacting with EPCC",
	SilenceUsage: false,
	//Hidden:       false,
}

func initRunbookCommands() {
	runbooks.InitializeBuiltInRunbooks()

	runbookGlobalCmd.AddCommand(initRunbookShowCommands())
	runbookGlobalCmd.AddCommand(initRunbookRunCommands())
	runbookGlobalCmd.AddCommand(initRunbookDevCommands())
}

var AbortRunbookExecution = atomic.Bool{}

func initRunbookShowCommands() *cobra.Command {

	// epcc runbook show
	runbookShowCommand := &cobra.Command{
		Use:          "show",
		Short:        "Display the runbook contents",
		SilenceUsage: true,
	}

	for _, runbook := range runbooks.GetRunbooks() {
		// Create a copy of runbook scoped to the loop
		runbook := runbook

		// epcc runbook show <runbook_name>
		runbookShowRunbookCmd := &cobra.Command{
			Use:   runbook.Name,
			Long:  runbook.Description.Long,
			Short: runbook.Description.Short,
		}
		runbookShowCommand.AddCommand(runbookShowRunbookCmd)

		for _, runbookAction := range runbook.RunbookActions {
			// Create a copy of runbookAction scoped to the loop
			runbookAction := runbookAction

			runbookStringArguments := runbooks.CreateMapForRunbookArgumentPointers(runbookAction)

			// epcc runbook show <runbook> <action>
			runbookShowRunbookActionCmd := &cobra.Command{
				Use:   runbookAction.Name,
				Long:  runbookAction.Description.Long,
				Short: runbookAction.Description.Short,
				RunE: func(cmd *cobra.Command, args []string) error {
					for stepIdx, cmd := range runbookAction.RawCommands {
						templateName := fmt.Sprintf("Runbook: %s Action: %s Step: %d", runbook.Name, runbookAction.Name, stepIdx)

						rawCmdLines, err := runbooks.RenderTemplates(templateName, cmd, runbookStringArguments, runbookAction.Variables)
						if err != nil {
							return err
						}
						for _, line := range rawCmdLines {
							if len(strings.Trim(line, " \n")) > 0 {
								println(line)
							}

						}
					}
					return nil
				},
			}

			processRunbookVariablesOnCommand(runbookShowRunbookActionCmd, runbookStringArguments, runbookAction.Variables, false)

			runbookShowRunbookCmd.AddCommand(runbookShowRunbookActionCmd)
		}
	}

	return runbookShowCommand
}

type commandResult struct {
	error       error
	stepIdx     int
	commandIdx  int
	commandLine string
}

func initRunbookRunCommands() *cobra.Command {

	// epcc runbook run
	runbookRunCommand := &cobra.Command{
		Use:          "run",
		Aliases:      []string{"execute"},
		Short:        "Execute a runbook",
		SilenceUsage: true,
	}

	execTimeoutInSeconds := runbookRunCommand.PersistentFlags().Int64("execution-timeout", 900, "How long should the script take to execute before timing out")
	maxConcurrency := runbookRunCommand.PersistentFlags().Int("max-concurrency", 20, "Maximum number of commands that can run simultaneously")

	for _, runbook := range runbooks.GetRunbooks() {
		// Create a copy of runbook scoped to the loop
		runbook := runbook

		// epcc runbook run <runbook_name>
		runbookRunRunbookCmd := &cobra.Command{
			Use:   runbook.Name,
			Long:  runbook.Description.Long,
			Short: runbook.Description.Short,
		}

		runbookRunCommand.AddCommand(runbookRunRunbookCmd)

		for _, runbookAction := range runbook.RunbookActions {
			// Create a copy of runbookAction scoped to the loop
			runbookAction := runbookAction

			runbookStringArguments := runbooks.CreateMapForRunbookArgumentPointers(runbookAction)

			// epcc runbook run <runbook> <action>
			runbookActionRunActionCommand := &cobra.Command{
				Use:   runbookAction.Name,
				Long:  runbookAction.Description.Long,
				Short: runbookAction.Description.Short,
				RunE: func(cmd *cobra.Command, args []string) error {
					numSteps := len(runbookAction.RawCommands)

					parentCtx := context.Background()

					ctx, cancelFunc := context.WithCancel(parentCtx)

					concurrentRunSemaphore := semaphore.NewWeighted(int64(*maxConcurrency))
					factory := pool.NewPooledObjectFactorySimple(
						func(ctx2 context.Context) (interface{}, error) {
							return generateRunbookCmd(), nil
						})

					objectPool := pool.NewObjectPool(ctx, factory, &pool.ObjectPoolConfig{
						MaxTotal: *maxConcurrency,
						MaxIdle:  *maxConcurrency,
					})
					for stepIdx, rawCmd := range runbookAction.RawCommands {

						// Create a copy of loop variables
						stepIdx := stepIdx
						rawCmd := rawCmd

						log.Infof("Executing> %s", rawCmd)
						templateName := fmt.Sprintf("Runbook: %s Action: %s Step: %d", runbook.Name, runbookAction.Name, stepIdx)
						rawCmdLines, err := runbooks.RenderTemplates(templateName, rawCmd, runbookStringArguments, runbookAction.Variables)

						if err != nil {
							cancelFunc()
							return err
						}
						resultChan := make(chan *commandResult, *maxConcurrency*2)
						funcs := make([]func(), 0, len(rawCmdLines))

						for commandIdx, rawCmdLine := range rawCmdLines {

							commandIdx := commandIdx
							rawCmdLine := strings.Trim(rawCmdLine, " \n")

							if rawCmdLine == "" {
								// Allow blank lines
								continue
							}

							if !strings.HasPrefix(rawCmdLine, "epcc ") {
								// Some commands like sleep don't have prefix
								// This hack allows them to run
								rawCmdLine = "epcc " + rawCmdLine
							}
							rawCmdArguments, err := shellwords.SplitPosix(strings.Trim(rawCmdLine, " \n"))

							if err != nil {
								cancelFunc()
								return err
							}

							funcs = append(funcs, func() {

								log.Tracef("(Step %d/%d Command %d/%d) Building Commmand", stepIdx+1, numSteps, commandIdx+1, len(funcs))

								stepCmdObject, err := objectPool.BorrowObject(ctx)
								defer objectPool.ReturnObject(ctx, stepCmdObject)

								if err == nil {
									commandAndResetFunc := stepCmdObject.(*CommandAndReset)
									commandAndResetFunc.reset()
									stepCmd := commandAndResetFunc.cmd

									stepCmd.SetArgs(rawCmdArguments[1:])
									log.Tracef("(Step %d/%d Command %d/%d) Starting Command", stepIdx+1, numSteps, commandIdx+1, len(funcs))

									stepCmd.ResetFlags()
									err = stepCmd.ExecuteContext(ctx)
									log.Tracef("(Step %d/%d Command %d/%d) Complete Command", stepIdx+1, numSteps, commandIdx+1, len(funcs))
								}

								commandResult := &commandResult{
									stepIdx:     stepIdx,
									commandIdx:  commandIdx,
									commandLine: rawCmdLine,
									error:       err,
								}

								resultChan <- commandResult

							})

						}

						if len(funcs) > 1 {
							log.Debugf("Running %d commands", len(funcs))
						}

						// Start processing all the functions
						go func() {
							for idx, fn := range funcs {
								idx := idx
								if shutdown.ShutdownFlag.Load() {
									log.Infof("Aborting runbook execution, after %d scheduled executions", idx)
									cancelFunc()
									break
								}

								fn := fn
								log.Tracef("Run %d is waiting on semaphore", idx)
								if err := concurrentRunSemaphore.Acquire(ctx, 1); err == nil {
									go func() {
										log.Tracef("Run %d is starting", idx)
										defer concurrentRunSemaphore.Release(1)
										fn()
									}()
								} else {
									log.Warnf("Run %d failed to get semaphore %v", idx, err)
								}
							}
						}()

						errorCount := 0
						for i := 0; i < len(funcs); i++ {
							select {
							case result := <-resultChan:
								if !shutdown.ShutdownFlag.Load() {
									if result.error != nil {
										log.Warnf("(Step %d/%d Command %d/%d) %v", result.stepIdx+1, numSteps, result.commandIdx+1, len(funcs), fmt.Errorf("error processing command [%s], %w", result.commandLine, result.error))
										errorCount++
									} else {
										log.Debugf("(Step %d/%d Command %d/%d) finished successfully ", result.stepIdx+1, numSteps, result.commandIdx+1, len(funcs))
									}
								} else {
									log.Tracef("Shutdown flag enabled, completion result %v", result)
									cancelFunc()
								}
							case <-time.After(time.Duration(*execTimeoutInSeconds) * time.Second):
								return fmt.Errorf("timeout of %d seconds reached, only %d of %d commands finished of step %d/%d", *execTimeoutInSeconds, i+1, len(funcs), stepIdx+1, numSteps)

							}
						}

						if len(funcs) > 1 {
							log.Debugf("Running %d commands complete", len(funcs))
						}

						if !runbookAction.IgnoreErrors && errorCount > 0 {
							return fmt.Errorf("error occurred while processing script aborting")
						}
					}
					defer cancelFunc()
					return nil
				},
			}
			processRunbookVariablesOnCommand(runbookActionRunActionCommand, runbookStringArguments, runbookAction.Variables, true)

			runbookRunRunbookCmd.AddCommand(runbookActionRunActionCommand)
		}
	}

	return runbookRunCommand
}

func processRunbookVariablesOnCommand(runbookActionRunActionCommand *cobra.Command, runbookStringArguments map[string]*string, variables map[string]runbooks.Variable, enableRequiredVars bool) {
	for key, variable := range variables {
		key := key
		variable := variable

		if variable.Required && enableRequiredVars {
			runbookActionRunActionCommand.Flags().StringVar(runbookStringArguments[key], key, "", variable.Description.Short)
			err := runbookActionRunActionCommand.MarkFlagRequired(key)

			if err != nil {
				log.Errorf("Could not set flag as required, this is a bug of some kind %s: %v", key, err)
			}
		} else {
			runbookActionRunActionCommand.Flags().StringVar(runbookStringArguments[key], key, variable.Default, variable.Description.Short)
		}

		runbookActionRunActionCommand.RegisterFlagCompletionFunc(key, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

			if strings.HasPrefix(variable.Type, "RESOURCE_ID:") {
				if resourceInfo, ok := resources.GetResourceByName(variable.Type[12:]); ok {
					return completion.Complete(completion.Request{
						Type:     completion.CompleteAlias,
						Resource: resourceInfo,
					})

				}
			}
			return []string{}, cobra.ShellCompDirectiveNoFileComp

		})
	}
}

// Creates a new instance of a cobra.Command
// We use a new instance for each step so that we can benefit from flags in runbooks

type CommandAndReset struct {
	cmd   *cobra.Command
	reset func()
}

func generateRunbookCmd() *CommandAndReset {
	root := &cobra.Command{
		Use:          "epcc",
		SilenceUsage: true,
	}

	resetCreateCmd := NewCreateCommand(root)
	resetUpdateCmd := NewUpdateCommand(root)
	resetDeleteCmd := NewDeleteCommand(root)
	resetGetCmd := NewGetCommand(root)
	resetDeleteAllCmd := NewDeleteAllCommand(root)
	getDevCommands(root)

	return &CommandAndReset{
		root,
		func() {
			// We need to reset the state of all commands since we are reusing the objects
			resetCreateCmd()
			resetUpdateCmd()
			resetDeleteCmd()
			resetGetCmd()
			resetDeleteAllCmd()
		},
	}
}

func initRunbookDevCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "dev",
		Hidden:       true,
		SilenceUsage: true,
	}

	getDevCommands(cmd)
	return cmd
}

func getDevCommands(parent *cobra.Command) {
	parent.AddCommand(&cobra.Command{
		Use:   "sleep",
		Short: "Sleep for a predefined duration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			timeToSleep, err := strconv.Atoi(args[0])

			if err != nil {
				return fmt.Errorf("could not sleep due to error: %v", err)
			}
			log.Infof("Sleeping for %d seconds", timeToSleep)
			time.Sleep(time.Duration(timeToSleep) * time.Second)

			return nil

		},
	})

}

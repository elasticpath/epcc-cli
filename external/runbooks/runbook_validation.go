package runbooks

import (
	"fmt"
	"github.com/buildkite/shellwords"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"strconv"
	"strings"
)

func ValidateRunbook(runbook *Runbook) error {

	if len(runbook.Name) == 0 {
		return fmt.Errorf("Runbook has no name")
	}

	if len(runbook.RunbookActions) == 0 {
		return fmt.Errorf("number of actions is zero")
	}

	for _, runbookAction := range runbook.RunbookActions {
		if (len(runbookAction.RawCommands)) == 0 {
			return fmt.Errorf("number of commands in action '%s' is zero", runbookAction.Name)
		}

		argumentsWithDefaults := CreateMapForRunbookArgumentPointers(runbookAction)

		cmds := runbookAction.RawCommands
		for stepIdx := 0; stepIdx < len(cmds); stepIdx++ {
			rawCmd := cmds[stepIdx]

			templateName := fmt.Sprintf("Runbook: %s Action: %s Step: %d", runbook.Name, runbookAction.Name, stepIdx+1)
			rawCmdLines, err := RenderTemplates(templateName, rawCmd, argumentsWithDefaults, runbookAction.Variables)

			if err != nil {
				return fmt.Errorf("error rendering template: %w", err)
			}

			joinedString := strings.Join(rawCmdLines, "\n")
			renderedCmd := []string{}
			err = yaml.Unmarshal([]byte(joinedString), &renderedCmd)

			if err == nil {
				log.Tracef("Line %d is a Yaml array %s, inserting into stack", stepIdx, joinedString)
				newCmds := make([]string, 0, len(cmds)+len(renderedCmd)-1)
				newCmds = append(newCmds, cmds[0:stepIdx]...)
				newCmds = append(newCmds, renderedCmd...)
				newCmds = append(newCmds, cmds[stepIdx+1:]...)
				cmds = newCmds
				stepIdx--
				continue
			}

			for commandIdx, rawCmdLine := range rawCmdLines {
				rawCmdLine := strings.Trim(rawCmdLine, " \n")

				if rawCmdLine == "" {
					// Allow blank lines
					continue
				}

				rawCmdArguments, err := shellwords.SplitPosix(strings.Trim(rawCmdLine, " \n"))

				if err != nil {
					return fmt.Errorf("error processing line at step %d line %d, %v", stepIdx+1, commandIdx+1, err)
				}

				if len(rawCmdArguments) < 1 {
					return fmt.Errorf("Each command should must have atleast one argument, but the line at step %d line %d does not:\n\t%s", stepIdx+1, commandIdx+1, rawCmdLine)
				}

				if rawCmdArguments[0] == "epcc" {
					if len(rawCmdArguments) < 2 {
						return fmt.Errorf("Each epcc command should be followed by a verb but the line at step %d line %d following line does not:\n\t%s", stepIdx+1, commandIdx+1, rawCmdLine)
					}

					switch rawCmdArguments[1] {
					case "get":
					case "delete":
					case "delete-all":
					case "create":
					case "update":
					default:
						return fmt.Errorf("Each command needs to have a valid verb of { get, create, update, delete, delete-all }, but we got %s in step %d line: %d", rawCmdArguments[1], stepIdx+1, commandIdx+1)
					}
				} else if rawCmdArguments[0] == "sleep" {
					_, err := strconv.Atoi(rawCmdArguments[1])
					if err != nil {
						return fmt.Errorf("Invalid argument to sleep %v, must be an integer in step %d line %d", rawCmdArguments[1], stepIdx+1, commandIdx+1)
					}
				} else {
					return fmt.Errorf("Each command needs be a recognized command, either { epcc, sleep }, but the line in step %d line %d is not:\n\t%s", stepIdx+1, commandIdx+1, rawCmdArguments[0])
				}
			}

		}

	}

	return nil

}

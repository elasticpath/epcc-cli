package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/profiles"
	"github.com/spf13/cobra"
	"strconv"
)

var LogsClear = &cobra.Command{
	Use:   "clear",
	Short: "Clears all HTTP request and response logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		return profiles.ClearAllRequestLogs()
	},
}

var LogsList = &cobra.Command{
	Use:   "list",
	Short: "List all HTTP logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		files, err := profiles.GetAllRequestLogTitles()
		if err != nil {
			return err
		}

		for idx, name := range files {
			fmt.Printf("%d %s\n", idx, name)
		}
		return nil
	},
}

var LogsShow = &cobra.Command{
	Use:   "show <NUMBER>",
	Short: "Show HTTP logs for specific number",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		i, err := strconv.Atoi(args[0])

		if err != nil {
			return fmt.Errorf("Could not get the %s entry => %w", args[0], err)
		}

		content, err := profiles.GetNthRequestLog(i)

		if err != nil {
			return fmt.Errorf("Couldn't print logs: %v", err)
		}

		fmt.Println(content)

		return nil
	},
}

var Logs = &cobra.Command{Use: "logs", Short: "Retrieve information about previous requests"}

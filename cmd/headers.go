package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/headergroups"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewHeadersCommand(parentCmd *cobra.Command) func() {

	var headersCmd = &cobra.Command{
		Use:          "headers",
		Short:        "Set headers that should be used on all subsequent calls",
		SilenceUsage: true,
	}

	parentCmd.AddCommand(headersCmd)

	var setGroup = "default"
	var delGroup = "default"
	resetFunc := func() {
		setGroup = "default"
		delGroup = "default"
	}

	var setHeaderCmd = &cobra.Command{
		Use:   "set [HEADER_KEY] [HEADER_VALUE] ...",
		Short: "Set a header to be used on all subsequent requests",
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

			if len(args)%2 == 0 {
				return completion.Complete(completion.Request{
					Type:       completion.CompleteHeaderKey,
					ToComplete: toComplete,
				})
			} else {
				return completion.Complete(completion.Request{
					Type: completion.CompleteHeaderValue,
					// Get the second last value
					Header:     args[len(args)-1],
					ToComplete: toComplete,
				})
			}

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args)%2 != 0 {
				return fmt.Errorf("Invalid number of arguments received, should be an even number of key and values: %d", len(args))
			}

			for i := 0; i < len(args); i += 2 {
				headergroups.AddHeaderToGroup(setGroup, args[i], args[i+1])
			}

			return nil
		},
	}

	var delHeaderCmd = &cobra.Command{
		Use:   "delete HEADER_KEY...",
		Short: "Deletes a header from a group",
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			results, compDir := completion.Complete(completion.Request{
				Type:       completion.CompleteHeaderKey,
				ToComplete: toComplete,
			})

			for h := range headergroups.GetAllHeaders() {
				results = append(results, h)
			}

			return results, compDir
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			for _, headerKey := range args {
				headergroups.RemoveHeaderFromGroup(delGroup, headerKey)
			}
			return nil
		},
	}

	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Displays the current headers that are set, and the header groups that are in use.",
		RunE: func(cmd *cobra.Command, args []string) error {
			hgs := headergroups.GetAllHeaderGroups()

			for _, hg := range hgs {
				log.Infof("We are using a header group: %s", hg)
			}

			for k, v := range headergroups.GetAllHeaders() {
				log.Infof("Using header %s: %s", k, v)
			}

			log.Infof("Header information stored in %v", headergroups.GetHeaderGroupPath())
			return nil
		},
	}

	var clearGroupCmd = &cobra.Command{
		Use:   "clear [GROUP_NAME]",
		Short: "Clears a header group",
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return headergroups.GetAllHeaderGroups(), cobra.ShellCompDirectiveNoFileComp
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("expected exactly one argument, the group name, got %d", len(args))
			}

			headergroups.RemoveHeaderGroup(args[0])
			return nil
		},
	}

	setHeaderCmd.PersistentFlags().StringVar(&setGroup, "group", "default", "Stores the header with a group (so that you can easily clear them)")
	delHeaderCmd.PersistentFlags().StringVar(&delGroup, "group", "default", "Removes the header from within a group")
	headersCmd.AddCommand(setHeaderCmd)
	headersCmd.AddCommand(delHeaderCmd)
	headersCmd.AddCommand(statusCmd)
	headersCmd.AddCommand(clearGroupCmd)

	return resetFunc

}

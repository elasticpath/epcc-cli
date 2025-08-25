package cmd

import (
	"fmt"
	"strings"

	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewResourceInfoCommand(parentCmd *cobra.Command) func() {
	resetFunc := func() {
		// No state to reset for now
	}

	var resourceInfoCmd = &cobra.Command{
		Use:          "resource-info",
		Short:        "Shows information about resources",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("please specify a resource, epcc resource-info [RESOURCE], see epcc resource-info --help")
			} else {
				return fmt.Errorf("invalid resource [%s] specified, see all with epcc resource-info --help", args[0])
			}
		},
	}

	e := config.GetEnv()
	hiddenResources := map[string]struct{}{}

	for _, v := range e.EPCC_CLI_DISABLE_RESOURCES {
		hiddenResources[v] = struct{}{}
	}

	for _, resource := range resources.GetPluralResources() {

		if _, ok := hiddenResources[resource.SingularName]; ok {
			log.Tracef("Hiding resource %s", resource.SingularName)
			continue
		}

		if _, ok := hiddenResources[resource.PluralName]; ok {
			log.Tracef("Hiding resource %s", resource.PluralName)
			continue
		}

		resource := resource

		// Create aliases slice - include singular name if different from plural
		var aliases []string
		if resource.SingularName != resource.PluralName {
			aliases = []string{resource.SingularName}
		}

		// Create the main command using the plural name with singular as alias
		pluralCmd := &cobra.Command{
			Use:     resource.PluralName,
			Aliases: aliases,
			Short:   fmt.Sprintf("Show information about %s resource", resource.PluralName),
			RunE: func(cmd *cobra.Command, args []string) error {
				// For now, just print a placeholder message
				fmt.Printf(GenerateResourceInfo(&resource))
				return nil
			},
		}

		resourceInfoCmd.AddCommand(pluralCmd)
	}

	parentCmd.AddCommand(resourceInfoCmd)

	return resetFunc
}

// getIndefiniteArticle returns "a" or "an" based on the first letter/sound of the word
func getIndefiniteArticle(word string) string {
	if len(word) == 0 {
		return "a"
	}

	// Convert to lowercase for checking
	lower := strings.ToLower(word)

	// Words that start with vowel sounds but use "a"
	vowelExceptions := map[string]bool{
		"university": true,
		"user":       true,
		"uniform":    true,
		"unit":       true,
		"unique":     true,
		"usage":      true,
		"utility":    true,
	}

	// Words that start with consonants but use "an"
	consonantExceptions := map[string]bool{
		"hour":   true,
		"honest": true,
		"honor":  true,
		"heir":   true,
	}

	// Check exceptions first
	if vowelExceptions[lower] {
		return "a"
	}
	if consonantExceptions[lower] {
		return "an"
	}

	// Default vowel rule
	firstChar := lower[0]
	if firstChar == 'a' || firstChar == 'e' || firstChar == 'i' || firstChar == 'o' || firstChar == 'u' {
		return "an"
	}

	return "a"
}

func GenerateResourceInfo(r *resources.Resource) string {
	sb := strings.Builder{}

	tabs := "  "
	article := getIndefiniteArticle(r.SingularName)

	sb.WriteString("Operations: \n")

	if r.GetCollectionInfo != nil {
		usageString := GetGetUsageString(r.PluralName, r.GetCollectionInfo.Url, collectionResourceRequest, *r)
		sb.WriteString(fmt.Sprintf("%sepcc get %s - get a page of %s\n", tabs, usageString, r.PluralName))

		types, _ := resources.GetSingularTypesOfVariablesNeeded(r.GetCollectionInfo.Url)

		if len(types) > 0 {

			sb.WriteString(tabs + "Parent Resource ID Parameters (Mandatory):\n")

			for _, t := range types {
				paramName := ConvertSingularTypeToCmdArg(t)
				sb.WriteString(fmt.Sprintf("  %-20s - An ID or alias for a %s\n", paramName, strings.Title(t)))
			}
		}
	}

	sb.WriteString("\n\n")

	if r.CreateEntityInfo != nil {
		usageString := GetCreateUsageString(*r)
		sb.WriteString(fmt.Sprintf("%sepcc create %s - create %s %s\n", tabs, usageString, article, r.SingularName))
		types, _ := resources.GetSingularTypesOfVariablesNeeded(r.CreateEntityInfo.Url)

		if len(types) > 0 {

			sb.WriteString(tabs + "Parent Resource ID Parameters (Mandatory):\n")

			for _, t := range types {
				paramName := ConvertSingularTypeToCmdArg(t)
				sb.WriteString(fmt.Sprintf("  %-20s - An ID or alias for a %s\n", paramName, strings.Title(t)))
			}
		}
	}

	sb.WriteString("\n\n")

	if r.GetEntityInfo != nil {
		usageString := GetGetUsageString(r.SingularName, r.GetEntityInfo.Url, singularResourceRequest, *r)
		sb.WriteString(fmt.Sprintf("%sepcc get %s - get %s %s\n", tabs, usageString, article, r.SingularName))

		types, _ := resources.GetSingularTypesOfVariablesNeeded(r.GetEntityInfo.Url)

		if len(types) > 0 {

			sb.WriteString(tabs + "Parent Resource ID Parameters (Mandatory):\n")

			for _, t := range types {
				paramName := ConvertSingularTypeToCmdArg(t)
				sb.WriteString(fmt.Sprintf("  %-20s - An ID or alias for a %s\n", paramName, strings.Title(t)))
			}
		}
	}

	sb.WriteString("\n\n")

	if r.UpdateEntityInfo != nil {
		usageString := GetUpdateUsage(*r)
		sb.WriteString(fmt.Sprintf("%sepcc update %s - update %s %s\n", tabs, usageString, article, r.SingularName))

		types, _ := resources.GetSingularTypesOfVariablesNeeded(r.UpdateEntityInfo.Url)

		if len(types) > 0 {

			sb.WriteString(tabs + "Parent Resource ID Parameters (Mandatory):\n")

			for _, t := range types {
				paramName := ConvertSingularTypeToCmdArg(t)
				sb.WriteString(fmt.Sprintf("  %-20s - An ID or alias for a %s\n", paramName, strings.Title(t)))
			}
		}
	}

	sb.WriteString("\n\n")

	if r.DeleteEntityInfo != nil {
		usageString := GetDeleteUsage(*r)
		sb.WriteString(fmt.Sprintf("%sepcc delete %s - delete %s %s\n", tabs, usageString, article, r.SingularName))

		types, _ := resources.GetSingularTypesOfVariablesNeeded(r.DeleteEntityInfo.Url)

		if len(types) > 0 {

			sb.WriteString(tabs + "Parent Resource ID Parameters (Mandatory):\n")

			for _, t := range types {
				paramName := ConvertSingularTypeToCmdArg(t)
				sb.WriteString(fmt.Sprintf("  %-20s - An ID or alias for a %s\n", paramName, strings.Title(t)))
			}
		}
	}

	return sb.String()
}

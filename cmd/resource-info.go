package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/openapi"
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

		var openApiFlag = false

		// Create the main command using the plural name with singular as alias
		pluralCmd := &cobra.Command{
			Use:     resource.PluralName,
			Aliases: aliases,
			Short:   fmt.Sprintf("Show information about %s resource", resource.PluralName),
			RunE: func(cmd *cobra.Command, args []string) error {
				// For now, just print a placeholder message
				fmt.Printf(GenerateResourceInfo(&resource))

				if openApiFlag {
					fmt.Printf(GenerateOpenApiInfo(&resource))
				}

				return nil
			},
		}

		pluralCmd.Flags().BoolVarP(&openApiFlag, "openapi", "", false, "display openapi information")

		resourceInfoCmd.AddCommand(pluralCmd)
	}

	parentCmd.AddCommand(resourceInfoCmd)

	return resetFunc
}

// GetOtherReferences finds commands from other resources that reference the current resource
func GetOtherReferences(currentResource *resources.Resource) string {
	sb := strings.Builder{}

	sb.WriteString("\n\n*** Referenced By ***\n\n")

	currentResourceName := currentResource.SingularName
	foundUrlReferences := []string{}
	foundBodyReferences := []string{}

	foundAliasTypes := []string{}

	for _, resource := range resources.GetPluralResources() {

		// Skip entity-relationship resources
		if strings.Contains(resource.SingularName, "entity-relationship") {
			continue
		}

		for _, aType := range currentResource.AlternateJsonApiTypesForAliases {
			if aType == resource.JsonApiType {
				foundAliasTypes = append(foundAliasTypes, aType)
			}
		}

		// Check each operation URL for references to the current resource
		operations := []struct {
			info *resources.CrudEntityInfo
			verb string
			name string
		}{
			{resource.GetCollectionInfo, "get", resource.PluralName},
			{resource.CreateEntityInfo, "create", resource.SingularName},
			{resource.GetEntityInfo, "get", resource.SingularName},
			{resource.UpdateEntityInfo, "update", resource.SingularName},
			{resource.DeleteEntityInfo, "delete", resource.SingularName},
		}

		for _, op := range operations {
			if resource.SingularName == currentResourceName {
				continue // Skip the current resource
			}

			if op.info == nil {
				continue
			}

			// Get the required parameters for this operation
			types, err := resources.GetSingularTypesOfVariablesNeeded(op.info.Url)
			if err != nil {
				continue
			}

			// Check if the current resource is referenced in this operation
			foundCurrentResource := false
			for _, t := range types {
				if t == currentResourceName {
					foundCurrentResource = true
					break
				}
			}

			if foundCurrentResource {
				// Build the command with just the mandatory parameters
				cmdParts := []string{"epcc", op.verb, op.name}
				for _, t := range types {
					cmdParts = append(cmdParts, ConvertSingularTypeToCmdArg(t))
				}

				command := strings.Join(cmdParts, " ")
				foundUrlReferences = append(foundUrlReferences, command)
			}
		}

		// Check if this resource's attributes reference the current resource in body parameters
		for k, attr := range resource.Attributes {
			if strings.HasPrefix(attr.Type, "RESOURCE_ID:") {
				referencedResourceType := strings.TrimPrefix(attr.Type, "RESOURCE_ID:")
				if currentResource.SingularName == referencedResourceType || currentResource.PluralName == referencedResourceType {
					// This resource has a body parameter that references the current resource
					var operations []string

					if resource.CreateEntityInfo != nil {
						operations = append(operations, "create")
					}
					if resource.UpdateEntityInfo != nil {
						operations = append(operations, "update")
					}

					if len(operations) > 0 {
						var opString string
						if len(operations) == 2 {
							opString = "{create,update}"
						} else {
							opString = operations[0]
						}

						bodyRef := fmt.Sprintf("%s in epcc %s %s", k, opString, resource.SingularName)
						foundBodyReferences = append(foundBodyReferences, bodyRef)
					}
				}
			}
		}

	}

	// Remove duplicates and sort URL references by resource name
	uniqueUrlRefs := make(map[string]bool)
	for _, ref := range foundUrlReferences {
		uniqueUrlRefs[ref] = true
	}

	sortedUrlRefs := make([]string, 0, len(uniqueUrlRefs))
	for ref := range uniqueUrlRefs {
		sortedUrlRefs = append(sortedUrlRefs, ref)
	}

	// Remove duplicates from body references
	uniqueBodyRefs := make(map[string]bool)
	for _, ref := range foundBodyReferences {
		uniqueBodyRefs[ref] = true
	}

	sortedBodyRefs := make([]string, 0, len(uniqueBodyRefs))
	for ref := range uniqueBodyRefs {
		sortedBodyRefs = append(sortedBodyRefs, ref)
	}

	// Define verb order for sorting
	verbOrder := map[string]int{
		"get":    0, // get-collection will be handled by resource name (plural vs singular)
		"create": 1,
		"update": 2,
		"delete": 3,
	}

	// Sort URL references by resource name first, then by verb order
	sort.Slice(sortedUrlRefs, func(i, j int) bool {
		// Extract resource name and verb from command
		partsI := strings.Fields(sortedUrlRefs[i])
		partsJ := strings.Fields(sortedUrlRefs[j])

		if len(partsI) >= 3 && len(partsJ) >= 3 {
			resourceI := partsI[2] // resource name is 3rd element
			resourceJ := partsJ[2]
			verbI := partsI[1] // verb is 2nd element
			verbJ := partsJ[1]

			// If resource names are different, sort by resource name
			if resourceI != resourceJ {
				return resourceI < resourceJ
			}

			// Same resource, sort by verb order
			orderI, okI := verbOrder[verbI]
			orderJ, okJ := verbOrder[verbJ]

			if okI && okJ {
				return orderI < orderJ
			}

			// Fallback to alphabetical for unknown verbs
			return verbI < verbJ
		}

		// Fallback to string comparison
		return sortedUrlRefs[i] < sortedUrlRefs[j]
	})

	// Sort body references alphabetically
	sort.Strings(sortedBodyRefs)

	sortedAliasedResources := []string{}

	// For every resource we reference, add them as another type
	for _, alias := range foundAliasTypes {
		for name, r := range resources.GetPluralResources() {
			if r.JsonApiType == alias {
				sortedAliasedResources = append(sortedAliasedResources, name)
			}
		}
	}

	// For every resource that references us add them as another type
	for name, r := range resources.GetPluralResources() {
		for _, a := range r.AlternateJsonApiTypesForAliases {
			if a == currentResource.JsonApiType {
				sortedAliasedResources = append(sortedAliasedResources, name)
			}
		}
	}

	sort.Strings(sortedAliasedResources)

	// Check if we have any references at all
	if len(sortedUrlRefs) == 0 && len(sortedBodyRefs) == 0 && len(sortedAliasedResources) == 0 {
		sb.WriteString("No other commands reference this resource.\n")
	} else {
		// URL References subsection
		if len(sortedUrlRefs) > 0 {
			sb.WriteString("**** URL Parameter ****\n")
			var lastResource string
			for _, ref := range sortedUrlRefs {
				// Extract resource name to detect when we switch to a new resource
				parts := strings.Fields(ref)
				if len(parts) >= 3 {
					currentResource := parts[2]

					// Add a newline between different resources (but not before the first one)
					if lastResource != "" && lastResource != currentResource {
						sb.WriteString("\n")
					}

					lastResource = currentResource
				}

				sb.WriteString(ref + "\n")
			}
		}

		// Body Parameters subsection
		if len(sortedBodyRefs) > 0 {
			if len(sortedUrlRefs) > 0 {
				sb.WriteString("\n") // Add spacing between sections
			}
			sb.WriteString("**** Attributes ****\n\n")
			for _, ref := range sortedBodyRefs {
				sb.WriteString("  - " + ref + "\n")
			}

			sb.WriteString("\n")
		}

		// Aliases
		if len(sortedAliasedResources) > 0 {
			sb.WriteString("\n**** In URL ****\n\n")
			sb.WriteString("These resources share ids and so probably have related lifecycles\n")

			for _, alias := range sortedAliasedResources {
				sb.WriteString(" - " + alias + "\n")
			}
		}
	}

	return sb.String()
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

			sb.WriteString("\n" + tabs + tabs + "Resource ID Parameters (Mandatory):\n")

			for _, t := range types {
				paramName := ConvertSingularTypeToCmdArg(t)
				article := getIndefiniteArticle(strings.Title(t))
				sb.WriteString(fmt.Sprintf("    %-20s - An ID or alias for %s %s\n", paramName, article, strings.Title(t)))
			}
		}
	}

	sb.WriteString("\n")

	if r.CreateEntityInfo != nil {
		usageString := GetCreateUsageString(*r)
		sb.WriteString(fmt.Sprintf("%sepcc create %s - create %s %s\n", tabs, usageString, article, r.SingularName))
		types, _ := resources.GetSingularTypesOfVariablesNeeded(r.CreateEntityInfo.Url)

		if len(types) > 0 {

			sb.WriteString("\n" + tabs + tabs + "Resource ID Parameters (Mandatory):\n")

			for _, t := range types {
				paramName := ConvertSingularTypeToCmdArg(t)
				article := getIndefiniteArticle(strings.Title(t))
				sb.WriteString(fmt.Sprintf("    %-20s - An ID or alias for %s %s\n", paramName, article, strings.Title(t)))
			}
		}
	}

	sb.WriteString("\n")

	if r.GetEntityInfo != nil {
		usageString := GetGetUsageString(r.SingularName, r.GetEntityInfo.Url, singularResourceRequest, *r)
		sb.WriteString(fmt.Sprintf("%sepcc get %s - get %s %s\n", tabs, usageString, article, r.SingularName))

		types, _ := resources.GetSingularTypesOfVariablesNeeded(r.GetEntityInfo.Url)

		if len(types) > 0 {

			sb.WriteString("\n" + tabs + tabs + "Resource ID Parameters (Mandatory):\n")

			for _, t := range types {
				paramName := ConvertSingularTypeToCmdArg(t)
				article := getIndefiniteArticle(strings.Title(t))
				sb.WriteString(fmt.Sprintf("    %-20s - An ID or alias for %s %s\n", paramName, article, strings.Title(t)))
			}
		}
	}

	sb.WriteString("\n")

	if r.UpdateEntityInfo != nil {
		usageString := GetUpdateUsage(*r)
		sb.WriteString(fmt.Sprintf("%sepcc update %s - update %s %s\n", tabs, usageString, article, r.SingularName))

		types, _ := resources.GetSingularTypesOfVariablesNeeded(r.UpdateEntityInfo.Url)

		if len(types) > 0 {

			sb.WriteString("\n" + tabs + tabs + "Resource ID Parameters (Mandatory):\n")

			for _, t := range types {
				paramName := ConvertSingularTypeToCmdArg(t)
				article := getIndefiniteArticle(strings.Title(t))
				sb.WriteString(fmt.Sprintf("    %-20s - An ID or alias for %s %s\n", paramName, article, strings.Title(t)))
			}
		}
	}

	sb.WriteString("\n")

	if r.DeleteEntityInfo != nil {
		usageString := GetDeleteUsage(*r)
		sb.WriteString(fmt.Sprintf("%sepcc delete %s - delete %s %s\n", tabs, usageString, article, r.SingularName))

		types, _ := resources.GetSingularTypesOfVariablesNeeded(r.DeleteEntityInfo.Url)

		if len(types) > 0 {

			sb.WriteString(tabs + "Resource ID Parameters (Mandatory):\n")

			for _, t := range types {
				paramName := ConvertSingularTypeToCmdArg(t)
				article := getIndefiniteArticle(strings.Title(t))
				sb.WriteString(fmt.Sprintf("  %-20s - An ID or alias for %s %s\n", paramName, article, strings.Title(t)))
			}
		}
	}

	// Add body parameters section at the bottom (shared across all operations)
	if len(r.Attributes) > 0 {
		sb.WriteString("\n")
		bodyParamsUsage := GetParameterUsageForTypes(*r, []string{}, true)
		sb.WriteString(bodyParamsUsage)
	}

	// Add other references section at the very end
	otherReferences := GetOtherReferences(r)
	sb.WriteString(otherReferences)

	sb.WriteString("\n")

	return sb.String()
}

func GenerateOpenApiInfo(resource *resources.Resource) string {

	operationIds := map[string]bool{}
	if resource.GetCollectionInfo != nil {
		operationIds[resource.GetCollectionInfo.OpenApiOperationId] = true
	}

	if resource.GetEntityInfo != nil {
		operationIds[resource.GetEntityInfo.OpenApiOperationId] = true
	}

	if resource.UpdateEntityInfo != nil {
		operationIds[resource.UpdateEntityInfo.OpenApiOperationId] = true
	}

	if resource.DeleteEntityInfo != nil {
		operationIds[resource.DeleteEntityInfo.OpenApiOperationId] = true

	}

	if resource.CreateEntityInfo != nil {
		operationIds[resource.CreateEntityInfo.OpenApiOperationId] = true
	}

	sb := strings.Builder{}
	sb.WriteString("** OpenAPI Excerpts **")
	found := 0

	opIds := []string{}

	for k := range operationIds {
		opIds = append(opIds, k)
	}

	sort.Strings(opIds)

	opStrings := strings.Builder{}

	tags := map[string]bool{}

	for _, opId := range opIds {
		if opId != "" {
			op, err := openapi.FindOperationByID(opId)

			if err != nil || op == nil {
				log.Warnf("Could not find operation id: %d", op)
			}

			found++

			yaml, err := op.Operation.RenderInline()

			if err != nil {
				log.Warnf("Couldn't render operation")
			}

			opStrings.WriteString(fmt.Sprintf("\n*** Operation: %s ***\n", opId))
			opStrings.Write(yaml)
			opStrings.WriteString("\n")

			for _, v := range op.Operation.Tags {
				tags[v] = true
			}
		}
	}

	tagList := []string{}

	for k := range tags {
		tagList = append(tagList, k)
	}

	sort.Strings(tagList)

	for _, tag := range tagList {

		tagInfo, description, err := openapi.FindTagByName(tag)
		if err != nil {
			sb.WriteString(fmt.Sprintf("\n*** Tag: %s ***\n", tag))
			sb.WriteString("Could not find tag information")
		} else {
			if description != "" {
				sb.WriteString("\n*** Document Description***\n")
				sb.WriteString(description)
			}

			sb.WriteString(fmt.Sprintf("\n*** Tag: %s ***\n", tag))
			sb.WriteString(tagInfo.Description)
			sb.WriteString("\n")
		}
	}

	if found == 0 {
		log.Warnf("Could not find any OpenAPI operations for resource: %s", resource.PluralName)
	}

	return sb.String() + opStrings.String()
}

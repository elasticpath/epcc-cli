package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/completion"
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
				fmt.Println(GenerateResourceInfo(&resource))

				if openApiFlag {
					fmt.Println(GenerateOpenApiInfo(&resource))
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

	sb.WriteString("\nReferenced In:\n")

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
					// Build command lines for create and update separately since they have different URLs

					if resource.CreateEntityInfo != nil {
						// Build create command with URL parameters
						cmdParts := []string{"epcc", "create", resource.SingularName}

						// Add URL parameters
						types, err := resources.GetSingularTypesOfVariablesNeeded(resource.CreateEntityInfo.Url)
						if err == nil {
							for _, t := range types {
								cmdParts = append(cmdParts, ConvertSingularTypeToCmdArg(t))
							}
						}

						// Add body parameter and its value
						// Handle array parameters
						if strings.Contains(k, "[n]") {
							// Show two examples for array parameters
							baseKey := strings.ReplaceAll(k, "[n]", "")
							cmdParts = append(cmdParts, baseKey+"[0]", ConvertSingularTypeToCmdArg(currentResourceName))
							cmdParts = append(cmdParts, baseKey+"[1]", ConvertSingularTypeToCmdArg(currentResourceName))
						} else {
							cmdParts = append(cmdParts, k, ConvertSingularTypeToCmdArg(currentResourceName))
						}

						bodyRef := strings.Join(cmdParts, " ")
						foundBodyReferences = append(foundBodyReferences, bodyRef)
					}

					if resource.UpdateEntityInfo != nil {
						// Build update command with URL parameters
						cmdParts := []string{"epcc", "update", resource.SingularName}

						// Add URL parameters
						types, err := resources.GetSingularTypesOfVariablesNeeded(resource.UpdateEntityInfo.Url)
						if err == nil {
							for _, t := range types {
								cmdParts = append(cmdParts, ConvertSingularTypeToCmdArg(t))
							}
						}

						// Add body parameter and its value
						// Handle array parameters
						if strings.Contains(k, "[n]") {
							// Show two examples for array parameters
							baseKey := strings.ReplaceAll(k, "[n]", "")
							cmdParts = append(cmdParts, baseKey+"[0]", ConvertSingularTypeToCmdArg(currentResourceName))
							cmdParts = append(cmdParts, baseKey+"[1]", ConvertSingularTypeToCmdArg(currentResourceName))
						} else {
							cmdParts = append(cmdParts, k, ConvertSingularTypeToCmdArg(currentResourceName))
						}

						bodyRef := strings.Join(cmdParts, " ")
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

	// Sort body references by resource name first, then by verb order
	sort.Slice(sortedBodyRefs, func(i, j int) bool {
		partsI := strings.Fields(sortedBodyRefs[i])
		partsJ := strings.Fields(sortedBodyRefs[j])

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
		return sortedBodyRefs[i] < sortedBodyRefs[j]
	})

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

			// Group commands by resource name and parameters
			type commandKey struct {
				resource string
				params   string
			}
			commandGroups := make(map[commandKey][]string)

			for _, ref := range sortedUrlRefs {
				parts := strings.Fields(ref)
				if len(parts) >= 3 {
					verb := parts[1]
					resource := parts[2]
					params := strings.Join(parts[3:], " ")

					key := commandKey{resource: resource, params: params}
					commandGroups[key] = append(commandGroups[key], verb)
				}
			}

			// Sort the grouped commands
			type groupedCommand struct {
				resource string
				params   string
				verbs    []string
			}

			var grouped []groupedCommand
			for key, verbs := range commandGroups {
				// Sort verbs by the defined order
				sort.Slice(verbs, func(i, j int) bool {
					orderI := verbOrder[verbs[i]]
					orderJ := verbOrder[verbs[j]]
					return orderI < orderJ
				})
				grouped = append(grouped, groupedCommand{
					resource: key.resource,
					params:   key.params,
					verbs:    verbs,
				})
			}

			// Sort grouped commands by resource name
			sort.Slice(grouped, func(i, j int) bool {
				return grouped[i].resource < grouped[j].resource
			})

			// Output the grouped commands
			var lastResource string
			for _, cmd := range grouped {
				// Add a newline between different resources (but not before the first one)
				if lastResource != "" && lastResource != cmd.resource {
					sb.WriteString("\n")
				}
				lastResource = cmd.resource

				// Format the verb list
				var verbStr string
				if len(cmd.verbs) == 1 {
					verbStr = cmd.verbs[0]
				} else {
					verbStr = "{" + strings.Join(cmd.verbs, ",") + "}"
				}

				// Build the command line
				if cmd.params != "" {
					sb.WriteString(fmt.Sprintf("  epcc %s %s %s ...\n", verbStr, cmd.resource, cmd.params))
				} else {
					sb.WriteString(fmt.Sprintf("  epcc %s %s ...\n", verbStr, cmd.resource))
				}
			}
		}

		// Body Parameters subsection
		if len(sortedBodyRefs) > 0 {
			if len(sortedUrlRefs) > 0 {
				sb.WriteString("\n")
			}

			var lastResource string
			for _, ref := range sortedBodyRefs {
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

				sb.WriteString("  " + ref + " ...\n")
			}
		}

		// Aliases
		if len(sortedAliasedResources) > 0 {
			if len(sortedUrlRefs) > 0 || len(sortedBodyRefs) > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString("Related resources (share IDs):\n")

			for _, alias := range sortedAliasedResources {
				sb.WriteString(" - " + alias + "\n")
			}
		}

		// Created By
		if len(currentResource.CreatedBy) > 0 {
			if len(sortedUrlRefs) > 0 || len(sortedBodyRefs) > 0 || len(sortedAliasedResources) > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString("Created By:\n")

			for _, created := range currentResource.CreatedBy {
				sb.WriteString(fmt.Sprintf("  epcc %s %s \n", created.Verb, created.Resource))
			}
		}
	}

	return sb.String()
}

func GenerateResourceInfo(r *resources.Resource) string {
	sb := strings.Builder{}

	tabs := "  "
	article := getIndefiniteArticle(r.SingularName)

	sb.WriteString("Operations:\n")

	// Collect all unique URL parameters across all operations
	allUrlPathParams := make(map[string]bool)

	queryParams := make(map[string]*resources.QueryParameter)

	if r.GetCollectionInfo != nil {
		usageString := GetGetUsageString(r.PluralName, r.GetCollectionInfo.Url, completion.GetAll, *r)
		sb.WriteString(fmt.Sprintf("%sepcc get %s - get a page of %s\n\n", tabs, usageString, r.PluralName))

		types, _ := resources.GetSingularTypesOfVariablesNeeded(r.GetCollectionInfo.Url)
		for _, t := range types {
			allUrlPathParams[t] = true
		}

		for _, v := range r.GetCollectionInfo.QueryParameters {
			queryParams[v.Name] = &v
		}
	}

	if r.CreateEntityInfo != nil {
		usageString := GetCreateUsageString(*r)

		sb.WriteString(fmt.Sprintf("%sepcc create %s - create %s %s\n", tabs, usageString, article, r.SingularName))

		if r.CreateEntityInfo.Creates != "" {
			sb.WriteString(fmt.Sprintf("\n   Note: The created resource is %s %s", article, r.CreateEntityInfo.Creates))
		}

		sb.WriteString("\n")
		types, _ := resources.GetSingularTypesOfVariablesNeeded(r.CreateEntityInfo.Url)
		for _, t := range types {
			allUrlPathParams[t] = true
		}

		for _, v := range r.CreateEntityInfo.QueryParameters {
			queryParams[v.Name] = &v
		}
	}

	if r.GetEntityInfo != nil {
		usageString := GetGetUsageString(r.SingularName, r.GetEntityInfo.Url, completion.Get, *r)
		sb.WriteString(fmt.Sprintf("%sepcc get %s - get %s %s\n\n", tabs, usageString, article, r.SingularName))

		types, _ := resources.GetSingularTypesOfVariablesNeeded(r.GetEntityInfo.Url)
		for _, t := range types {
			allUrlPathParams[t] = true
		}

		for _, v := range r.GetEntityInfo.QueryParameters {
			queryParams[v.Name] = &v
		}
	}

	if r.UpdateEntityInfo != nil {
		usageString := GetUpdateUsage(*r)
		sb.WriteString(fmt.Sprintf("%sepcc update %s - update %s %s\n\n", tabs, usageString, article, r.SingularName))

		types, _ := resources.GetSingularTypesOfVariablesNeeded(r.UpdateEntityInfo.Url)
		for _, t := range types {
			allUrlPathParams[t] = true
		}

		for _, v := range r.UpdateEntityInfo.QueryParameters {
			queryParams[v.Name] = &v
		}
	}

	if r.DeleteEntityInfo != nil {
		usageString := GetDeleteUsage(*r)
		sb.WriteString(fmt.Sprintf("%sepcc delete %s - delete %s %s\n\n", tabs, usageString, article, r.SingularName))

		types, _ := resources.GetSingularTypesOfVariablesNeeded(r.DeleteEntityInfo.Url)
		for _, t := range types {
			allUrlPathParams[t] = true
		}

		for _, v := range r.DeleteEntityInfo.QueryParameters {
			queryParams[v.Name] = &v
		}
	}

	// Output consolidated parameters section (URL params + body params + query params)
	if len(allUrlPathParams) > 0 || len(r.Attributes) > 0 || len(queryParams) > 0 {

		// Collect all parameters with their descriptions
		type paramInfo struct {
			name        string
			description string
			when        string
		}

		allParams := map[string][]paramInfo{
			"": {},
		}

		// Add URL parameters
		for param := range allUrlPathParams {
			paramName := ConvertSingularTypeToCmdArg(param)
			article := getIndefiniteArticle(strings.Title(param))
			description := fmt.Sprintf("An ID or alias for %s %s", article, strings.Title(param))
			allParams[""] = append(allParams[""], paramInfo{name: paramName, description: description})
		}

		for param, info := range queryParams {

			description := GetParameterDescription(info.Name, info.Type, info.Usage, "", true)

			value := strings.Trim(NonAlphaCharacter.ReplaceAllString(strings.ToUpper(param), "_"), "_ ")
			value = strings.ReplaceAll(value, "A_Z", "")
			value = strings.ReplaceAll(value, "__", "_")

			allParams[""] = append(allParams[""], paramInfo{name: value, description: description})
		}
		// Add body parameters (converted to uppercase)
		for k, v := range r.Attributes {
			paramName := strings.ToUpper(k)

			description := GetParameterDescription(k, v.Type, v.Usage, v.AliasAttribute, false)

			if _, ok := allParams[v.When]; !ok {
				allParams[v.When] = []paramInfo{}
			}

			allParams[v.When] = append(allParams[v.When], paramInfo{name: paramName, description: description, when: v.When})
		}

		allWhens := []string{}
		for k := range allParams {
			allWhens = append(allWhens, k)
		}

		sort.Strings(allWhens)

		for _, w := range allWhens {
			if w == "" {
				sb.WriteString("\nParameters:\n")
			} else {
				sb.WriteString("\nParameters when: ")
				sb.WriteString(w)
				sb.WriteString("\n")
			}

			paramsForWhen := allParams[w]

			// Sort all parameters alphabetically
			sort.Slice(paramsForWhen, func(i, j int) bool {
				return paramsForWhen[i].name < paramsForWhen[j].name
			})

			// Find max length for alignment
			maxLen := 0
			for _, p := range paramsForWhen {
				if len(p.name) > maxLen {
					maxLen = len(p.name)
				}
			}

			// Output all parameters
			for _, p := range paramsForWhen {
				sb.WriteString(fmt.Sprintf("  %-*s - %s\n", maxLen, p.name, p.description))
			}

		}

		sb.WriteString("\nNotes:\n")
		sb.WriteString("  - Additional parameters are supported\n")
		sb.WriteString("  - Array parameters: use [0], [1], [2] instead of [n]\n")
		sb.WriteString("  - Required parameters are enforced by API\n")
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
				log.Warnf("Could not find operation id: %s", opId)
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

func GetParameterDescription(aName string, aType string, usage string, aliasAttribute string, useStandardDescriptions bool) string {
	description := ""
	if usage != "" {
		description = usage
	} else if aType == "BOOL" {
		description = "A boolean value"
	} else if aType == "STRING" {
		description = "A string value"
		if useStandardDescriptions == true {
			switch aName {
			case "filter":
				description = "A filtering expression"
			case "include":
				description = "Related resources that can be included"
			}
		}

	} else if strings.HasPrefix(aType, "ENUM:") {
		description = "One of the following values: " + strings.ReplaceAll(strings.ReplaceAll(aType, "ENUM:", ""), ",", ", ")
		if useStandardDescriptions {
			switch aName {
			case "sort":
				description = "Controls the order of how records are returned, one of: " + strings.ReplaceAll(strings.ReplaceAll(aType, "ENUM:", ""), ",", ", ")
			}
		}
	} else if strings.HasPrefix(aType, "CONST:") {
		description = "Only: " + strings.ReplaceAll(strings.ReplaceAll(aType, "CONST:", ""), ",", ", ") + " (note: the epcc will auto-populate this if an adjacent attribute is set)"
	} else if aType == "INT" {
		description = "An integer value"

		if useStandardDescriptions {
			switch aName {
			case "page[limit]":
				description = "The number of items to return"
			case "page[offset]":
				description = "The offset (in number of records) to start from"
			}
		}
	} else if aType == "FLOAT" {
		description = "A floating point value"
	} else if aType == "URL" {
		description = "A url"
	} else if aType == "JSON_API_TYPE" {
		description = "A value that matches a `type` used by the API"
	} else if aType == "CURRENCY" {
		description = "A three letter currency code"
	} else if aType == "FILE" {
		description = "A filename"
	} else if aType == "PRIMITIVE" {
		description = "Any of an int, float, string, or boolean value"
	} else if aType == "SINGULAR_RESOURCE_TYPE" {
		description = "A resource name used by the epcc cli"
	} else if strings.HasPrefix(aType, "RESOURCE_ID") {
		resName := strings.ReplaceAll(aType, "RESOURCE_ID:", "")
		if res, ok := resources.GetResourceByName(resName); ok {
			attribute := "id"
			if aliasAttribute != "" {
				attribute = aliasAttribute
			}
			description = fmt.Sprintf("The %s of a %s resource", attribute, res.SingularName)
		} else {
			description = "A resource id for " + resName
		}
	} else {
		description = "Unknown:" + aType
	}

	return description
}

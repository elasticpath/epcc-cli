package cmd

/*
get-all: Export all instances of one or more resource types from an EPCC store.

Usage:
  epcc get-all <resource>                    # Single resource (uses subcommand)
  epcc get-all <resource1> <resource2> ...   # Multiple resources

Algorithm Overview:

1. CROSS-RESOURCE DEPENDENCY RESOLUTION (for multiple resources)
   When multiple resources are requested, we determine the order to process them:
   - Build a dependency graph based on URL templates (e.g., account-addresses depends on accounts
     because its URL is /v2/accounts/{accountId}/addresses)
   - Also consider RESOURCE_ID attribute dependencies and explicit export-depends-on declarations
     (for subtle dependencies not visible in URL structure, e.g., custom-fields -> custom-api-settings-entries)
   - Use topological sort to determine processing order
   - Process resources in dependency order so parent aliases exist before children reference them

2. PARENT RESOLUTION (per resource)
   Many EPCC resources are nested under parent resources (e.g., customer-addresses
   are under customers, entries are under flows). Before we can fetch the target
   resource, we must first discover all parent resource IDs. This is done recursively:
   - Parse the resource URL template to find parent types (e.g., /v2/customers/{customerId}/addresses)
   - For each parent type, recursively fetch all IDs using the same algorithm
   - This produces a list of "parent paths" - each path is a list of IDs leading to the target

3. PAGINATION
   For each parent path, we paginate through the target resource collection:
   - Fetch pages of 100 items at a time using page[limit] and page[offset]
   - Continue until we get an empty page or detect duplicate results (some endpoints don't paginate)
   - Send each page's raw JSON to the output processor via a channel

4. OUTPUT PROCESSING (runs concurrently)
   A goroutine receives pages and processes them according to the output format:
   - jsonl/json/csv: Transform and output the data directly
   - epcc-cli/epcc-cli-runbook: Generate `epcc create` commands to recreate resources

5. TOPOLOGICAL SORTING (for self-referential resources)
   Some resources reference other resources of the same type (e.g., hierarchical nodes
   with parent_id pointing to another node). For these:
   - Build a dependency graph as we process records
   - Use Kahn's algorithm to determine creation order (dependencies before dependents)
   - Output commands in stages where each stage can be run in parallel

Output Formats:
  - jsonl: One JSON object per line (default, streamable)
  - json:  Single JSON array containing all results
  - csv:   Flattened CSV with dot-notation headers
  - epcc-cli: Shell commands to recreate resources via `epcc create`
  - epcc-cli-runbook: Same as epcc-cli but formatted for runbook YAML
*/

import (
	"context"
	gojson "encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/elasticpath/epcc-cli/external/apihelper"
	"github.com/elasticpath/epcc-cli/external/clictx"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/id"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/elasticpath/epcc-cli/external/toposort"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag"
	"github.com/yukithm/json2csv"
	"github.com/yukithm/json2csv/jsonpointer"
)

type OutputFormat enumflag.Flag

const (
	Jsonl OutputFormat = iota
	Json
	Csv
	EpccCli
	EpccCliRunbook
)

var OutputFormatIds = map[OutputFormat][]string{
	Jsonl:          {"jsonl"},
	Json:           {"json"},
	Csv:            {"csv"},
	EpccCli:        {"epcc-cli"},
	EpccCliRunbook: {"epcc-cli-runbook"},
}

// outputFormatCompletionFunc provides tab completion for the --output-format flag
var outputFormatCompletionFunc = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"jsonl\tJSON Lines format (default, one object per line)",
		"json\tSingle JSON array with all results",
		"csv\tCSV format with flattened fields",
		"epcc-cli\tGenerate epcc create commands",
		"epcc-cli-runbook\tGenerate epcc create commands for runbook YAML",
	}, cobra.ShellCompDirectiveNoFileComp
}

func NewGetAllCommand(parentCmd *cobra.Command) func() {

	var outputFile string
	var outputFormat OutputFormat
	var truncateOutput bool

	// Note: Both the parent get-all command and each resource subcommand have RunE handlers.
	// This is intentional to support two usage patterns:
	// 1. Multi-resource mode: "epcc get-all accounts customers" - handled by parent RunE
	//    Cobra doesn't match "accounts" as a subcommand when followed by "customers"
	// 2. Single-resource mode: "epcc get-all accounts" - handled by subcommand RunE
	//    Allows resource-specific help and tab completion
	// Both ultimately call getAllInternal, but subcommands provide better UX for single resources.
	var getAll = &cobra.Command{
		Use:   "get-all [resource1] [resource2] ...",
		Short: "Get all of one or more resources",
		Long: `Get all instances of one or more resource types.

When multiple resources are specified, they are processed in dependency order
(parent resources before children) so that aliases are available for reference.

Examples:
  epcc get-all accounts                           # Get all accounts
  epcc get-all accounts account-addresses         # Get accounts then their addresses
  epcc get-all --output-format epcc-cli accounts  # Output as epcc create commands`,
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("please specify a resource, epcc get-all [RESOURCE...], see epcc get-all --help")
			}
			// This handles unknown resources or when called directly without subcommand routing
			return getAllInternal(clictx.Ctx, outputFormat, outputFile, truncateOutput, args)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// Complete with plural resource names that support GET collection
			return completion.Complete(completion.Request{
				Type: completion.CompletePluralResource,
				Verb: completion.Get, // Get verb checks for GetCollectionInfo
			})
		},
	}

	// Add flags to the root get-all command for multi-resource mode
	getAll.Flags().StringVarP(&outputFile, "output-file", "o", "", "The file to output results to")
	getAll.Flags().BoolVarP(&truncateOutput, "truncate-output", "t", false, "Truncate the output file before writing (instead of appending)")
	getAll.Flags().VarP(
		enumflag.New(&outputFormat, "output-format", OutputFormatIds, enumflag.EnumCaseInsensitive),
		"output-format", "f",
		"sets output format; can be 'jsonl', 'json', 'csv', 'epcc-cli', 'epcc-cli-runbook'")
	_ = getAll.RegisterFlagCompletionFunc("output-format", outputFormatCompletionFunc)

	for _, resource := range resources.GetPluralResources() {
		if resource.GetCollectionInfo == nil {
			continue
		}

		resourceName := resource.PluralName

		// Each subcommand gets its own flags
		var subOutputFile string
		var subOutputFormat OutputFormat
		var subTruncateOutput bool

		var getAllResourceCmd = &cobra.Command{
			Use:    resourceName,
			Short:  GetGetAllShort(resource),
			Hidden: false,
			RunE: func(cmd *cobra.Command, args []string) error {
				return getAllInternal(clictx.Ctx, subOutputFormat, subOutputFile, subTruncateOutput, append([]string{resourceName}, args...))
			},
			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				// Complete with additional plural resource names that support GET collection
				return completion.Complete(completion.Request{
					Type: completion.CompletePluralResource,
					Verb: completion.Get,
				})
			},
		}

		getAllResourceCmd.Flags().StringVarP(&subOutputFile, "output-file", "o", "", "The file to output results to")
		getAllResourceCmd.Flags().BoolVarP(&subTruncateOutput, "truncate-output", "t", false, "Truncate the output file before writing (instead of appending)")

		getAllResourceCmd.Flags().VarP(
			enumflag.New(&subOutputFormat, "output-format", OutputFormatIds, enumflag.EnumCaseInsensitive),
			"output-format", "f",
			"sets output format; can be 'jsonl', 'json', 'csv', 'epcc-cli', 'epcc-cli-runbook'")
		_ = getAllResourceCmd.RegisterFlagCompletionFunc("output-format", outputFormatCompletionFunc)

		getAll.AddCommand(getAllResourceCmd)
	}

	parentCmd.AddCommand(getAll)
	return func() {}

}

// sortResourcesByDependency orders resources so that parent resources come before children.
// Dependencies are determined by URL templates (e.g., account-addresses depends on accounts
// because its URL is /v2/accounts/{accountId}/addresses).
func sortResourcesByDependency(resourceList []resources.Resource) ([]resources.Resource, error) {
	if len(resourceList) <= 1 {
		return resourceList, nil
	}

	// Build a map of resource names for quick lookup
	requestedResources := make(map[string]resources.Resource)
	for _, r := range resourceList {
		requestedResources[r.PluralName] = r
		requestedResources[r.SingularName] = r
	}

	// Build dependency graph
	graph := toposort.NewGraph()

	for _, resource := range resourceList {
		graph.AddNode(resource.PluralName)

		// Get URL template dependencies
		if resource.GetCollectionInfo != nil {
			deps, err := resources.GetTypesOfVariablesNeeded(resource.GetCollectionInfo.Url)
			if err != nil {
				log.Warnf("Could not get URL dependencies for %s: %v", resource.PluralName, err)
				continue
			}

			for _, dep := range deps {
				// Check if this dependency is in our requested list
				if depResource, ok := requestedResources[dep]; ok {
					// Add edge: dependency -> resource (dependency must come first)
					graph.AddEdge(depResource.PluralName, resource.PluralName)
					log.Debugf("Resource %s depends on %s (URL template)", resource.PluralName, depResource.PluralName)
				}
			}
		}

		// Also check attribute-level RESOURCE_ID dependencies
		for attrName, attr := range resource.Attributes {
			if strings.HasPrefix(attr.Type, "RESOURCE_ID:") {
				depType := strings.TrimPrefix(attr.Type, "RESOURCE_ID:")
				if depResource, ok := requestedResources[depType]; ok {
					// Only add if not self-referential (that's handled separately)
					if depResource.PluralName != resource.PluralName {
						graph.AddEdge(depResource.PluralName, resource.PluralName)
						log.Debugf("Resource %s depends on %s (attribute %s)", resource.PluralName, depResource.PluralName, attrName)
					}
				}
			}
		}

		// Check explicit export dependencies (for subtle dependencies not visible in URL structure)
		if len(resource.ExportDependsOn) > 0 {
			log.Debugf("Resource %s has export-depends-on: %v", resource.PluralName, resource.ExportDependsOn)
		}
		for _, dep := range resource.ExportDependsOn {
			if depResource, ok := requestedResources[dep]; ok {
				graph.AddEdge(depResource.PluralName, resource.PluralName)
				log.Debugf("Resource %s depends on %s (explicit export-depends-on)", resource.PluralName, depResource.PluralName)
			} else {
				log.Debugf("Resource %s has export-depends-on %s but it's not in the requested list", resource.PluralName, dep)
			}
		}
	}

	// Topologically sort
	sortedNames, err := graph.TopologicalSort()
	if err != nil {
		return nil, fmt.Errorf("dependency cycle detected: %w", err)
	}

	// Convert back to Resource slice
	result := make([]resources.Resource, 0, len(sortedNames))
	for _, name := range sortedNames {
		if r, ok := requestedResources[name]; ok {
			// Avoid duplicates (since we added both plural and singular names)
			found := false
			for _, existing := range result {
				if existing.PluralName == r.PluralName {
					found = true
					break
				}
			}
			if !found {
				result = append(result, r)
			}
		}
	}

	return result, nil
}

// getResourceNames returns a slice of plural names for logging.
func getResourceNames(resourceList []resources.Resource) []string {
	names := make([]string, len(resourceList))
	for i, r := range resourceList {
		names[i] = r.PluralName
	}
	return names
}

func writeJson(obj interface{}, writer io.Writer) error {
	line, err := gojson.Marshal(&obj)

	if err != nil {
		return fmt.Errorf("could not create JSON for %s, error: %v", line, err)

	}

	_, err = writer.Write(line)

	if err != nil {
		return fmt.Errorf("could not save line %s, error: %v", line, err)

	}

	_, err = writer.Write([]byte{10})

	if err != nil {
		return fmt.Errorf("could not save line %s, error: %v", line, err)
	}

	return nil
}

func getAllInternal(ctx context.Context, outputFormat OutputFormat, outputFile string, truncateOutput bool, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no resources specified")
	}

	// Truncate output file if requested (do this once before processing any resources)
	if truncateOutput && outputFile != "" {
		if err := os.Truncate(outputFile, 0); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("could not truncate output file: %w", err)
		}
	}

	// Write shebang for epcc-cli format (shell script)
	if outputFormat == EpccCli && outputFile != "" {
		f, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0755)
		if err != nil {
			return fmt.Errorf("could not open output file for shebang: %w", err)
		}
		_, err = f.WriteString("#!/bin/bash\nset -e\n")
		f.Close()
		if err != nil {
			return fmt.Errorf("could not write shebang: %w", err)
		}
	}

	// Validate all resources and deduplicate
	seen := make(map[string]bool)
	resourceList := make([]resources.Resource, 0, len(args))

	for _, name := range args {
		resource, ok := resources.GetResourceByName(name)
		if !ok {
			return fmt.Errorf("could not find resource %s", name)
		}
		if resource.GetCollectionInfo == nil {
			return fmt.Errorf("resource %s doesn't support GET collection", name)
		}
		// Deduplicate by plural name
		if !seen[resource.PluralName] {
			seen[resource.PluralName] = true
			resourceList = append(resourceList, resource)
		} else {
			log.Debugf("Skipping duplicate resource %s", name)
		}
	}

	// Sort resources by dependency if there's more than one
	if len(resourceList) > 1 {
		var err error
		resourceList, err = sortResourcesByDependency(resourceList)
		if err != nil {
			return fmt.Errorf("could not sort resources by dependency: %w", err)
		}
		log.Infof("Processing %d resources in dependency order: %v", len(resourceList), getResourceNames(resourceList))
	}

	// Process each resource
	for i, resource := range resourceList {
		if len(resourceList) > 1 {
			log.Infof("Processing resource %d/%d: %s", i+1, len(resourceList), resource.PluralName)
		}

		err := getAllSingleResource(ctx, outputFormat, outputFile, resource)
		if err != nil {
			return fmt.Errorf("error processing resource %s: %w", resource.PluralName, err)
		}
	}

	// Make output file executable for epcc-cli format
	if outputFormat == EpccCli && outputFile != "" {
		if err := os.Chmod(outputFile, 0755); err != nil {
			return fmt.Errorf("could not make output file executable: %w", err)
		}
	}

	// Log success message
	if outputFile != "" {
		log.Infof("Successfully exported %d resource type(s) to %s", len(resourceList), outputFile)
	} else {
		log.Infof("Successfully exported %d resource type(s) to stdout", len(resourceList))
	}

	if outputFormat == EpccCli || outputFormat == EpccCliRunbook {
		log.Warnf("Output to EPCC CLI format is currently BETA, please report any bugs on GitHub")
	}

	return nil
}

// getAllSingleResource fetches all instances of a single resource type.
func getAllSingleResource(ctx context.Context, outputFormat OutputFormat, outputFile string, resource resources.Resource) error {
	allParentEntityIds, err := getParentIdsForGetAll(ctx, resource)

	if err != nil {
		return fmt.Errorf("could not retrieve parent ids for resource %s, error: %w", resource.PluralName, err)
	}

	if len(allParentEntityIds) == 1 {
		log.Debugf("Resource %s is a top level resource need to scan only one path to get all resources", resource.PluralName)
	} else {
		log.Debugf("Resource %s is not a top level resource, need to scan %d paths to get all resources", resource.PluralName, len(allParentEntityIds))
	}

	var syncGroup = sync.WaitGroup{}

	syncGroup.Add(1)

	type idableAttributesWithType struct {
		id.IdableAttributes
		Type        string `yaml:"type,omitempty" json:"type,omitempty"`
		EpccCliType string `yaml:"epcc_cli_type,omitempty" json:"epcc_cli_type,omitempty"`
	}

	type msg struct {
		txt []byte
		id  []idableAttributesWithType
	}
	var sendChannel = make(chan msg, 0)

	var writer io.Writer
	if outputFile == "" {
		writer = os.Stdout
	} else {
		file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return fmt.Errorf("could not open output file: %w", err)
		}
		defer file.Close()
		writer = file
	}

	topoSortNeeded := false

	topoSortKeys := make([]string, 0)
	for k, v := range resource.Attributes {
		if (v.Type == fmt.Sprintf("RESOURCE_ID:%s", resource.PluralName)) || (v.Type == fmt.Sprintf("RESOURCE_ID:%s", resource.SingularName)) {
			topoSortKeys = append(topoSortKeys, k)
			topoSortNeeded = true
		}

	}

	lines := map[string]string{}
	graph := toposort.NewGraph()

	outputWriter := func() {
		defer syncGroup.Done()

		csvLines := make([]interface{}, 0)

		if outputFormat == EpccCliRunbook && !topoSortNeeded {
			// We need to prefix
			_, err := writer.Write([]byte("- |\n"))

			if err != nil {
				log.Errorf("Error writing command: %v", err)
			}
		}

	endMessages:
		for msgs := 0; ; msgs++ {
			select {
			case result, ok := <-sendChannel:
				if !ok {
					log.Debugf("Channel closed, we are done.")
					break endMessages
				}
				var obj interface{}
				err = gojson.Unmarshal(result.txt, &obj)

				if err != nil {
					log.Errorf("Couldn't unmarshal JSON response %s due to error: %v", result, err)
					continue
				}

				newObjs, err := json.RunJQWithArray(".data[]", obj)

				if err != nil {
					log.Errorf("Couldn't process response %s due to error: %v", result, err)
					continue
				}

				// Check if this is an array-notation no-wrapping resource
				usesArrayNotation := false
				if resource.NoWrapping {
					for attrKey := range resource.Attributes {
						if strings.Contains(attrKey, "[n]") || strings.Contains(attrKey, "[0]") {
							usesArrayNotation = true
							break
						}
					}
				}

				// For array-notation no-wrapping resources with epcc-cli output,
				// aggregate all items into a single command with incrementing indices
				if usesArrayNotation && (outputFormat == EpccCli || outputFormat == EpccCliRunbook) && len(newObjs) > 0 {
					sb := &strings.Builder{}

					sb.WriteString("epcc create -s --skip-alias-processing ")
					sb.WriteString(resource.SingularName)

					// Use first item's ID for the alias (or create a composite alias)
					var firstId = ""
					if mp, ok := newObjs[0].(map[string]interface{}); ok {
						if id, ok := mp["id"].(string); ok {
							firstId = id
						}
					}

					sb.WriteString(" ")
					sb.WriteString("--save-as-alias")
					sb.WriteString(" ")
					sb.WriteString("exported_source_id=")
					sb.WriteString(firstId)

					// Add parent resource references
					for _, resId := range result.id {
						sb.WriteString(" ")
						sb.WriteString(resources.MustGetResourceByName(resId.EpccCliType).JsonApiType)
						sb.WriteString("/")
						sb.WriteString("exported_source_id=")
						sb.WriteString(resId.Id)
					}

					// Process each item with its index
					for itemIdx, newObj := range newObjs {
						kvs, err := json2csv.JSON2CSV(newObj)
						if err != nil {
							log.Errorf("Error generating Key/Value pairs for item %d: %v", itemIdx, err)
							continue
						}

						for _, kv := range kvs {
							keys := kv.Keys()
							sort.Strings(keys)

						nextKeyArray:
							for _, k := range keys {
								v := kv[k]

								jp, err := jsonpointer.New(k)
								if err != nil {
									log.Errorf("Couldn't generate JSON Pointer for %s: %v", k, err)
									continue
								}

								jsonPointerKey := jp.DotNotation(true)

								if strings.HasPrefix(jsonPointerKey, "meta.") {
									continue
								}
								if strings.HasPrefix(jsonPointerKey, "links.") {
									continue
								}

								// Skip timestamps and other read-only fields
								excludedPrefixes := []string{"created_at", "updated_at", "timestamps."}
								for _, prefix := range excludedPrefixes {
									if strings.HasPrefix(jsonPointerKey, prefix) {
										continue nextKeyArray
									}
								}

								// Skip resource-specific excluded JSON pointers
								for _, excluded := range resource.ExcludedJsonPointersFromImport {
									if strings.HasPrefix(jsonPointerKey, excluded) {
										continue nextKeyArray
									}
								}

								sb.WriteString(" ")
								// Use incrementing index for each item
								sb.WriteString(fmt.Sprintf("data[%d].", itemIdx))
								sb.WriteString(jsonPointerKey)
								sb.WriteString(" ")

								// Check if this attribute is a RESOURCE_ID type by looking up the attribute definition
								// Convert jsonPointerKey to the generic attribute key format (e.g., "id" -> "data[n].id")
								attrKey := "data[n]." + jsonPointerKey
								isResourceId := false
								if attr, ok := resource.Attributes[attrKey]; ok {
									if strings.HasPrefix(attr.Type, "RESOURCE_ID:") {
										isResourceId = true
									}
								}

								if s, ok := v.(string); ok {
									if isResourceId {
										// Use alias reference format for RESOURCE_ID attributes
										sb.WriteString(`"`)
										sb.WriteString("exported_source_id=")
										sb.WriteString(s)
										sb.WriteString(`"`)
									} else {
										sb.WriteString(`"`)
										quoteArgument := json.ValueNeedsQuotes(s)
										if quoteArgument {
											sb.WriteString("\\\"")
										}
										value := strings.ReplaceAll(s, `\`, `\\`)
										value = strings.ReplaceAll(value, `$`, `\$`)
										value = strings.ReplaceAll(value, `"`, `\"`)
										sb.WriteString(value)
										if quoteArgument {
											sb.WriteString("\\\"")
										}
										sb.WriteString(`"`)
									}
								} else {
									sb.WriteString(fmt.Sprintf("%v", v))
								}
							}
						}
					}

					sb.WriteString("\n")

					if outputFormat == EpccCliRunbook {
						_, err := writer.Write([]byte("  "))
						if err != nil {
							log.Errorf("Error writing command: %v", err)
						}
					}

					_, err = writer.Write([]byte(sb.String()))
					if err != nil {
						log.Errorf("Error writing command: %v", err)
					}

					// Still need to handle jsonl/json/csv for array notation resources
					for _, newObj := range newObjs {
						wrappedObj := map[string]interface{}{
							"data": newObj,
							"meta": map[string]interface{}{
								"_epcc_cli_parent_resources": result.id,
							},
						}
						if outputFormat == Jsonl {
							err = writeJson(wrappedObj, writer)
							if err != nil {
								log.Errorf("Error writing JSON line: %v", err)
							}
						} else if outputFormat == Json || outputFormat == Csv {
							csvLines = append(csvLines, wrappedObj)
						}
					}
					continue // Skip the per-item processing below
				}

				for _, newObj := range newObjs {

					wrappedObj := map[string]interface{}{
						"data": newObj,
						"meta": map[string]interface{}{
							"_epcc_cli_parent_resources": result.id,
						},
					}

					if outputFormat == Jsonl {
						err = writeJson(wrappedObj, writer)

						if err != nil {
							log.Errorf("Error writing JSON line: %v", err)
							continue
						}
					} else if outputFormat == Json || outputFormat == Csv {
						csvLines = append(csvLines, wrappedObj)
					} else if outputFormat == EpccCli || outputFormat == EpccCliRunbook {
						sb := &strings.Builder{}

						sb.WriteString("epcc create -s --skip-alias-processing ")
						sb.WriteString(resource.SingularName)

						sb.WriteString(" ")
						sb.WriteString("--save-as-alias")
						sb.WriteString(" ")
						sb.WriteString("exported_source_id=")

						var myId = ""
						if mp, ok := newObj.(map[string]interface{}); ok {
							// Try id at root level first, then under data
							if id, ok := mp["id"].(string); ok {
								myId = id
							} else if dataMap, ok := mp["data"].(map[string]interface{}); ok {
								if id, ok := dataMap["id"].(string); ok {
									myId = id
								}
							}
							sb.WriteString(myId)
						} else {
							log.Errorf("Error casting newObj to map[string]interface{}")
							sb.WriteString("\n")
							continue
						}

						if topoSortNeeded {
							graph.AddNode(myId)
						}

						for _, resId := range result.id {
							sb.WriteString(" ")
							sb.WriteString(resources.MustGetResourceByName(resId.EpccCliType).JsonApiType)
							sb.WriteString("/")
							sb.WriteString("exported_source_id=")
							sb.WriteString(resId.Id)

						}

						kvs, err := json2csv.JSON2CSV(newObj)
						if err != nil {
							log.Errorf("Error generating Key/Value pairs: %v", err)
							sb.WriteString("\n")
							continue
						}

						for _, kv := range kvs {

							keys := kv.Keys()

							sort.Strings(keys)

						nextKey:
							for _, k := range keys {
								v := kv[k]

								jp, err := jsonpointer.New(k)

								if err != nil {
									log.Errorf("Couldn't generate JSON Pointer for %s: %v", k, err)

									continue
								}

								jsonPointerKey := jp.DotNotation(true)

								if strings.HasPrefix(jsonPointerKey, "meta.") {
									continue
								}

								if strings.HasPrefix(jsonPointerKey, "links.") {
									continue
								}

								// Skip id fields (id, data.id, data[n].id) unless no-wrapping (where data.id is needed for relationships)
								if !resource.NoWrapping {
									if jsonPointerKey == "id" || strings.HasPrefix(jsonPointerKey, "data.id") ||
										strings.HasPrefix(jsonPointerKey, "data[") && strings.HasSuffix(jsonPointerKey, "].id") {
										continue
									}
								}

								// Skip type field unless no-wrapping (where data.type is needed)
								if jsonPointerKey == "type" && !resource.NoWrapping {
									continue
								}

								// Skip timestamps and other read-only fields
								excludedPrefixes := []string{"created_at", "updated_at", "timestamps."}
								for _, prefix := range excludedPrefixes {
									if strings.HasPrefix(jsonPointerKey, prefix) {
										continue nextKey
									}
								}

								// Skip resource-specific excluded JSON pointers
								for _, excluded := range resource.ExcludedJsonPointersFromImport {
									if strings.HasPrefix(jsonPointerKey, excluded) {
										continue nextKey
									}
								}

								sb.WriteString(" ")
								// For no-wrapping resources, we need to prefix keys with "data."
								// (array notation is handled separately above)
								if resource.NoWrapping {
									sb.WriteString("data.")
								}
								sb.WriteString(jsonPointerKey)
								sb.WriteString(" ")

								if s, ok := v.(string); ok {

									writeValueFromJson := true

									for _, topoKey := range topoSortKeys {
										if jsonPointerKey == topoKey {
											dependentId := fmt.Sprintf("%s", v)
											graph.AddEdge(dependentId, myId)
											writeValueFromJson = false
											sb.WriteString(`"`)
											sb.WriteString("exported_source_id=")
											sb.WriteString(dependentId)
											sb.WriteString(`"`)
										}
									}

									if writeValueFromJson {
										// This is to prevent shell characters from interpreting things
										sb.WriteString(`"`)

										quoteArgument := json.ValueNeedsQuotes(s)

										if quoteArgument {
											// This is to force the EPCC CLI to interpret the value as a string
											sb.WriteString("\\\"")
										}
										value := strings.ReplaceAll(s, `\`, `\\`)
										value = strings.ReplaceAll(value, `$`, `\$`)
										value = strings.ReplaceAll(value, `"`, `\"`)
										sb.WriteString(value)

										if quoteArgument {
											// This is to force the EPCC CLI to interpret the value as a string
											sb.WriteString("\\\"")
										}
										// This is to prevent shell characters from interpreting things
										sb.WriteString(`"`)
									}
								} else {
									sb.WriteString(fmt.Sprintf("%v", v))
								}

							}
						}

						sb.WriteString("\n")
						if topoSortNeeded {
							lines[myId] = sb.String()
						} else {
							if outputFormat == EpccCliRunbook {
								// We need to prefix
								_, err := writer.Write([]byte("  "))

								if err != nil {
									log.Errorf("Error writing command: %v", err)
								}
							}

							_, err = writer.Write([]byte(sb.String()))

							if err != nil {
								log.Errorf("Error writing command: %v", err)
							}
						}
					}
				}
			}
		}

		if outputFormat == Json {
			err = writeJson(csvLines, writer)

			if err != nil {
				log.Errorf("Error writing JSON line: %v", err)
			}
		} else if outputFormat == Csv {

			// Create writer that saves to string
			results, err := json2csv.JSON2CSV(csvLines)

			if err != nil {
				log.Errorf("Error converting to CSV: %v", err)
				return
			}

			csvWriter := json2csv.NewCSVWriter(writer)

			csvWriter.HeaderStyle = json2csv.DotBracketStyle
			csvWriter.Transpose = false

			if err := csvWriter.WriteCSV(results); err != nil {
				log.Errorf("Error writing CSV: %v", err)
				return
			}
		} else if (outputFormat == EpccCli || outputFormat == EpccCliRunbook) && topoSortNeeded {
			stages, err := graph.ParallelizableStages()

			if err != nil {
				log.Errorf("Error sorting data: %v", err)
				return
			}

			for idx, stage := range stages {
				writer.Write([]byte(fmt.Sprintf("# Stage %d\n", idx)))
				if outputFormat == EpccCliRunbook {
					writer.Write([]byte(fmt.Sprintf("- |\n")))
				}

				for _, id := range stage {
					if outputFormat == EpccCliRunbook {
						writer.Write([]byte(fmt.Sprintf("  ")))
					}

					_, err = writer.Write([]byte(lines[id]))

					if err != nil {
						log.Errorf("Error writing command: %v", err)
					}
				}
			}

		}

	}

	go outputWriter()

	for _, parentEntityIds := range allParentEntityIds {
		lastIds := make([][]id.IdableAttributes, 1)

		for offset := 0; ; offset += 100 {
			// Check if context has been cancelled (e.g., user pressed Ctrl+C)
			select {
			case <-ctx.Done():
				close(sendChannel)
				syncGroup.Wait()
				return ctx.Err()
			default:
			}

			if offset > 10000 {
				// Most pagination limits have a max offset of 10,000
				log.Warnf("Maximum pagination offset reached, could not retrieve all records")
				break
			}
			resourceURL, err := resources.GenerateUrlViaIdableAttributes(resource.GetCollectionInfo, parentEntityIds)

			if err != nil {
				return err
			}

			types, err := resources.GetSingularTypesOfVariablesNeeded(resource.GetCollectionInfo.Url)

			if err != nil {
				return err
			}

			params := url.Values{}
			params.Add("page[limit]", "100")
			params.Add("page[offset]", strconv.Itoa(offset))

			resp, err := httpclient.DoRequest(ctx, "GET", resourceURL, params.Encode(), nil)

			if err != nil {
				return err
			}

			if resp.StatusCode >= 400 {
				log.Warnf("Could not retrieve page of data, aborting")

				break
			}

			bodyTxt, err := io.ReadAll(resp.Body)

			if err != nil {

				return err
			}

			ids, totalCount, err := apihelper.GetResourceIdsFromBody(bodyTxt)
			resp.Body.Close()

			allIds := make([][]id.IdableAttributes, 0)
			for _, id := range ids {
				allIds = append(allIds, append(parentEntityIds, id))
			}

			if reflect.DeepEqual(allIds, lastIds) {
				log.Warnf("Data on the previous two pages did not change. Does this resource support pagination? Aborting export")

				break
			} else {
				lastIds = allIds
			}

			idsWithType := make([]idableAttributesWithType, len(types))

			for i, t := range types {
				idsWithType[i].IdableAttributes = parentEntityIds[i]
				idsWithType[i].EpccCliType = t
				idsWithType[i].Type = resources.MustGetResourceByName(t).JsonApiType
			}

			sendChannel <- msg{
				bodyTxt,
				idsWithType,
			}

			if len(allIds) == 0 {
				log.Infof("Total ids retrieved for %s in %s is %d, we are done", resource.PluralName, resourceURL, len(allIds))

				break
			} else {
				if totalCount >= 0 {
					log.Infof("Total number of %s in %s is %d", resource.PluralName, resourceURL, totalCount)
				} else {
					log.Infof("Total number %s in %s is unknown", resource.PluralName, resourceURL)
				}
			}

		}
	}

	close(sendChannel)

	syncGroup.Wait()

	return nil
}

// getParentIdsForGetAll retrieves all parent entity IDs for a resource.
// This is similar to getParentIds in delete-all.go but uses a default page length.
func getParentIdsForGetAll(ctx context.Context, resource resources.Resource) ([][]id.IdableAttributes, error) {
	const defaultPageLength uint16 = 25

	myEntityIds := make([][]id.IdableAttributes, 0)
	if resource.GetCollectionInfo == nil {
		return myEntityIds, fmt.Errorf("resource %s doesn't support GET collection", resource.PluralName)
	}

	types, err := resources.GetTypesOfVariablesNeeded(resource.GetCollectionInfo.Url)

	if err != nil {
		return myEntityIds, err
	}

	if len(types) == 0 {
		myEntityIds = append(myEntityIds, make([]id.IdableAttributes, 0))
		return myEntityIds, nil
	} else {
		immediateParentType := types[len(types)-1]

		parentResource, ok := resources.GetResourceByName(immediateParentType)

		if !ok {
			return myEntityIds, fmt.Errorf("could not find parent resource %s", immediateParentType)
		}

		return apihelper.GetAllIds(ctx, defaultPageLength, &parentResource)
	}
}

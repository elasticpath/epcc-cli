package cmd

import (
	"context"
	gojson "encoding/json"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/apihelper"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/id"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/yukithm/json2csv"
	"github.com/yukithm/json2csv/jsonpointer"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/elasticpath/epcc-cli/external/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag"
	"io"
	"net/url"
	"reflect"
	"strconv"
)

type OutputFormat enumflag.Flag

const (
	Jsonl OutputFormat = iota
	Json
	Csv
	EpccCli
)

var OutputFormatIds = map[OutputFormat][]string{
	Jsonl:   {"jsonl"},
	Json:    {"json"},
	Csv:     {"csv"},
	EpccCli: {"epcc-cli"},
}

func NewGetAllCommand(parentCmd *cobra.Command) func() {

	var getAll = &cobra.Command{
		Use:          "get-all",
		Short:        "Get all of a resource",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("please specify a resource, epcc get-all [RESOURCE], see epcc delete-all --help")
			} else {
				return fmt.Errorf("invalid resource [%s] specified, see all with epcc delete-all --help", args[0])
			}
		},
	}

	for _, resource := range resources.GetPluralResources() {
		if resource.GetCollectionInfo == nil {
			continue
		}

		resourceName := resource.PluralName

		var outputFile string
		var outputFormat OutputFormat

		var getAllResourceCmd = &cobra.Command{
			Use:    resourceName,
			Short:  GetGetAllShort(resource),
			Hidden: false,
			RunE: func(cmd *cobra.Command, args []string) error {
				return getAllInternal(context.Background(), outputFormat, outputFile, append([]string{resourceName}, args...))
			},
		}

		getAllResourceCmd.Flags().StringVarP(&outputFile, "output-file", "", "", "The file to output results to")

		getAllResourceCmd.Flags().VarP(
			enumflag.New(&outputFormat, "output-format", OutputFormatIds, enumflag.EnumCaseInsensitive),
			"output-format", "",
			"sets output format; can be 'jsonl', 'csv', 'epcc-cli'")

		getAll.AddCommand(getAllResourceCmd)
	}

	parentCmd.AddCommand(getAll)
	return func() {}

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
		return fmt.Errorf("Could not save line %s, error: %v", line, err)
	}

	return nil
}

func getAllInternal(ctx context.Context, outputFormat OutputFormat, outputFile string, args []string) error {
	// Find Resource
	resource, ok := resources.GetResourceByName(args[0])
	if !ok {
		return fmt.Errorf("could not find resource %s", args[0])
	}

	if resource.GetCollectionInfo == nil {
		return fmt.Errorf("resource %s doesn't support GET collection", args[0])
	}

	allParentEntityIds, err := getParentIds(ctx, resource)

	if err != nil {
		return fmt.Errorf("could not retrieve parent ids for for resource %s, error: %w", resource.PluralName, err)
	}

	if len(allParentEntityIds) == 1 {
		log.Debugf("Resource %s is a top level resource need to scan only one path to delete all resources", resource.PluralName)
	} else {
		log.Debugf("Resource %s is not a top level resource, need to scan %d paths to delete all resources", resource.PluralName, len(allParentEntityIds))
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
			panic(err)
		}
		defer file.Close()
		writer = file
	}

	outputWriter := func() {
		defer syncGroup.Done()

		csvLines := make([]interface{}, 0)

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
					} else if outputFormat == EpccCli {
						sb := &strings.Builder{}

						sb.WriteString("epcc create ")
						sb.WriteString(resource.SingularName)

						sb.WriteString(" ")
						sb.WriteString("--save-as-alias")
						sb.WriteString(" ")
						sb.WriteString("exported_source_id=")
						sb.WriteString(resource.SingularName)
						sb.WriteString("/")

						if mp, ok := newObj.(map[string]interface{}); ok {
							sb.WriteString(fmt.Sprintf("%s", mp["id"]))
						} else {
							log.Errorf("Error casting newObj to map[string]interface{}")
							sb.WriteString("\n")
							continue
						}

						for _, resId := range result.id {
							sb.WriteString(" ")
							sb.WriteString("exported_source_id=")
							sb.WriteString(resources.MustGetResourceByName(resId.EpccCliType).SingularName)
							sb.WriteString("/")
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

								if jsonPointerKey == "id" {
									continue
								}

								if jsonPointerKey == "type" {
									continue
								}

								sb.WriteString(" ")
								sb.WriteString(jsonPointerKey)
								sb.WriteString(" ")

								if s, ok := v.(string); ok {
									sb.WriteString(`'`)
									sb.WriteString(strings.ReplaceAll(s, `'`, `\'`))
									sb.WriteString(`'`)
								} else {
									sb.WriteString(fmt.Sprintf("%v", v))
								}

							}
						}

						sb.WriteString("\n")
						_, err = writer.Write([]byte(sb.String()))

						if err != nil {
							log.Errorf("Error writing command: %v", err)
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
		}

	}

	go outputWriter()

	for _, parentEntityIds := range allParentEntityIds {
		lastIds := make([][]id.IdableAttributes, 1)
		for offset := 0; offset <= 10000; offset += 100 {
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

			ids, totalCount, err := apihelper.GetResourceIdsFromHttpResponse(bodyTxt)
			resp.Body.Close()

			allIds := make([][]id.IdableAttributes, 0)
			for _, id := range ids {
				allIds = append(allIds, append(parentEntityIds, id))
			}

			if reflect.DeepEqual(allIds, lastIds) {
				log.Warnf("Data on the previous two pages did not change. Does this resource support pagination? Aborting export", resource.PluralName, len(allIds))

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

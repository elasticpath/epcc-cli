package runbooks

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/elasticpath/epcc-cli/external/templates"
)

func CreateMapForRunbookArgumentPointers(runbookAction *RunbookAction) map[string]*string {
	runbookStringArguments := map[string]*string{}

	for key, value := range runbookAction.Variables {
		var s = templates.Render(value.Default)
		runbookStringArguments[key] = &s
	}
	return runbookStringArguments
}

func RenderTemplates(templateName string, rawCmd string, stringVars map[string]*string, variableDefinitions map[string]Variable) ([]string, error) {
	tpl, err := template.New(templateName).Funcs(sprig.FuncMap()).Funcs(templates.AddlFuncs).Parse(rawCmd)

	if err != nil {
		// Handle this case better
		return nil, err
	}

	var renderedTpl bytes.Buffer

	data := map[string]interface{}{}
	for key, val := range stringVars {
		if variableDef, ok := variableDefinitions[key]; ok {

			if variableDef.Type == "INT" {
				parsedVal, err := strconv.Atoi(*val)

				if err != nil {
					return nil, fmt.Errorf("error processing variable %s, value %v is not an integer: %w", key, *val, err)
				}
				data[key] = parsedVal
			} else if variableDef.Type == "STRING" {
				data[key] = val
			} else if strings.HasPrefix(variableDef.Type, "RESOURCE_ID:") {
				data[key] = val
			} else if strings.HasPrefix(variableDef.Type, "ENUM:") {
				// ENUM types are treated as strings, validate the value is one of the enum options
				enumValues := strings.Split(variableDef.Type[5:], ",")
				validValue := false
				for _, enumVal := range enumValues {
					if *val == enumVal {
						validValue = true
						break
					}
				}
				if !validValue {
					return nil, fmt.Errorf("error processing variable %s, value %q is not a valid enum option. Valid options are: [%s]", key, *val, strings.Join(enumValues, ", "))
				}
				data[key] = val
			} else {
				return nil, fmt.Errorf("error processing variable %s, unknown type [%s] specified in template", key, variableDef.Type)
			}

		} else {
			return nil, fmt.Errorf("undefined variable %s", key)
		}
	}

	err = tpl.Execute(&renderedTpl, data)

	if err != nil {
		return nil, fmt.Errorf("error executing template: %w", err)
	}

	// This algorithm is broken if you have an escaped "\n"
	// Remove line continuation characters
	tmpl := strings.ReplaceAll(renderedTpl.String(), "\\\n", "")
	rawCmdLines := strings.Split(tmpl, "\n")
	return rawCmdLines, nil
}

package runbooks

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/elasticpath/epcc-cli/external/templates"
	"math"
	"strconv"
	"strings"
	"text/template"
)

func CreateMapForRunbookArgumentPointers(runbookAction *RunbookAction) map[string]*string {
	runbookStringArguments := map[string]*string{}

	for key, value := range runbookAction.Variables {
		var s = value.Default
		runbookStringArguments[key] = &s
	}
	return runbookStringArguments
}

func RenderTemplates(templateName string, rawCmd string, stringVars map[string]*string, variableDefinitions map[string]Variable) ([]string, error) {
	tpl, err := template.New(templateName).Funcs(sprig.FuncMap()).Funcs(
		map[string]any{
			"pow":                func(a, b int) int { return int(math.Pow(float64(a), float64(b))) },
			"pseudoRandAlphaNum": templates.RandAlphaNum,
			"pseudoRandAlpha":    templates.RandAlpha,
			"pseudoRandNumeric":  templates.RandNumeric,
			"pseudoRandString":   templates.RandString,
			"pseudoRandInt":      templates.RandInt,
		}).Parse(rawCmd)

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
			} else {
				return nil, fmt.Errorf("error processing variable %s, unknown type [%s] specified in template", key, variableDef.Type)
			}

		} else {
			return nil, fmt.Errorf("undefined variable %s", key)
		}
	}

	err = tpl.Execute(&renderedTpl, data)

	if err != nil {
		return nil, err
	}

	// This algorithm is broken if you have an escaped "\n"
	rawCmdLines := strings.Split(renderedTpl.String(), "\n")
	return rawCmdLines, nil
}

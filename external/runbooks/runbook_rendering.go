package runbooks

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"math"
	"math/rand"
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
			"pseudoRandAlphaNum": randAlphaNum,
			"pseudoRandAlpha":    randAlpha,
			"pseudoRandNumeric":  randNumeric,
			"pseudoRandString":   randString,
			"pseudoRandInt":      randInt,
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
				data[strings.ReplaceAll(key, "-", "_")] = parsedVal
			} else if variableDef.Type == "STRING" {
				data[key] = val
				data[strings.ReplaceAll(key, "-", "_")] = val
			} else if strings.HasPrefix(variableDef.Type, "RESOURCE_ID:") {
				data[key] = val
				data[strings.ReplaceAll(key, "-", "_")] = val
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

// randString is the internal function that generates a random string.
// It takes the length of the string and a string of allowed characters as parameters.
func randString(letters string, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// randAlphaNum generates a string consisting of characters in the range 0-9, a-z, and A-Z.
func randAlphaNum(n int) string {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return randString(letters, n)
}

// randAlpha generates a string consisting of characters in the range a-z and A-Z.
func randAlpha(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return randString(letters, n)
}

// randNumeric generates a string consisting of characters in the range 0-9.
func randNumeric(n int) string {
	const digits = "0123456789"
	return randString(digits, n)
}

func randInt(min, max int) int {
	return rand.Intn(max-min) + min
}

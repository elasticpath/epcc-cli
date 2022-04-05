package resources

import (
	"fmt"
	"github.com/yosida95/uritemplate/v3"
)

func GenerateUrl(url string, args []string) (string, error) {
	template, err := uritemplate.New(url)

	if err != nil {
		return "", fmt.Errorf("Could not generate URI template for URL: %w", err)
	}

	vars := template.Varnames()

	if len(vars) > len(args) {
		return "", fmt.Errorf("URI Template requires %d arguments, but only %d were passed", len(vars), len(args))
	}

	values := uritemplate.Values{}

	for idx, varName := range vars {
		values[varName] = uritemplate.String(args[idx])
	}

	return template.Expand(values)
}

func GetNumberOfVariablesNeeded(url string) (int, error) {
	template, err := uritemplate.New(url)

	if err != nil {
		return 0, fmt.Errorf("Could not generate URI template for URL: %w", err)
	}

	return len(template.Varnames()), nil
}

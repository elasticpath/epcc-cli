package templates

import (
	"bytes"
	"math"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/elasticpath/epcc-cli/config"
	log "github.com/sirupsen/logrus"
)

func init() {

}

var AddlFuncs = map[string]any{
	"pow":                   func(a, b int) int { return int(math.Pow(float64(a), float64(b))) },
	"pseudoRandAlphaNum":    RandAlphaNum,
	"pseudoRandAlpha":       RandAlpha,
	"pseudoRandNumeric":     RandNumeric,
	"pseudoRandString":      RandString,
	"pseudoRandInt":         RandInt,
	"pseudoRandNorm":        RandNorm,
	"weightDatedTimeSample": WeightedDateTimeSampler,
	"nRandInt":              NRandInt,
	"fake":                  Fake,
	"seed":                  Seed,
	"formatPrice":           FormatPrice,
}

func Render(templateString string) string {

	if config.GetEnv().EPCC_CLI_DISABLE_TEMPLATE_EXECUTION {
		return templateString
	}

	if !strings.Contains(templateString, "{{") {
		return templateString
	}

	tpl, err := template.New("templateName").Funcs(sprig.FuncMap()).Funcs(AddlFuncs).Parse(templateString)

	if err != nil {
		log.Warnf("Could not process argument template: %s, due to %v", templateString, err)
		return templateString
	}

	var renderedTpl bytes.Buffer

	err = tpl.Execute(&renderedTpl, nil)

	if err != nil {
		log.Warnf("Could not process argument template: %s, due to %v", templateString, err)
		return templateString
	}

	return renderedTpl.String()
}

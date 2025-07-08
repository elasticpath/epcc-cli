package templates

import (
	"bytes"
	"github.com/Masterminds/sprig/v3"
	"github.com/elasticpath/epcc-cli/config"
	log "github.com/sirupsen/logrus"
	"strings"
	"text/template"
)

func init() {

}
func Render(templateString string) string {

	if config.GetEnv().EPCC_CLI_DISABLE_TEMPLATE_EXECUTION {
		return templateString
	}

	if !strings.Contains(templateString, "{{") {
		return templateString
	}

	tpl, err := template.New("templateName").Funcs(sprig.FuncMap()).Funcs(
		map[string]any{
			"pseudoRandAlphaNum": RandAlphaNum,
			"pseudoRandAlpha":    RandAlpha,
			"pseudoRandNumeric":  RandNumeric,
			"pseudoRandString":   RandString,
			"pseudoRandInt":      RandInt,
		}).Parse(templateString)

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

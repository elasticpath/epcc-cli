package templates

import (
	"bytes"
	"github.com/Masterminds/sprig/v3"
	log "github.com/sirupsen/logrus"
	"strings"
	"text/template"
)

func init() {

}
func Render(templateString string) string {

	if !strings.Contains(templateString, "{{") {
		return templateString
	}

	tpl, err := template.New("templateName").Funcs(sprig.FuncMap()).Parse(templateString)

	if err != nil {
		log.Warn("Could not process argument template: %s, due to %v", templateString, err)
		return templateString
	}

	var renderedTpl bytes.Buffer

	err = tpl.Execute(&renderedTpl, nil)

	if err != nil {
		log.Warn("Could not process argument template: %s, due to %v", templateString, err)
		return templateString
	}

	return renderedTpl.String()
}

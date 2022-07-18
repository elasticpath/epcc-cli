package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"html/template"
	"os"
)

const testTemplate = `#!/usr/bin/env bats
{{if .GetCollectionInfo -}}
@test "{{ .PluralName }} empty get collection" {
	run epcc get {{ .PluralName }}
	[ $status -eq 0 ]
}
{{end}}
`

var MakeTests = &cobra.Command{
	Use:    "make-tests",
	Short:  "Make a bunch of BATS tests",
	Hidden: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		err := os.MkdirAll("tests/resources", 0755)
		if err != nil {
			return err
		}

		for _, resourceName := range resources.GetPluralResourceNames() {
			testFilename := fmt.Sprintf("./tests/resources/%s.bats", resourceName)
			_, err := os.Stat(testFilename)

			if os.IsNotExist(err) {
				log.Infof("Tests for %s do not exist", testFilename)
			} else {
				log.Infof("Tests for %s exist already", testFilename)
			}

			tmpl, err := template.New("test").Parse(testTemplate)

			if err != nil {
				return err
			}

			resource, ok := resources.GetResourceByName(resourceName)

			if !ok {
				panic("Could not find resource for " + resourceName)
			}

			f, err := os.OpenFile(testFilename, os.O_CREATE+os.O_TRUNC+os.O_WRONLY, 0755)

			defer f.Close()
			if err != nil {
				return err
			}

			err = tmpl.Execute(f, resource)

			if err != nil {
				return err
			}

		}

		return nil
	},
}

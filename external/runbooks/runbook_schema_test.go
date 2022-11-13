package runbooks

import (
	"fmt"
	"github.com/santhosh-tekuri/jsonschema/v5"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestJsonSchemaValidate(t *testing.T) {
	sch, err := jsonschema.Compile("runbook_schema.json")
	if err != nil {
		log.Fatalf("%#v", err)
	}

	entries, err := embeddedRunbooks.ReadDir(".")

	if err != nil {
		log.Errorf("Could not load any built-in runbooks due to error reading names: %v. This is almost certainly a bug.", err)
		return
	}

	for _, entry := range entries {

		fileContents, err := embeddedRunbooks.ReadFile(entry.Name())
		if err != nil {
			log.Fatal(err)
		}

		var v interface{}
		if err := yaml.Unmarshal(fileContents, &v); err != nil {
			require.NoError(t, fmt.Errorf("Could not validate JSON Schema %s, %w", entry.Name(), err))
		}

		if err = sch.Validate(v); err != nil {
			require.NoError(t, fmt.Errorf("Could not validate JSON Schema %s, %w", entry.Name(), err))
		}
	}

}

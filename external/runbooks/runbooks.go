package runbooks

import (
	"embed"
	"fmt"
	"github.com/elasticpath/epcc-cli/config"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"sort"
	"strings"
)

//go:embed *.epcc.yml
var embeddedRunbooks embed.FS

var runbooks map[string]Runbook

type Runbook struct {

	// The name of the runbook
	Name string `yaml:"name"`

	Description *RunbookDescription `yaml:"description"`

	// The location of relevant documentation
	Docs string `yaml:"docs"`

	RunbookActions map[string]*RunbookAction `yaml:"actions"`
}

type RunbookAction struct {
	Name         string
	Description  *RunbookDescription
	RawCommands  []string `yaml:"commands"`
	IgnoreErrors bool     `yaml:"ignore_errors"`

	Variables map[string]Variable `yaml:"variables"`
}

type Variable struct {
	Name        string
	Type        string `yaml:"type"`
	Default     string `yaml:"default"`
	Description *RunbookDescription
}

type RunbookDescription struct {
	Long  string `yaml:"long"`
	Short string `yaml:"short"`
}

func init() {
	runbooks = make(map[string]Runbook)
}

func InitializeBuiltInRunbooks() {
	LoadBuiltInRunbooks(embeddedRunbooks)
	if config.Envs.EPCC_RUNBOOK_DIRECTORY != "" {
		LoadRunbooksFromDirectory(config.Envs.EPCC_RUNBOOK_DIRECTORY)
	}

}

func LoadRunbooksFromDirectory(dir string) {

	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Warnf("Could not read Runbooks from directory %s due to error: %v", dir, err)
		return
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			log.Warnf("Could not read Runbooks from entry %v due to error: %v", entry, err)
			continue
		}

		filename := path.Clean(fmt.Sprintf("%s/%s", dir, info.Name()))

		lFilename := strings.ToLower(filename)
		if strings.HasSuffix(lFilename, ".epcc.yml") || strings.HasSuffix(lFilename, ".epcc.yaml") {
			contents, err := os.ReadFile(filename)

			if err != nil {
				log.Warnf("Could not read Runbooks from file %s due to error: %v", filename, err)
				continue
			}

			err = AddRunbookFromYaml(string(contents))

			if err != nil {
				log.Warnf("Could not read Runbooks from file %s due to error: %v", filename, err)
				continue
			}
		} else {
			log.Tracef("File %s does not end in .epcc.yml, not parsing.", filename)
		}

	}
}

func GetRunbookNames() []string {
	keys := make([]string, 0, len(runbooks))

	for key := range runbooks {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}

func GetRunbooks() map[string]Runbook {
	runbookCopy := make(map[string]Runbook, 0)

	for key, val := range runbooks {
		runbookCopy[key] = val
	}

	return runbooks
}

func AddRunbookFromYaml(yaml string) error {

	runbook, err := loadRunbookFromString(yaml)
	if err != nil {
		return err
	}

	err = validateRunbook(runbook)
	if err != nil {
		return err
	} else {

		log.Tracef("Loaded runbook: %s, %s", runbook.Name, yaml)

		if _, ok := runbooks[runbook.Name]; ok {
			log.Warnf("Runbook name %s already loaded and is being overriden", runbook.Name)
		}

		runbooks[runbook.Name] = *runbook
	}

	return nil
}

func LoadBuiltInRunbooks(fs embed.FS) {
	entries, err := fs.ReadDir(".")

	if err != nil {
		log.Errorf("Could not load any built-in runbooks due to error reading names: %v. This is almost certainly a bug.", err)
		return
	}

	for _, entry := range entries {
		log.Tracef("Loading runbook: %s", entry.Name())

		fileContents, err := fs.ReadFile(entry.Name())
		if err != nil {
			log.Errorf("Could not load built-in runbook %s due to error reading data: %v. This is almost certainly a bug.", entry.Name(), err)
		}

		err = AddRunbookFromYaml(string(fileContents))

		if err != nil {
			log.Errorf("Could not load built-in runbook %s due to error: %v. This is almost certainly a bug.", entry.Name(), err)
		}
	}
}

func loadRunbookFromString(fileContents string) (*Runbook, error) {
	runbook := new(Runbook)
	err := yaml.Unmarshal([]byte(fileContents), &runbook)

	if err != nil {
		return nil, err

	}
	postProcessRunbook(runbook)

	return runbook, nil
}

func postProcessRunbook(runbook *Runbook) {
	for key := range runbook.RunbookActions {

		runbook.RunbookActions[key].Name = key
		if runbook.RunbookActions[key].Description == nil {
			runbook.RunbookActions[key].Description = &RunbookDescription{
				Long:  "",
				Short: "",
			}
		}
		for k2, variable := range runbook.RunbookActions[key].Variables {
			variable.Name = k2
			runbook.RunbookActions[key].Variables[k2] = variable
		}

	}
}

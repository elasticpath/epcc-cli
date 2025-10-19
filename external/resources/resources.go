package resources

import (
	"embed"
	_ "embed"
	"fmt"
	"io/fs"
	"regexp"
	"strings"

	"github.com/elasticpath/epcc-cli/config"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

//go:embed yaml/*.yaml
var resourceFiles embed.FS

var resources map[string]Resource

var resourcesSingular = map[string]Resource{}

// This will match a /v2/<FOO>/{BAR} with an optional slash at the end.

var topLevelResourceRegexp = regexp.MustCompile("^/v2/[^/]+/\\{[^}]+}/?$")

type Resource struct {
	// The type as far as the EPCC CLI is concerned.
	Type string `yaml:"-"`

	// A link to the generic documentation page about a type in the EPCC API
	Docs string `yaml:"docs"`

	// The type that should be used in the JSON API.
	JsonApiType string `yaml:"json-api-type"`

	// Alterative types used for aliases
	AlternateJsonApiTypesForAliases []string `yaml:"alternate-json-type-for-aliases,omitempty"`

	// Some resources (e.g., PCM, accelerator svc, bury most attributes under the attributes key). This is considered "compliant", other services just bury attributes under data, this is "legacy.
	JsonApiFormat string `yaml:"json-api-format"`

	// Information about how to get a collection
	GetCollectionInfo *CrudEntityInfo `yaml:"get-collection,omitempty"`

	// Information about how to get a single object.
	GetEntityInfo *CrudEntityInfo `yaml:"get-entity,omitempty"`

	// Information about how to create an entity.
	CreateEntityInfo *CrudEntityInfo `yaml:"create-entity,omitempty"`

	// Information about how to update an entity.
	UpdateEntityInfo *CrudEntityInfo `yaml:"update-entity,omitempty"`

	// Information about how to delete an entity.
	DeleteEntityInfo *CrudEntityInfo `yaml:"delete-entity,omitempty"`

	Attributes map[string]*CrudEntityAttribute `yaml:"attributes,omitempty"`

	// If true, don't wrap json in a data tag
	NoWrapping bool `yaml:"no-wrapping,omitempty"`

	// The singular name version of the resource.
	SingularName string `yaml:"singular-name"`

	PluralName string `yaml:"-"`

	// Use this value to silence warnings about a resource not supporting resets.
	// This should only be used for cases where we manually fix things, or where
	// a store reset would clear a resource another way (e.g., the resource represents a projection).
	SuppressResetWarning bool `yaml:"suppress-reset-warning,omitempty"`

	Legacy bool `yaml:"legacy"`

	// If another resource is used to create this resource, list it here
	CreatedBy []VerbResource `yaml:"created_by,omitempty"`

	// Source Filename
	SourceFile string
}

type QueryParameter struct {
	// The name of the query parameter
	Name string `yaml:"name"`
}

type VerbResource struct {
	Verb     string `yaml:"verb"`
	Resource string `yaml:"resource"`
}

type CrudEntityInfo struct {

	// A link to the docs specific for the Crud operation in EPCC.
	Docs string `yaml:"docs"`

	// The Url we should use when invoking this method.
	Url string `yaml:"url"`

	// Content type to send
	ContentType string `yaml:"content-type,omitempty"`

	// A list of valid query parameters
	QueryParameters []QueryParameter `yaml:"query,omitempty"`

	// Minimum resources so we don't keep trying to delete in
	MinResources int `yaml:"min,omitempty"`

	// Override the attribute we use in the URL for a specific key
	ParentResourceValueOverrides map[string]string `yaml:"parent_resource_value_overrides,omitempty"`

	OpenApiOperationId string `yaml:"openapi-operation-id"`

	// Only valid on create, if set we report that the type created by this is different.
	Creates string `yaml:"creates"`
}

type CrudEntityAttribute struct {
	// The name of the attribute
	Key string `yaml:"-"`

	// The type of the attribute
	Type           string `yaml:"type"`
	Usage          string `yaml:"usage"`
	AutoFill       string `yaml:"autofill,omitempty"`
	AliasAttribute string `yaml:"alias_attribute,omitempty"`

	// Expr for when the attribute is enabled
	When string `yaml:"when,omitempty"`
}

func GetPluralResourceNames() []string {
	keys := make([]string, 0, len(resources))

	for key := range resources {
		keys = append(keys, key)
	}

	return keys
}

func GetPluralResources() map[string]Resource {
	return resources
}

func GetSingularResourceNames() []string {
	keys := make([]string, 0, len(resourcesSingular))

	for key := range resourcesSingular {
		keys = append(keys, key)
	}

	return keys
}

func GetResourceByName(name string) (Resource, bool) {
	res, ok := resources[name]

	if ok {
		return res, true
	}

	res, ok = resourcesSingular[name]

	if ok {
		return res, true
	}

	return Resource{}, false
}

func MustGetResourceByName(name string) Resource {
	res, ok := GetResourceByName(name)

	if !ok {
		panic("Could not find resource: " + name)
	}

	return res
}

func GetSingularResourceByName(name string) (Resource, bool) {
	res, ok := resourcesSingular[name]

	if ok {
		return res, true
	}

	return Resource{}, false
}

func GenerateResourceMetadataFromYaml() (map[string]Resource, error) {
	resources := make(map[string]Resource)

	entries, err := fs.ReadDir(resourceFiles, "yaml")

	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		yamlBytes, err := resourceFiles.ReadFile("yaml/" + entry.Name())

		if err != nil {
			return nil, fmt.Errorf("Couldn't read resource.yaml %s: %w", entry.Name(), err)
		}

		r := map[string]Resource{}
		err = yaml.Unmarshal(yamlBytes, &r)
		if err != nil {
			return nil, fmt.Errorf("Couldn't unmarshal resource.yaml %s: %w", entry.Name(), err)
		}

		for k, v := range r {

			if _, ok := resources[k]; ok {
				return nil, fmt.Errorf("Duplicate resource %s", k)
			}
			v.SourceFile = entry.Name()
			log.Infof("Loaded %s from %s", k, entry.Name())
			resources[k] = v
		}
	}

	if err != nil {
		return nil, err
	}

	return resources, nil
}

func AppendResourceData(newResources map[string]Resource) {
	resourceCount := len(resources)
	for key, val := range newResources {
		resources[key] = val
	}

	log.Tracef("Loading %d new resources, total resources went from %d to %d ", len(newResources), resourceCount, len(resources))

	postProcessResourceMetadata()
}

func PublicInit() {
	if resources != nil {
		return
	}
	resourceData, err := GenerateResourceMetadataFromYaml()

	if err != nil {
		panic("Couldn't load the resource meta data: " + err.Error())
	}
	e := config.GetEnv()

	if e.EPCC_DISABLE_LEGACY_RESOURCES {
		for k, v := range resourceData {
			if v.Legacy {
				delete(resourceData, k)
			}
		}
	}

	resources = resourceData

	createFlowEntityRelationships()
	postProcessResourceMetadata()
}

func postProcessResourceMetadata() {
	resourcesSingular = make(map[string]Resource)

	for key, val := range resources {
		// Fix the key
		val.Type = key

		val.PluralName = key
		for attributeName, attributeVal := range val.Attributes {
			// Fix the key
			attributeVal.Key = attributeName
		}
		resourcesSingular[val.SingularName] = val
		resources[key] = val
	}

}

func createFlowEntityRelationships() {
	// Warning when this function runs the resources are in a "half"
	// initialized state (e.g., some fields like type might be null)
	resourcesAdded := make([]string, 0)
	for key, val := range resources {

		// This check isn't perfect, these resources exist no matter what.
		// However this seems like a good first attempt.
		// A slightly better attempt
		if val.GetEntityInfo == nil || !topLevelResourceRegexp.MatchString(val.GetEntityInfo.Url) {
			continue
		}
		newResource := Resource{
			Type:              key + "-entity-relationships",
			Docs:              "https://elasticpath.dev/docs/api/flows/entry-relationships",
			JsonApiType:       key + "-entity-relationships",
			JsonApiFormat:     "legacy",
			GetCollectionInfo: nil,
			GetEntityInfo: &CrudEntityInfo{
				Docs:            "https://elasticpath.dev/docs/api/flows/entry-relationships",
				Url:             val.GetEntityInfo.Url + "/relationships/{fields}",
				ContentType:     "",
				QueryParameters: nil,
				MinResources:    0,
				ParentResourceValueOverrides: map[string]string{
					"fields": "slug",
				},
			},
			CreateEntityInfo: &CrudEntityInfo{
				Docs:            "https://elasticpath.dev/docs/api/flows/entry-relationships",
				Url:             val.GetEntityInfo.Url + "/relationships/{fields}",
				ContentType:     "",
				QueryParameters: nil,
				MinResources:    0,
				ParentResourceValueOverrides: map[string]string{
					"fields": "slug",
				},
			},
			UpdateEntityInfo: &CrudEntityInfo{
				Docs:            "https://elasticpath.dev/docs/api/flows/entry-relationships",
				Url:             val.GetEntityInfo.Url + "/relationships/{fields}",
				ContentType:     "",
				QueryParameters: nil,
				MinResources:    0,
				ParentResourceValueOverrides: map[string]string{
					"fields": "slug",
				},
			},
			DeleteEntityInfo: &CrudEntityInfo{
				Docs:            "https://elasticpath.dev/docs/api/flows/delete-an-entry-relationship",
				Url:             val.GetEntityInfo.Url + "/relationships/{fields}",
				ContentType:     "",
				QueryParameters: nil,
				MinResources:    0,
				ParentResourceValueOverrides: map[string]string{
					"fields": "slug",
				},
			},
			Attributes: map[string]*CrudEntityAttribute{
				"data[n].id": {
					Key:  "data[n].id",
					Type: "STRING",
				},
				"data[n].type": {
					Key:  "data[n].type",
					Type: "SINGULAR_RESOURCE_TYPE",
				},
				"data.id": {
					Key:  "data.id",
					Type: "STRING",
				},
				"data.type": {
					Key:  "data.type",
					Type: "SINGULAR_RESOURCE_TYPE",
				}},
			NoWrapping:           true,
			SingularName:         val.SingularName + "-entity-relationship",
			PluralName:           key + "-entity-relationships",
			SuppressResetWarning: true,
		}

		_, ok := resources[newResource.Type]

		if ok {
			log.Warnf("Can not create dynamic resource as one already exists for %s", newResource.Type)
		} else {
			resources[newResource.Type] = newResource
			log.Tracef("Adding new dynamically generated resource %s", newResource.Type)
			resourcesAdded = append(resourcesAdded, newResource.Type)
		}

	}

	log.Debugf("The following resources have been generated dynamically %v", resourcesAdded)

}

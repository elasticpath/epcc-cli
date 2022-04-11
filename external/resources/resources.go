package resources

import (
	_ "embed"
	"gopkg.in/yaml.v3"
)

//go:embed resources.yaml
var resourceMetaData string

var resources map[string]Resource

var resourcesSingular = map[string]Resource{}

func init() {

	err := yaml.Unmarshal([]byte(resourceMetaData), &resources)
	if err != nil {
		panic("Couldn't load the resource meta data")
	}

	for key, val := range resources {
		// Fix the key
		val.Type = key

		val.PluralName = key
		for attributeName, attributeVal := range val.Attributes {
			// Fix the key
			attributeVal.Key = attributeName
		}
		resourcesSingular[val.SingularName] = val
	}

}

type Resource struct {
	// The type as far as the EPCC CLI is concerned.
	Type string

	// A link to the generic documentation page about a type in the EPCC API
	Docs string `yaml:"docs"`

	// The type that should be used in the JSON API.
	JsonApiType string `yaml:"json-api-type"`

	// Some resources (e.g., PCM, accelerator svc, bury most attributes under the attributes key). This is considered "compliant", other services just bury attributes under data, this is "legacy.
	JsonApiFormat string `yaml:"json-api-format"`

	// Information about how to get a collection
	GetCollectionInfo *CrudEntityInfo `yaml:"get-collection"`

	// Information about how to get a single object.
	GetEntityInfo *CrudEntityInfo `yaml:"get-entity"`

	// Information about how to create an entity.
	CreateEntityInfo *CrudEntityInfo `yaml:"create-entity"`

	// Information about how to update an entity.
	UpdateEntityInfo *CrudEntityInfo `yaml:"update-entity"`

	// Information about how to delete an entity.
	DeleteEntityInfo *CrudEntityInfo `yaml:"delete-entity"`

	Attributes map[string]*CrudEntityAttribute `yaml:"attributes"`

	// If true, don't wrap json in a data tag
	NoWrapping bool `yaml:"no-wrapping"`

	// The singular name version of the resource.
	SingularName string `yaml:"singular-name"`

	PluralName string
}

type CrudEntityInfo struct {

	// A link to the docs specific for the Crud operation in EPCC.
	Docs string `yaml:"docs"`

	// The Url we should use when invoking this method.
	Url string `yaml:"url"`

	// Content type to send
	ContentType string `yaml:"content-type"`

	// A list of valid query parameters
	QueryParameters string `yaml:"query"`
}

type CrudEntityAttribute struct {
	// The name of the attribute
	Key string

	// The type of the attribute
	Type string `yaml:"type"`
}

func GetPluralResourceNames() []string {
	keys := make([]string, len(resources))

	for key := range resources {
		keys = append(keys, key)
	}

	return keys
}

func GetPluralResources() map[string]Resource {
	return resources
}

func GetSingularResourceNames() []string {
	keys := make([]string, len(resourcesSingular))

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

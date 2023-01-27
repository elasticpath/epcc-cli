package resources

import (
	_ "embed"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"regexp"
)

//go:embed yaml/resources.yaml
var resourceMetaData string

var resources map[string]Resource

var resourcesSingular = map[string]Resource{}

// This will match a /v2/<FOO>/{BAR} with an optional slash at the end.

var topLevelResourceRegexp = regexp.MustCompile("^/v2/[^/]+/\\{[^}]+}/?$")

type Resource struct {
	// The type as far as the EPCC CLI is concerned.
	Type string

	// A link to the generic documentation page about a type in the EPCC API
	Docs string `yaml:"docs"`

	// The type that should be used in the JSON API.
	JsonApiType string `yaml:"json-api-type"`

	// Alterative types used for aliases
	AlternateJsonApiTypesForAliases []string `yaml:"alternate-json-type-for-aliases"`

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

	// Use this value to silence warnings about a resource not supporting resets.
	// This should only be used for cases where we manually fix things, or where
	// a store reset would clear a resource another way (e.g., the resource represents a projection).
	SuppressResetWarning bool `yaml:"suppress-reset-warning"`
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

	// Minimum resources so we don't keep trying to delete in
	MinResources int `yaml:"min"`

	// Override the attribute we use in the URL for a specific key
	ParentResourceValueOverrides map[string]string `yaml:"parent_resource_value_overrides"`
}

type CrudEntityAttribute struct {
	// The name of the attribute
	Key string

	// The type of the attribute
	Type           string `yaml:"type"`
	AutoFill       string `yaml:"autofill"`
	AliasAttribute string `yaml:"alias_attribute"`
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

func GetSingularResourceByName(name string) (Resource, bool) {
	res, ok := resourcesSingular[name]

	if ok {
		return res, true
	}

	return Resource{}, false
}

func GenerateResourceMetadataFromYaml(yamlTxt string) (map[string]Resource, error) {
	resources := make(map[string]Resource)

	err := yaml.Unmarshal([]byte(yamlTxt), &resources)
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

func init() {

	resourceData, err := GenerateResourceMetadataFromYaml(resourceMetaData)

	if err != nil {
		panic("Couldn't load the resource meta data")
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
			Docs:              "https://documentation.elasticpath.com/commerce-cloud/docs/api/advanced/custom-data/entry-relationships/index.html",
			JsonApiType:       key + "-entity-relationships",
			JsonApiFormat:     "legacy",
			GetCollectionInfo: nil,
			GetEntityInfo: &CrudEntityInfo{
				Docs:            "https://documentation.elasticpath.com/commerce-cloud/docs/api/advanced/custom-data/entry-relationships/index.html",
				Url:             val.GetEntityInfo.Url + "/relationships/{fields}",
				ContentType:     "",
				QueryParameters: "",
				MinResources:    0,
				ParentResourceValueOverrides: map[string]string{
					"fields": "slug",
				},
			},
			CreateEntityInfo: &CrudEntityInfo{
				Docs:            "https://documentation.elasticpath.com/commerce-cloud/docs/api/advanced/custom-data/entry-relationships/create-an-entry-relationship.html",
				Url:             val.GetEntityInfo.Url + "/relationships/{fields}",
				ContentType:     "",
				QueryParameters: "",
				MinResources:    0,
				ParentResourceValueOverrides: map[string]string{
					"fields": "slug",
				},
			},
			UpdateEntityInfo: &CrudEntityInfo{
				Docs:            "https://documentation.elasticpath.com/commerce-cloud/docs/api/advanced/custom-data/entry-relationships/update-entry-relationships.html",
				Url:             val.GetEntityInfo.Url + "/relationships/{fields}",
				ContentType:     "",
				QueryParameters: "",
				MinResources:    0,
				ParentResourceValueOverrides: map[string]string{
					"fields": "slug",
				},
			},
			DeleteEntityInfo: &CrudEntityInfo{
				Docs:            "https://documentation.elasticpath.com/commerce-cloud/docs/api/advanced/custom-data/entry-relationships/delete-entry-relationships.html",
				Url:             val.GetEntityInfo.Url + "/relationships/{fields}",
				ContentType:     "",
				QueryParameters: "",
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

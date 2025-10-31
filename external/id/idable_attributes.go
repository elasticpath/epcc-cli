package id

type IdableAttributes struct {
	Id          string `yaml:"id"`
	Slug        string `yaml:"slug,omitempty"`
	Sku         string `yaml:"sku,omitempty"`
	Code        string `yaml:"code,omitempty"`
	ExternalRef string `yaml:"external_ref,omitempty"`
}

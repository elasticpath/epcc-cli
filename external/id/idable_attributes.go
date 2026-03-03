package id

type IdableAttributes struct {
	Id          string `yaml:"id" json:"id"`
	Slug        string `yaml:"slug,omitempty" json:"slug,omitempty"`
	Sku         string `yaml:"sku,omitempty" json:"sku,omitempty"`
	Code        string `yaml:"code,omitempty" json:"code,omitempty"`
	ExternalRef string `yaml:"external_ref,omitempty" json:"external_ref,omitempty"`
}

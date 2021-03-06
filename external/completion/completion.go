package completion

import (
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/spf13/cobra"
	"log"
	"strconv"
	"strings"
)

const (
	CompletePluralResource    = 1
	CompleteSingularResource  = 2
	CompleteAttributeKey      = 4
	CompleteAttributeValue    = 8
	CompleteQueryParam        = 16
	CompleteCrudAction        = 32
	CompleteAlias             = 64
	CompleteLoginLogoutAPI    = 128
	CompleteLoginClientID     = 256
	CompleteLoginClientSecret = 1024
)

const (
	Get    = 1
	Create = 2
	Update = 4
	Delete = 8
	GetAll = 16
)

type Request struct {
	Type       int
	Resource   resources.Resource
	Attributes map[string]int
	Verb       int
	Attribute  string
}

func Complete(c Request) ([]string, cobra.ShellCompDirective) {
	results := make([]string, 0)
	compDir := cobra.ShellCompDirectiveNoFileComp

	if c.Type&CompletePluralResource > 0 {
		for k := range resources.GetPluralResources() {
			r, _ := resources.GetResourceByName(k) // Not worried about the bool here as resources come from the list already
			if c.Verb&Get > 0 {
				if r.GetCollectionInfo != nil {
					results = append(results, k)
				}
			} else if c.Verb&Delete > 0 {
				if r.DeleteEntityInfo != nil {
					results = append(results, k)
				}
			} else {
				results = append(results, k)
			}
		}
	}

	if c.Type&CompleteSingularResource > 0 {
		for _, v := range resources.GetSingularResourceNames() {
			r, _ := resources.GetResourceByName(v) // Not worried about the bool here as resources come from the list already
			if c.Verb&Create > 0 {
				if r.CreateEntityInfo != nil {
					results = append(results, v)
				}
			} else if c.Verb&Update > 0 {
				if r.UpdateEntityInfo != nil {
					results = append(results, v)
				}
			} else if c.Verb&Delete > 0 {
				if r.DeleteEntityInfo != nil {
					results = append(results, v)
				}
			} else if c.Verb&Get > 0 {
				if r.GetEntityInfo != nil {
					results = append(results, v)
				}
			} else {
				results = append(results, v)
			}
		}
	}

	if c.Type&CompleteCrudAction > 0 {
		results = append(results, "create", "update", "delete", "get")
	}

	if c.Type&CompleteLoginLogoutAPI > 0 {
		results = append(results, "api")
	}

	if c.Type&CompleteLoginClientID > 0 {
		results = append(results, "client_id")
	}

	if c.Type&CompleteLoginClientSecret > 0 {
		results = append(results, "client_secret")
	}

	if c.Type&CompleteAttributeKey > 0 {
		for k := range c.Resource.Attributes {
			if strings.Contains(k, "[n]") {
				i := strings.Index(k, "[n]")
				prefix := k[:i+1]
				max := -1
				for s := range c.Attributes {
					if strings.HasPrefix(s, prefix) {
						n := strings.TrimPrefix(s, prefix)
						i2 := strings.Index(n, "]")
						n = n[:i2]
						m, _ := strconv.Atoi(n)
						if m > max {
							max = m
						}
					}
				}
				for j := 0; j <= max+1; j++ {
					l := strings.Replace(k, "[n]", "["+strconv.Itoa(j)+"]", 1)
					if _, ok := c.Attributes[l]; !ok {
						results = append(results, l)
					}
				}
			} else {
				if _, ok := c.Attributes[k]; !ok {
					results = append(results, k)
				}
			}
		}
	}

	if c.Type&CompleteAttributeValue > 0 {
		if c.Attribute != "" {
			attr := c.Attribute
			i := strings.Index(attr, "[")
			j := strings.Index(attr, "]")
			if i != -1 && j != -1 {
				attr = attr[:i+1] + "n" + attr[j:]
			}
			log.Println("here: " + attr)
			if c.Resource.Attributes[attr].Type == "BOOL" {
				results = append(results, "true", "false")
			} else if strings.HasPrefix(c.Resource.Attributes[attr].Type, "ENUM:") {
				enums := strings.Replace(c.Resource.Attributes[attr].Type, "ENUM:", "", 1)
				for _, k := range strings.Split(enums, ",") {
					results = append(results, k)
				}
			} else if c.Resource.Attributes[attr].Type == "URL" {
				results = append(results, "https://")
				compDir = compDir | cobra.ShellCompDirectiveNoSpace
			} else if strings.HasPrefix(c.Resource.Attributes[attr].Type, "RESOURCE_ID:") {
				resourceType := strings.Replace(c.Resource.Attributes[attr].Type, "RESOURCE_ID:", "", 1)

				if aliasType, ok := resources.GetResourceByName(resourceType); ok {
					for alias := range aliases.GetAliasesForJsonApiType(aliasType.JsonApiType) {
						results = append(results, alias)
					}
				}
			} else if c.Resource.Attributes[c.Attribute].Type == "CURRENCY" {
				currencies := []string{"AED", "AFN", "ALL", "AMD", "ANG", "AOA", "ARS", "AUD", "AWG", "AZN",
					"BAM", "BBD", "BDT", "BGN", "BHD", "BIF", "BMD", "BND", "BOB", "BRL", "BSD", "BTN", "BWP", "BYN", "BZD",
					"CAD", "CDF", "CHF", "CLP", "CNY", "COP", "CRC", "CUC", "CUP", "CVE", "CZK",
					"DJF", "DKK", "DOP", "DZD",
					"EGP", "ERN", "ETB", "EUR",
					"FJD", "FKP",
					"GBP", "GEL", "GGP", "GHS", "GIP", "GMD", "GNF", "GTQ", "GYD",
					"HKD", "HNL", "HRK", "HTG", "HUF",
					"IDR", "ILS", "IMP", "INR", "IQD", "IRR", "ISK",
					"JEP", "JMD", "JOD", "JPY",
					"KES", "KGS", "KHR", "KMF", "KPW", "KRW", "KWD", "KYD", "KZT",
					"LAK", "LBP", "LKR", "LRD", "LSL", "LYD",
					"MAD", "MDL", "MGA", "MKD", "MMK", "MNT", "MOP", "MRU", "MUR", "MVR", "MWK", "MXN", "MYR", "MZN",
					"NAD", "NGN", "NIO", "NOK", "NPR", "NZD",
					"OMR",
					"PAB", "PEN", "PGK", "PHP", "PKR", "PLN", "PYG",
					"QAR",
					"RON", "RSD", "RUB", "RWF",
					"SAR", "SBD", "SCR", "SDG", "SEK", "SGD", "SHP", "SLL", "SOS", "SPL", "SRD", "STN", "SVC", "SYP", "SZL",
					"THB", "TJS", "TMT", "TND", "TOP", "TRY", "TTD", "TVD", "TWD", "TZS",
					"UAH", "UGX", "USD", "UYU", "UZS",
					"VEF", "VND", "VUV",
					"WST",
					"XAF", "XCD", "XDR", "XOF", "XPF",
					"YER",
					"ZAR", "ZMW", "ZWD"}
				for _, currency := range currencies {
					results = append(results, currency)
				}
			}
		}
	}

	if c.Type&CompleteQueryParam > 0 {
		if c.Verb&GetAll > 0 {
			for _, k := range strings.Split(c.Resource.GetCollectionInfo.QueryParameters, ",") {
				results = append(results, k)
			}
		} else if c.Verb&Get > 0 {
			for _, k := range strings.Split(c.Resource.GetEntityInfo.QueryParameters, ",") {
				results = append(results, k)
			}
		}

	}

	if c.Type&CompleteAlias > 0 {
		jsonApiType := c.Resource.JsonApiType
		aliasesForJsonApiType := aliases.GetAliasesForJsonApiType(jsonApiType)

		for alias := range aliasesForJsonApiType {
			results = append(results, alias)
		}
	}

	// This is dead code since I hacked the aliases to never return spaces.
	newResults := make([]string, 0, len(results))

	for _, result := range results {
		newResults = append(newResults, strings.ReplaceAll(result, " ", "\\ "))
	}

	return newResults, compDir
}

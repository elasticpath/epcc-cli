package completion

import (
	gojson "encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	CompletePluralResource            = 1
	CompleteSingularResource          = 2
	CompleteAttributeKey              = 4
	CompleteAttributeValue            = 8
	CompleteQueryParamKey             = 16
	CompleteQueryParamValue           = 32
	CompleteCrudAction                = 64
	CompleteAlias                     = 128
	CompleteLoginLogoutAPI            = 256
	CompleteLoginClientID             = 512
	CompleteLoginClientSecret         = 1024
	CompleteLoginAccountManagementKey = 2048

	CompleteHeaderKey   = 4096
	CompleteHeaderValue = 8192

	CompleteCurrency = 16384
	CompleteBool     = 32768
)

const (
	Get       = 1
	Create    = 2
	Update    = 4
	Delete    = 8
	GetAll    = 16
	DeleteAll = 32
)

type Request struct {
	Type     int
	Resource resources.Resource

	// These are consumed attributes
	Attributes map[string]string

	// This only makes sense on get, and is largely the current state of the resource.
	ExistingResourceAttributes map[string]string

	SkipWhenChecksAndAddAll bool
	Verb                    int
	Attribute               string
	QueryParam              string
	Header                  string
	// The current string argument being completed
	ToComplete     string
	NoAliases      bool
	AllowTemplates bool
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
			} else if c.Verb&DeleteAll > 0 {
				if r.DeleteEntityInfo != nil && r.GetCollectionInfo != nil {
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

	if c.Type&CompleteBool > 0 {
		results = append(results, "true", "false")
	}

	if c.Type&CompleteLoginClientID > 0 {
		results = append(results, "client_id")
	}

	if c.Type&CompleteLoginClientSecret > 0 {
		results = append(results, "client_secret")
	}

	if c.Type&CompleteAttributeKey > 0 {
		autoCompleteAttributes := []string{}

		rt := NewRegexCompletionTree()

		usedAttributes := []string{}
		for s := range c.Attributes {
			usedAttributes = append(usedAttributes, s)
		}

		for k, v := range c.Resource.Attributes {
			if (strings.HasPrefix(k, "^")) && (strings.HasSuffix(k, "$")) {
				// We don't support when: or [n] indexes on these cases.
				rt.AddRegex(k)
			} else {
				attributesToCheck := []resources.CrudEntityAttribute{}

				resolvedAttributes := resolveArrayParametersForIndexes(k, usedAttributes)

				for _, r := range resolvedAttributes {
					newAttribute := *v

					newAttribute.Key = r

					attributesToCheck = append(attributesToCheck, newAttribute)
				}

				for _, attr := range attributesToCheck {
					if (attr.When == "" && attr.Type != "CONDITIONAL") || c.SkipWhenChecksAndAddAll {
						autoCompleteAttributes = append(autoCompleteAttributes, attr.Key)
					} else {
						// Conditional attributes may occur when the main attribute itself has a when or not.
						orClauses := []string{}

						for _, cond := range attr.Conditions {
							orClauses = append(orClauses, cond.When)
						}

						whenCondition := ""
						if len(strings.Trim(attr.When, " ")) == 0 {
							whenCondition = "(" + strings.Join(orClauses, " || ") + ")"
						} else if len(orClauses) == 0 {
							whenCondition = attr.When
						} else {
							whenCondition = "( " + attr.When + " ) && (" + strings.Join(orClauses, " || ") + ")"
						}
						patcher := arrayReferencePatcher{
							attr: attr.Key,
						}

						prog, err := expr.Compile(whenCondition, expr.AsBool(), expr.AllowUndefinedVariables(), expr.Patch(&patcher))

						//newWhenCondition := prog.Node().String()
						//log.Infof("New Program: %s", newWhenCondition)
						// We will need to visit the resulting compiled expression and rewrite anything with "[n]"
						if err != nil {
							log.Tracef("Invalid when condition on resource `%s` and attribute `%s`: %s", c.Resource.PluralName, attr.Key, whenCondition)
							// We will add it anyway, as it's a better experience I suppose.
							autoCompleteAttributes = append(autoCompleteAttributes, attr.Key)
							continue
						}

						allAttributes := make(map[string]string, len(c.Attributes)+len(c.ExistingResourceAttributes))

						flatAttributes := make([]string, 0, len(c.Attributes)+len(c.ExistingResourceAttributes))
						for k, v := range c.ExistingResourceAttributes {
							allAttributes[k] = v

							flatAttributes = append(flatAttributes, k, v)
						}

						for k, v := range c.Attributes {
							allAttributes[k] = v
							flatAttributes = append(flatAttributes, k, v)
						}

						// TODO clean up all these different json formats
						jsonTxt, err := json.ToJson(flatAttributes, true, false, make(map[string]*resources.CrudEntityAttribute), false, false)

						if err == nil {
							var jsonBlob any
							gojson.Unmarshal([]byte(jsonTxt), &jsonBlob)

							var output any

							output, err = expr.Run(prog, jsonBlob)

							if err != nil {
								log.Warnf("Program did not successfully output: %v", err)
							} else if output.(bool) {
								resolvedAttributes := resolveArrayParametersForIndexes(k, usedAttributes)

								autoCompleteAttributes = append(autoCompleteAttributes, resolvedAttributes...)
							}
						} else {
							log.Tracef("Error while evaluating `%s` and attribute `%s`: %s, with values %v", c.Resource.PluralName, k, whenCondition, c.Attributes)
						}

						if err != nil {
							continue
						}
					}
				}
			}
		}

		for s := range c.Attributes {
			rt.AddExistingValue(s)
		}

		if regexOptions, err := rt.GetCompletionOptions(); err == nil {
			for _, v := range regexOptions {
				resolvedAttributes := resolveArrayParametersForIndexes(v, usedAttributes)
				autoCompleteAttributes = append(autoCompleteAttributes, resolvedAttributes...)
			}

		}

		for _, k := range autoCompleteAttributes {
			if _, ok := c.Attributes[k]; !ok {
				results = append(results, k)
			}
		}
	}

	if c.Type&CompleteAttributeValue > 0 && c.Attribute != "" {
		attr := c.Attribute
		i := strings.Index(attr, "[")
		j := strings.Index(attr, "]")
		if i != -1 && j != -1 {
			attr = attr[:i+1] + "n" + attr[j:]
		}
		attribute := c.Resource.Attributes[attr]

		if attribute == nil {
			for k, v := range c.Resource.Attributes {
				if k[0] == '^' && k[len(k)-1] == '$' {
					// Unit tests should stop a panic
					r := regexp.MustCompile(k)

					if r.MatchString(attr) {
						attribute = v
						break
					}
				}
			}
		}

		if attribute != nil {
			results, compDir = resolveAttributeCompletions(c, attribute, results, compDir)
		}

		if c.AllowTemplates {
			lastPipe := strings.LastIndex(c.ToComplete, "|")
			prefix := ""
			if lastPipe == -1 {
				prefix = "{{ "
			} else {
				prefix = c.ToComplete[0:lastPipe+1] + " "
			}

			myResults := []string{}
			myResults = append(myResults,
				prefix+"date",
				prefix+"now",
				prefix+"randAlphaNum",
				prefix+"randAlpha",
				prefix+"randAscii",
				prefix+"randNumeric",
				prefix+"randAlphaNum",
				prefix+"randAlpha",
				prefix+"randAscii",
				prefix+"randNumeric",
				prefix+"pseudoRandAlphaNum",
				prefix+"pseudoRandAlpha",
				prefix+"pseudoRandNumeric",
				prefix+"pseudoRandString",
				prefix+"pseudoRandInt",
				prefix+"uuidv4",
				prefix+"duration",
			)

			if prefix != "{{ " {
				// Functions that make sense as continuations
				myResults = append(myResults,
					prefix+"trim",
					prefix+"trimAll",
					prefix+"trimSuffix",
					prefix+"trimPrefix",
					prefix+"upper",
					prefix+"lower",
					prefix+"title",
					prefix+"repeat",
					prefix+"substr",
					prefix+"nospace",
					prefix+"trunc",
					prefix+"abbrev",
					prefix+"initials",
					prefix+"wrap",
					prefix+"cat",
					prefix+"replace",
					prefix+"snakecase",
					prefix+"camelcase",
					prefix+"kebabcase",
					prefix+"swapcase",
					prefix+"shufflecase",
				)
			}

			re := regexp.MustCompile(`env\s+[A-Za-z]*\s*$`)
			if re.MatchString(c.ToComplete) {
				for _, v := range os.Environ() {
					myResults = append(myResults,
						fmt.Sprintf("%venv \"%v\"", prefix, strings.Split(v, "=")[0]),
					)
				}
			} else {
				myResults = append(myResults, prefix+"env")
			}

			for _, r := range myResults {
				results = append(results, r+" |", r+" }}")
			}
		}

	}

	if c.Type&CompleteQueryParamKey > 0 {
		if c.Verb&GetAll > 0 {
			if c.Resource.GetCollectionInfo != nil {
				if len(c.Resource.GetCollectionInfo.QueryParameters) == 0 {
					// Static shared list
					results = append(results, "sort", "filter", "include", "page[limit]", "page[offset]", "page[total_method]")
				} else {
					for _, param := range c.Resource.GetCollectionInfo.QueryParameters {
						results = append(results, param.Name)
					}
				}

			}

		} else if c.Verb&Get > 0 {
			if c.Resource.GetEntityInfo != nil {
				if len(c.Resource.GetEntityInfo.QueryParameters) == 0 {
					results = append(results, "include")
				} else {
					for _, param := range c.Resource.GetEntityInfo.QueryParameters {
						results = append(results, param.Name)
					}
				}
			}
		}
	}

	if c.Type&CompleteAlias > 0 {
		jsonApiType := c.Resource.JsonApiType
		if !c.NoAliases {
			aliasesForJsonApiType := aliases.GetAliasesForJsonApiTypeAndAlternates(jsonApiType, c.Resource.AlternateJsonApiTypesForAliases)

			for alias := range aliasesForJsonApiType {
				results = append(results, alias)
			}
		}
	}

	if c.Type&CompleteQueryParamValue > 0 {
		if c.Verb&GetAll > 0 {
			if c.QueryParam == "sort" {
				for key := range c.Resource.Attributes {
					results = append(results, key, "-"+key)
				}

				results = append(results, "updated_at", "-updated_at", "created_at", "-created_at")
			} else if c.QueryParam == "filter" {
				results = append(results, GetFilterCompletion(c.ToComplete, c.Resource)...)
				compDir = compDir | cobra.ShellCompDirectiveNoSpace
			} else if c.QueryParam == "page[total_method]" {
				results = append(results, "exact", "estimate", "lower_bound", "observed", "cached", "none")
			}
		}
	}

	if c.Type&CompleteLoginAccountManagementKey > 0 {
		results = append(results, "account_id", "account_name")
	}

	if c.Type&CompleteHeaderKey > 0 {

		headersMutex.RLock()
		defer headersMutex.RUnlock()

		for k := range supportedHeadersToCompletionRequest {
			results = append(results, supportedHeadersOriginalCasing[k])
		}
	}

	if c.Type&CompleteHeaderValue > 0 {
		headersMutex.RLock()
		defer headersMutex.RUnlock()

		v := supportedHeadersToCompletionRequest[strings.ToLower(c.Header)]

		if v != nil {
			r, _ := Complete(*v)

			results = append(results, r...)
		}
	}

	if c.Type&CompleteCurrency > 0 {
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

		results = append(results, currencies...)
	}

	// This is dead code since I hacked the aliases to never return spaces.
	newResults := make([]string, 0, len(results))

	for _, result := range results {
		newResults = append(newResults, strings.ReplaceAll(result, " ", "\\ "))
	}

	return newResults, compDir
}

func resolveAttributeCompletions(c Request, attribute *resources.CrudEntityAttribute, results []string, compDir cobra.ShellCompDirective) ([]string, cobra.ShellCompDirective) {
	if attribute.Type == "BOOL" {
		results = append(results, "true", "false")
	} else if strings.HasPrefix(attribute.Type, "ENUM:") {
		enums := strings.Replace(attribute.Type, "ENUM:", "", 1)
		for _, k := range strings.Split(enums, ",") {
			results = append(results, k)
		}
	} else if attribute.Type == "URL" {
		results = append(results, "https://")
		compDir = compDir | cobra.ShellCompDirectiveNoSpace
	} else if strings.HasPrefix(attribute.Type, "RESOURCE_ID:") {
		resourceType := strings.Replace(attribute.Type, "RESOURCE_ID:", "", 1)

		if resourceType == "*" {
			toComplete := c.ToComplete
			fullyQualifiedAlias := strings.Split(toComplete, "/")

			switch len(fullyQualifiedAlias) {
			case 1:
				results = append(results, "alias/")
				compDir = compDir | cobra.ShellCompDirectiveNoSpace

			case 2:
				for _, v := range resources.GetPluralResources() {
					results = append(results, "alias/"+v.JsonApiType+"/")
				}
				compDir = compDir | cobra.ShellCompDirectiveNoSpace

			case 3, 4:
				if aliasType, ok := resources.GetResourceByName(fullyQualifiedAlias[1]); ok {
					if !c.NoAliases {
						for alias := range aliases.GetAliasesForJsonApiTypeAndAlternates(aliasType.JsonApiType, aliasType.AlternateJsonApiTypesForAliases) {
							results = append(results, "alias/"+aliasType.JsonApiType+"/"+alias+"/id")

							if _, ok2 := aliasType.Attributes["sku"]; ok2 {
								results = append(results, "alias/"+aliasType.JsonApiType+"/"+alias+"/sku")
							}

							if _, ok2 := aliasType.Attributes["slug"]; ok2 {
								results = append(results, "alias/"+aliasType.JsonApiType+"/"+alias+"/slug")
							}

							if _, ok2 := aliasType.Attributes["code"]; ok2 {
								results = append(results, "alias/"+aliasType.JsonApiType+"/"+alias+"/code")
							}

						}
					}
				}

			}

		} else if aliasType, ok := resources.GetResourceByName(resourceType); ok {

			if !c.NoAliases {
				for alias := range aliases.GetAliasesForJsonApiTypeAndAlternates(aliasType.JsonApiType, aliasType.AlternateJsonApiTypesForAliases) {
					results = append(results, alias)
				}
			}
		}
	} else if attribute.Type == "SINGULAR_RESOURCE_TYPE" {
		results = append(results, resources.GetSingularResourceNames()...)

	} else if attribute.Type == "JSON_API_TYPE" {
		for _, v := range resources.GetPluralResources() {
			results = append(results, v.JsonApiType)
		}

	} else if attribute.Type == "CURRENCY" {
		res, _ := Complete(Request{
			Type: CompleteCurrency,
		})

		results = append(results, res...)

	} else if attribute.Type == "FILE" {
		compDir = cobra.ShellCompDirectiveFilterFileExt

		// https://documentation.elasticpath/epcc-cli/docs/api/advanced/files/create-a-file.html#post-create-a-file
		supportedFileTypes := []string{
			"gif",
			"jpg", "jpeg",
			"png",
			"webp",
			"mp4",
			"mov",
			"pdf",
			"svg",
			"usdz",
			"glb",
			"jp2",
			"jxr",
			"aac",
			"vrml",
			"doc", "docx",
			"ppt", "pptx",
			"xls", "xlsx",
		}
		results = append(results, supportedFileTypes...)
	}
	return results, compDir
}

func resolveArrayParametersForIndexes(key string, attributes []string) []string {

	if !strings.Contains(key, "[n]") {
		return []string{key}
	}

	// Count [n] occurrences
	nCount := strings.Count(key, "[n]")

	// Track maximum index at each [n] position
	maxAtPosition := make([]int, nCount)
	for i := range maxAtPosition {
		maxAtPosition[i] = -1
	}

	// Convert pattern to regex to match existing attributes
	regexPattern := strings.ReplaceAll(regexp.QuoteMeta(key), `\[n\]`, `\[(\d+)\]`)
	re := regexp.MustCompile("^" + regexPattern + "$")

	// Find maximum index at each position
	for _, s := range attributes {
		matches := re.FindStringSubmatch(s)
		if matches != nil {
			for i := 1; i < len(matches) && i-1 < len(maxAtPosition); i++ {
				if idx, err := strconv.Atoi(matches[i]); err == nil {
					if idx > maxAtPosition[i-1] {
						maxAtPosition[i-1] = idx
					}
				}
			}
		}
	}

	attributes = []string{}

	// Generate new attributes by incrementing each position independently
	for pos := 0; pos < nCount; pos++ {
		newAttr := key
		for i := 0; i < nCount; i++ {
			var replaceWith string
			if i == pos {
				// Increment this position
				replaceWith = "[" + strconv.Itoa(maxAtPosition[i]+1) + "]"
			} else {
				// Use current max for other positions (or 0 if no max found)
				maxVal := maxAtPosition[i]
				if maxVal < 0 {
					maxVal = 0
				}
				replaceWith = "[" + strconv.Itoa(maxVal) + "]"
			}
			newAttr = strings.Replace(newAttr, "[n]", replaceWith, 1)
		}

		attributes = append(attributes, newAttr)

	}

	return attributes
}

type arrayReferencePatcher struct {
	attr       string
	lastMember string
}

func (a *arrayReferencePatcher) Visit(node *ast.Node) {
	if n, ok := (*node).(*ast.MemberNode); ok {
		a.lastMember = n.String()
		log.Tracef("found member, %v: %v", n, a.lastMember)
		v := *n
		v.Optional = true
		ast.Patch(node, &v)
	}

	if n, ok := (*node).(*ast.ChainNode); ok {
		log.Tracef("found operator, %v: %v", n, a.lastMember)
	}

	if n, ok := (*node).(*ast.IdentifierNode); ok {
		if n.Value == "n" {
			log.Tracef("found identifier %v[%v] while checking %v ", a.lastMember, n.Value, a.attr)

			regexPattern := strings.ReplaceAll(regexp.QuoteMeta(a.lastMember+"[n]"), `\[n\]`, `\[(\d+)\]`)
			re, err := regexp.Compile("^" + regexPattern)
			if err != nil {
				log.Warnf("Could not compile regex %v", err)
				return
			}

			matches := re.FindStringSubmatch(a.attr)

			if len(matches) == 2 {
				m, err := strconv.Atoi(matches[1])

				if err != nil {
					// regex should only match integers.
					panic(err)
				}

				log.Tracef("Found match %v", m)
				ast.Patch(node, &ast.IntegerNode{Value: m})
			}

		}

	}
}

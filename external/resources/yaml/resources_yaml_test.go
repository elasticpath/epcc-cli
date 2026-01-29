package resources__test

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
	"github.com/quasilyte/regex/syntax"
	"github.com/santhosh-tekuri/jsonschema/v4"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yosida95/uritemplate/v3"
	"gopkg.in/yaml.v3"
)

// This test largely exists to ensure we don't lose resources (this was important during refactoring)
func TestExpectedNumberOfResources(t *testing.T) {
	// Fixture Setup

	// Execute SUT
	resourceCount := len(resources.GetPluralResources())

	// Verification
	require.Equal(t, resourceCount, 159)
}

func TestCreatedByTemplatesAllReferenceValidResource(t *testing.T) {
	// Fixture Setup
	errors := ""

	// Execute SUT
	for key, val := range resources.GetPluralResources() {

		for _, created := range val.CreatedBy {

			targetResource, ok := resources.GetResourceByName(created.Resource)

			if !ok {
				errors += fmt.Sprintf("Resource %s references not-found resource %s in created by\n", key, created.Resource)
				continue
			}

			var unsupported bool
			switch created.Verb {
			case "get":
				unsupported = targetResource.GetEntityInfo == nil && targetResource.GetCollectionInfo == nil
			case "delete":
				unsupported = targetResource.DeleteEntityInfo == nil
			case "create":
				unsupported = targetResource.CreateEntityInfo == nil
			case "update":
				unsupported = targetResource.UpdateEntityInfo == nil
			default:
				errors += fmt.Sprintf("Resource %s references unknown verb %s for %s\n", key, created.Verb, created.Resource)
			}
			if unsupported {
				errors += fmt.Sprintf("Resource %s references resource %s with unsupported verb %s in created_by\n", key, created.Resource, created.Verb)
			}
		}

	}

	// Verification
	if len(errors) > 0 {
		t.Errorf("Unexpected errors:\n%s", errors)
	}
}

func TestCreatesAllReferenceValidResource(t *testing.T) {

	// Fixture Setup
	errors := ""

	// Execute SUT
	for key, val := range resources.GetPluralResources() {
		if val.CreateEntityInfo != nil {
			if val.CreateEntityInfo.Creates != "" {
				_, ok := resources.GetResourceByName(val.CreateEntityInfo.Creates)

				if !ok {
					errors += fmt.Sprintf("Resource %s references not-found resource %s in create-entity.creates\n", key, val.CreateEntityInfo.Creates)
					continue
				}
			}

		}
	}

	// Verification
	if len(errors) > 0 {
		t.Errorf("Unexpected errors:\n%s", errors)
	}
}

func TestUriTemplatesAllReferenceValidResource(t *testing.T) {
	// Fixture Setup

	// nothing needed.

	// Execute SUT
	errors := ""
	for key, val := range resources.GetPluralResources() {

		if val.CreateEntityInfo != nil {
			err := validateCrudEntityInfo(*val.CreateEntityInfo)
			if err != "" {
				errors += fmt.Sprintf("Could not process CREATE uri for resource `%s`, error:\n%s\n", key, err)
			}
		}

		if val.UpdateEntityInfo != nil {
			err := validateCrudEntityInfo(*val.UpdateEntityInfo)
			if err != "" {
				errors += fmt.Sprintf("Could not process UPDATE uri for resource `%s`, error:\n%s\n", key, err)
			}
		}

		if val.DeleteEntityInfo != nil {

			err := validateCrudEntityInfo(*val.DeleteEntityInfo)
			if err != "" {
				errors += fmt.Sprintf("Could not process DELETE uri for resource `%s`, error:\n%s\n", key, err)
			}
		}

		if val.GetEntityInfo != nil {
			err := validateCrudEntityInfo(*val.GetEntityInfo)
			if err != "" {
				errors += fmt.Sprintf("Could not process GET entity uri for resource `%s`, error:\n%s\n", key, err)
			}
		}

		if val.GetCollectionInfo != nil {
			err := validateCrudEntityInfo(*val.GetCollectionInfo)
			if err != "" {
				errors += fmt.Sprintf("Could not process GET collection uri for resource `%s`, error:\n%s\n", key, err)
			}
		}

		for attributeKey, attributeInfo := range val.Attributes {
			err := validateAttributeInfo(attributeInfo)
			if err != "" {
				errors += fmt.Sprintf("Couldn't process attributes for resource `%s` attribute `%s`, error:\n%s\n", key, attributeKey, err)
			}
		}
	}

	// Verification
	if len(errors) > 0 {
		t.Fatalf("Errors occurred while validating URI Templates:\n%s", errors)
	}
}

type visitor struct {
	Identifiers []string
}

func (v *visitor) Visit(node *ast.Node) {
	if n, ok := (*node).(*ast.IdentifierNode); ok {
		v.Identifiers = append(v.Identifiers, n.Value)
	}
}

func TestAllWhenConditionsAreValid(t *testing.T) {
	// Fixture Setup

	errors := ""
	// Execute SUT

	for key, val := range resources.GetPluralResources() {
		//allAttributes := make([]string, 0, len(val.Attributes))
		//
		//for k := range val.Attributes {
		//	allAttributes = append(allAttributes, k)
		//}

		for attr, attrObj := range val.Attributes {
			if attrObj.When != "" {
				if strings.Trim(attrObj.When, " ") != attrObj.When {
					errors += fmt.Sprintf("\t attribute `%s` of resource `%s` has a when condition that has leading or trailing whitespace: `%s`\n", attr, key, attrObj.When)
					continue
				}

				_, err := expr.Compile(attrObj.When, expr.AsBool())

				if err != nil {
					errors += fmt.Sprintf("\t attribute `%s` of resource `%s` has a when condition that doesn't compile: `%s`, error: `%v`\n", attr, key, attrObj.When, err)
					continue
				}

				tree, err := parser.Parse(attrObj.When)
				if err != nil {
					errors += fmt.Sprintf("\t attribute `%s` of resource `%s` has a when condition that doesn't compile: `%s`, error: `%v`\n", attr, key, attrObj.When, err)
				}

				v := &visitor{}
				ast.Walk(&tree.Node, v)

				for _, id := range v.Identifiers {
					if _, ok := val.Attributes[id]; !ok {
						errors += fmt.Sprintf("\t attribute %s of resource `%s` has a when condition `%s` with an unspecified value: `%s`\n", attr, key, attrObj.When, id)
						break
					}
				}
			}
		}
	}
	// Verification

	if len(errors) > 0 {
		t.Fatalf("Errors occurred while validating when conditions:\n%s", errors)
	}
}

var arrayLiteralIndex = regexp.MustCompile("\\[[0-9]+]")

func validateAttributeInfo(info *resources.CrudEntityAttribute) string {
	match := arrayLiteralIndex.Match([]byte(info.Key))
	errors := ""

	if info.Key[0] == '^' {
		if info.Key[len(info.Key)-1] != '$' {
			errors += fmt.Sprintf("\t attribute `%s` starts with a ^ but doesn't end with a $, this is likely a bug due to regex rules)\n", info.Key)
		} else {
			if _, err := regexp.Compile(info.Key); err != nil {
				errors += fmt.Sprintf("\t attribute `%s` is a regex, but it doesn't compile: %v", info.Key, err)
			}

			rt := completion.NewRegexCompletionTree()
			if err := rt.AddRegex(info.Key); err != nil {
				errors += fmt.Sprintf("\t attribute `%s` is a regex, but the completion tree doesn't support it: %v", info.Key, err)
			}

			p := syntax.NewParser(&syntax.ParserOptions{})
			parse, err := p.Parse(info.Key)

			err = validateRegexTreeContainsNoSingleCharClass(parse)

			if err != nil {
				errors += fmt.Sprintf("\t attribute `%s` is a regex, but it contains a single char class: %v", info.Key, err)
			}

		}
	}
	if match {
		errors += fmt.Sprintf("\t attribute `%s` has array index (e.g., [4] instead of [n], this is almost certainly a bug)\n", info.Key)
	}

	if strings.HasPrefix(info.Type, "RESOURCE_ID:") {
		resourceType := info.Type[len("RESOURCE_ID:"):]
		if resourceType != "*" {
			if _, ok := resources.GetResourceByName(resourceType); !ok {

				if _, ok := resources.GetSingularResourceByName(resourceType); !ok {
					errors += fmt.Sprintf("\t attribute `%s` references a resource type that doesn't exist: %s\n", info.Key, resourceType)
				}
			}
		}
	}

	return errors
}

func validateRegexTreeContainsNoSingleCharClass(tree *syntax.Regexp) error {

	var checkChildren func(e *syntax.Expr) error
	checkChildren = func(e *syntax.Expr) error {

		if e.Op == syntax.OpCharClass {
			if len(e.Args) == 1 {
				return fmt.Errorf("single char class found: %v", e.Value)
			}
		}

		for _, child := range e.Args {

			err := checkChildren(&child)

			if err != nil {
				return err
			}

		}

		return nil
	}

	return checkChildren(&tree.Expr)
}

func validateCrudEntityInfo(info resources.CrudEntityInfo) string {
	errors := ""

	template, err := uritemplate.New(info.Url)
	if err != nil {
		errors += fmt.Sprintf("\tCould not process Uri %s for templates error:%s\n", info.Url, err)
	} else {
		variables := map[string]bool{}
		for _, variable := range template.Varnames() {
			variables[variable] = true
			resourceName := strings.ReplaceAll(variable, "_", "-")
			if _, ok := resources.GetPluralResources()[resourceName]; !ok {
				errors += fmt.Sprintf("\tError processing Uri %s, the URI template references a resource %s, but could not find it\n", info.Url, resourceName)
			}
		}

		for key, value := range info.ParentResourceValueOverrides {
			if value != "slug" && value != "sku" && value != "id" {
				errors += fmt.Sprintf("\tUrl %s has an invalid override for %s => %s\n", info.Url, key, value)
			}

			if _, ok := variables[key]; !ok {
				errors += fmt.Sprintf("\tUrl %s has an invalid override for %s, this key doesn't exist in the URL", info.Url, key)
			}
		}

	}

	return errors
}

func TestJsonSchemaValidate(t *testing.T) {
	sch, err := jsonschema.Compile("../resources_schema.json")
	if err != nil {
		log.Fatalf("%#v", err)
	}

	dirEntries, err := os.ReadDir(".")

	if err != nil {
		log.Fatalf("%#v", err)
	}

	for _, file := range dirEntries {
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		data, err := os.ReadFile(file.Name())
		if err != nil {
			log.Fatal(err)
		}

		var v interface{}
		if err := yaml.Unmarshal(data, &v); err != nil {
			log.Fatalf("error processing file %s: %#v", file.Name(), err)
		}

		if err = sch.ValidateInterface(v); err != nil {
			if e2, ok := err.(*jsonschema.ValidationError); ok {
				t.Errorf("Error processing file %s:\n%s", file.Name(), e2.GoString())
			} else {
				t.Errorf("Error processing file %s:\n%#v", file.Name(), err)
			}

		}
	}

}

var redirectRegex = regexp.MustCompile(`window\.location\.href\s*=\s*'([^']+)'`)
var titleRegex = regexp.MustCompile(`<title[^>]*>([^<]*)</title`)

func TestResourceDocsExist(t *testing.T) {
	const httpStatusCodeOk = 200

	Resources := resources.GetPluralResources()
	linksReferenceCount := make(map[string]int, len(Resources))

	for resource := range Resources {
		linksReferenceCount[Resources[resource].Docs]++
		if Resources[resource].GetCollectionInfo != nil {
			linksReferenceCount[Resources[resource].GetCollectionInfo.Docs]++
		}
		if Resources[resource].CreateEntityInfo != nil {
			linksReferenceCount[Resources[resource].CreateEntityInfo.Docs]++
		}
		if Resources[resource].GetEntityInfo != nil {
			linksReferenceCount[Resources[resource].GetEntityInfo.Docs]++
		}
		if Resources[resource].UpdateEntityInfo != nil {
			linksReferenceCount[Resources[resource].UpdateEntityInfo.Docs]++
		}
		if Resources[resource].DeleteEntityInfo != nil {
			linksReferenceCount[Resources[resource].DeleteEntityInfo.Docs]++
		}
	}

	mutex := &sync.Mutex{}

	var rewriteUrlOne string = ""
	client := http.Client{
		Transport: nil,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {

			mutex.Lock()
			rewriteUrlOne = via[0].URL.String()
			defer mutex.Unlock()
			return nil
		},
		Jar:     nil,
		Timeout: 0,
	}

	links := []string{}

	for link := range linksReferenceCount {
		links = append(links, link)
	}

	sort.Strings(links)

	os.Mkdir("build", 0755)

	pageNotFound := 0
	oldDomain := 0
	brokenRedirectToRoot := 0

	maxLength := 0
	for _, l := range links {
		if len(l) > maxLength {
			maxLength = len(l)
		}
	}

	for _, link := range links {

		if link == "n/a" {
			continue
		}

		rewriteUrlOne = ""
		response, err := client.Get(link)

		if err != nil {
			t.Errorf("Error Retrieving Link\nLink: %s\nError Message: %s\nReference Count: %d", link, err, linksReferenceCount[link])
		} else {
			resp, err := io.ReadAll(response.Body)
			if err != nil {
				t.Errorf("Unexpected error reading response body")
			}

			if response.StatusCode != httpStatusCodeOk {

				//fmt.Printf("%s => %d\n", link, response.StatusCode)
				//t.Errorf("Unexpected Response\nLink: %s\nExpected Status Code: %d\nActual Status Code: %d\nReference Count: %d",
				//	link, httpStatusCodeOk, response.StatusCode, linksReferenceCount[link])
			}

			respString := string(resp)

			prefix := "# %-" + strconv.Itoa(maxLength) + "s"
			if strings.Index(respString, "Your Docusaurus site did not load properly") > 0 {
				fmt.Printf(prefix+"=> ERROR (Page Not Found (Maybe))\n", link)
				pageNotFound++
				continue
			}

			matches := redirectRegex.FindStringSubmatch(respString)

			if len(matches) >= 2 {
				if matches[1] != "/" && matches[1] != "/guides/Getting-Started/includes" {
					//fmt.Printf("\t Further Redirect to %s =>  %s \n", rewriteUrlOne, matches[1])
					//fmt.Printf("Rewrite %s => https://elasticpath.dev%s\n", link, matches[1])
					fmt.Printf("# %s => https://elasticpath.dev%s\n", link, matches[1])
					fmt.Printf("sed -E -i 's@%s@https://elasticpath.dev%s@g' *.yaml\n", link, matches[1])
				} else if matches[1] == "/" {
					fmt.Printf("\t Broken Redirect to =>  %s \n", matches[1])
					brokenRedirectToRoot++
				} else {
					fmt.Printf("\t Broken Redirect to unknown =>  %s \n", matches[1])
					brokenRedirectToRoot++
				}
			} else if rewriteUrlOne != "" {

				mutex.Lock()
				// Rewrite
				if link != rewriteUrlOne {
					fmt.Printf("# %s => %s\n", link, rewriteUrlOne)
					fmt.Printf("sed  -E -i 's@%s@%s@g *.yaml'\n", link, rewriteUrlOne)
				}

				mutex.Unlock()

			}

			if strings.Index(link, "documentation.elasticpath.com") > 0 {
				//fmt.Printf(" %s => FAIL (old link)", link)
				if rewriteUrlOne != "" {
					fmt.Printf("  Should Be => %s\n", rewriteUrlOne)
				} else {
					fmt.Printf("\n")
				}
				oldDomain++
			}

			matches = titleRegex.FindStringSubmatch(respString)

			if len(matches) >= 2 {
				fmt.Printf(prefix+"=> OK (%s)\n", link, matches[1])
			} else if strings.Index(respString, "openapi__method-endpoint") > 0 {
				fmt.Printf(prefix+" => OK\n", link)
				continue
			} else {
				fmt.Printf(prefix+" => ???\n", link)
			}

			if err := response.Body.Close(); err != nil {
			}

			fname, _ := sanitizeFilename(link)
			err = os.WriteFile("build/"+fname, resp, 0644)

			if err != nil {
				t.Errorf("Error writing file\nError Message: %s", err)
			}
		}
	}

	assert.Zerof(t, pageNotFound, "Page Not Found Count: %d", pageNotFound)
	assert.Zerof(t, oldDomain, "Old Domain Count: %d", oldDomain)
	assert.Zerof(t, brokenRedirectToRoot, "Broken Redirects: %d", brokenRedirectToRoot)

}

// sanitizeFilename converts a URL into a safe filename by replacing unsafe characters with dashes
func sanitizeFilename(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Construct the base filename from host and path
	base := parsedURL.Host + parsedURL.Path

	// Replace all non-alphanumeric, non-dash, and non-dot characters with a dash
	re := regexp.MustCompile(`[^a-zA-Z0-9.-]+`)
	safeFilename := re.ReplaceAllString(base, "-")

	// Ensure it has an .html suffix
	return safeFilename + ".html", nil
}

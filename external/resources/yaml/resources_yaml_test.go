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
	"github.com/santhosh-tekuri/jsonschema/v4"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/yosida95/uritemplate/v3"
	"gopkg.in/yaml.v3"
)

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

	data, err := os.ReadFile("resources.yaml")
	if err != nil {
		log.Fatal(err)
	}

	var v interface{}
	if err := yaml.Unmarshal(data, &v); err != nil {
		log.Fatal(err)
	}

	if err = sch.ValidateInterface(v); err != nil {
		log.Fatalf("%#v", err)
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
					fmt.Printf("sed -E -i 's@%s@https://elasticpath.dev%s@g' resources.yaml\n", link, matches[1])
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
					fmt.Printf("sed  -E -i 's@%s@%s@g resources.yaml'\n", link, rewriteUrlOne)
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

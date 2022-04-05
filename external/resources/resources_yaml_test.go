package resources

import (
	"net/http"
	"testing"
)

func TestResourceDocsExist(t *testing.T) {
	const httpStatusCodeOk = 200

	Resources := GetPluralResources()
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

	for link := range linksReferenceCount {
		response, err := http.DefaultClient.Head(link)
		if err != nil {
			t.Errorf("Error Retrieving Link\nLink: %s\nError Message: %s\nReference Count: %d", link, err, linksReferenceCount[link])
		} else {
			if response.StatusCode != httpStatusCodeOk {
				t.Errorf("Unexpected Response\nLink: %s\nExpected Status Code: %d\nActual Status Code: %d\nReference Count: %d",
					link, httpStatusCodeOk, response.StatusCode, linksReferenceCount[link])
			}
			if err := response.Body.Close(); err != nil {
				t.Errorf("Error Closing Reponse Body\nError Message: %s", err)
			}
		}
	}
}

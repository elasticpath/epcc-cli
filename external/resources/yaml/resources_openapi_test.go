package resources__test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/elasticpath/epcc-cli/external/openapi"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/yosida95/uritemplate/v3"
)

// TestOpenAPIOperationIDs verifies that all non-null openapi_operation_id values in resources.yaml
// actually exist in the OpenAPI specs, and that resources which should have an operation ID do have one
func TestOpenAPIOperationIDs(t *testing.T) {
	// Get all operation IDs from OpenAPI specs
	allOperationIDs, err := openapi.GetAllOperationIDs()
	if err != nil {
		t.Fatalf("Failed to get operation IDs from OpenAPI specs: %v", err)
	}

	t.Logf("Found %d operation IDs in OpenAPI specs", len(allOperationIDs))

	// Track results for reporting
	var invalidOperationIDs []struct {
		ResourceName  string
		OperationType string
		OperationID   string
	}

	// Create a map of normalized OpenAPI paths to operation IDs
	// This will help us find matching operations for resources
	openAPIPathMap := make(map[string]map[string][]struct {
		OperationID string
		SpecName    string
		Path        string
	})

	// Process all OpenAPI operations
	for opID, opInfo := range allOperationIDs {
		if opInfo.Path == "" || opInfo.Method == "" {
			continue
		}

		// Normalize the path
		normalizedPath, err := normalizeURLTemplate(extractPathPattern(opInfo.Path))
		if err != nil {
			t.Logf("Warning: Failed to normalize OpenAPI path %s: %v", opInfo.Path, err)
			continue
		}

		// Initialize the method map if needed
		if openAPIPathMap[normalizedPath] == nil {
			openAPIPathMap[normalizedPath] = make(map[string][]struct {
				OperationID string
				SpecName    string
				Path        string
			})
		}

		// Add the operation to the map
		openAPIPathMap[normalizedPath][opInfo.Method] = append(
			openAPIPathMap[normalizedPath][opInfo.Method],
			struct {
				OperationID string
				SpecName    string
				Path        string
			}{
				OperationID: opID,
				SpecName:    opInfo.SpecName,
				Path:        opInfo.Path,
			},
		)
	}

	// Track resources that should have an operation ID but don't
	var missingSuggestions []struct {
		ResourceName  string
		OperationType string
		Path          string
		Method        string
		SuggestedID   string
	}

	// Get all resources using the Resources API
	allResources := resources.GetPluralResources()

	// Check each resource operation
	for resourceName, resource := range allResources {
		// Check operations that might have openapi_operation_id
		operations := map[string]*resources.CrudEntityInfo{
			"get-collection": resource.GetCollectionInfo,
			"get-entity":     resource.GetEntityInfo,
			"create-entity":  resource.CreateEntityInfo,
			"update-entity":  resource.UpdateEntityInfo,
			"delete-entity":  resource.DeleteEntityInfo,
		}

		for opType, opInfo := range operations {
			if opInfo == nil {
				continue
			}

			// Check if the operation has an openapi_operation_id
			opID := opInfo.OpenApiOperationId

			// If it has an operation ID, verify it exists
			if opID != "" {
				if _, exists := allOperationIDs[opID]; !exists {
					invalidOperationIDs = append(invalidOperationIDs, struct {
						ResourceName  string
						OperationType string
						OperationID   string
					}{
						ResourceName:  resourceName,
						OperationType: opType,
						OperationID:   opID,
					})
				}
				continue // Skip further checks if it already has an operation ID
			}

			// If it doesn't have an operation ID, check if it should have one
			// by looking for matching paths in the OpenAPI specs
			path := extractPathPattern(opInfo.Url)
			method := resourceOperationToMethod(opType)

			// Skip if we couldn't determine the method
			if method == "" {
				continue
			}

			// Try to normalize the path
			normalizedPath, err := normalizeURLTemplate(path)
			if err != nil {
				t.Logf("Warning: Failed to normalize resource path %s: %v", path, err)
				continue
			}

			// Check if there's a matching path and method in the OpenAPI specs
			if methodMap, pathExists := openAPIPathMap[normalizedPath]; pathExists {
				if operations, methodExists := methodMap[method]; methodExists && len(operations) > 0 {
					// We found matching operations in the OpenAPI specs
					// This resource should have an operation ID

					// Generate a suggested operation ID based on the path and method
					// This follows the convention seen in the example: get-v2-accounts
					suggestedID := generateOperationID(method, path)

					// Check if the suggested ID exists in the OpenAPI specs
					var matchingOpID string
					for _, op := range operations {
						if op.OperationID == suggestedID {
							matchingOpID = op.OperationID
							break
						}
					}

					// If we didn't find an exact match, use the first operation ID
					if matchingOpID == "" {
						matchingOpID = operations[0].OperationID
					}

					// Add to the list of missing suggestions
					missingSuggestions = append(missingSuggestions, struct {
						ResourceName  string
						OperationType string
						Path          string
						Method        string
						SuggestedID   string
					}{
						ResourceName:  resourceName,
						OperationType: opType,
						Path:          path,
						Method:        method,
						SuggestedID:   matchingOpID,
					})
				}
			}
		}
	}

	// Report invalid operation IDs
	if len(invalidOperationIDs) > 0 {
		t.Errorf("Found %d invalid openapi_operation_id values:", len(invalidOperationIDs))
		for _, invalid := range invalidOperationIDs {
			t.Errorf("  - Resource: %s, Operation: %s, ID: %s",
				invalid.ResourceName, invalid.OperationType, invalid.OperationID)
		}
	} else {
		t.Logf("All openapi_operation_id values in resources.yaml are valid")
	}

	// Report missing operation IDs
	if len(missingSuggestions) > 0 {
		t.Errorf("Found %d resources that should have an openapi_operation_id but don't:", len(missingSuggestions))
		for _, missing := range missingSuggestions {
			t.Errorf("  - Resource: %s, Operation: %s, Path: %s, Method: %s, Suggested ID: %s",
				missing.ResourceName, missing.OperationType, missing.Path, missing.Method, missing.SuggestedID)
		}
	} else {
		t.Logf("All resources that should have an openapi_operation_id have one")
	}
}

// normalizeURLTemplate replaces all variable names in a URL template with a standard placeholder.
// This allows comparing URL templates that use different variable names but represent the same path structure.
// For example, "/v2/products/{id}" and "/v2/products/{product_id}" would both normalize to "/v2/products/{var}"
func normalizeURLTemplate(urlTemplate string) (string, error) {
	// Parse the URL template
	template, err := uritemplate.New(urlTemplate)
	if err != nil {
		return "", err
	}

	// Get all variable names in the template
	varNames := template.Varnames()

	// Replace each variable with a standard placeholder
	result := urlTemplate
	for _, varName := range varNames {
		// Create a pattern that matches the variable with its braces
		pattern := "{" + varName + "}"
		// Replace with a standard variable placeholder
		result = strings.Replace(result, pattern, "{var}", 1)
	}

	return result, nil
}

// extractPathPattern extracts the path pattern from a URL, removing any query parameters
// and normalizing the trailing slash
func extractPathPattern(url string) string {
	// Remove query parameters if present
	if idx := strings.Index(url, "?"); idx != -1 {
		url = url[:idx]
	}

	// Ensure consistent trailing slash handling
	url = strings.TrimSuffix(url, "/")

	return url
}

// pathsMatch determines if two URL paths match structurally, ignoring variable names
func pathsMatch(path1, path2 string) bool {
	// Extract path patterns
	pattern1 := extractPathPattern(path1)
	pattern2 := extractPathPattern(path2)

	// Normalize both templates
	normalized1, err1 := normalizeURLTemplate(pattern1)
	normalized2, err2 := normalizeURLTemplate(pattern2)

	// If either normalization fails, they don't match
	if err1 != nil || err2 != nil {
		return false
	}

	// Compare the normalized paths
	return normalized1 == normalized2
}

// resourceOperationToMethod maps resource operation types to HTTP methods
func resourceOperationToMethod(operationType string) string {
	switch operationType {
	case "get-collection", "get-entity":
		return "GET"
	case "create-entity":
		return "POST"
	case "update-entity":
		return "PUT"
	case "delete-entity":
		return "DELETE"
	default:
		return ""
	}
}

// generateOperationID creates an operation ID from a method and path
// following the convention seen in the example: get-v2-accounts
func generateOperationID(method, path string) string {
	// Convert method to lowercase
	method = strings.ToLower(method)

	// Remove leading slash and replace remaining slashes with hyphens
	path = strings.TrimPrefix(path, "/")
	path = strings.ReplaceAll(path, "/", "-")

	// Remove variable placeholders like {id}
	re := regexp.MustCompile(`\{[^}]+\}`)
	path = re.ReplaceAllString(path, "")

	// Remove any trailing hyphens
	path = strings.TrimSuffix(path, "-")

	// Combine method and path
	return fmt.Sprintf("%s-%s", method, path)
}

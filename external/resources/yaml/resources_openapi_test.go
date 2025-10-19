package resources__test

import (
	"testing"

	"github.com/elasticpath/epcc-cli/external/openapi"
	"github.com/elasticpath/epcc-cli/external/resources"
)

// TestOpenAPIOperationIDs verifies that all non-null openapi-operation-id values in resource yaml files
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

	// Track which operation IDs are referenced by resources
	referencedOperationIDs := make(map[string]bool)

	// Check each resource operation
	for resourceName, resource := range allResources {
		// Check operations that might have openapi-operation-id
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

			// Check if the operation has an openapi-operation-id
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
				} else {
					// Mark this operation ID as referenced
					referencedOperationIDs[opID] = true
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
		t.Errorf("Found %d invalid openapi-operation-id values:", len(invalidOperationIDs))
		for _, invalid := range invalidOperationIDs {
			t.Errorf("  - Resource: %s, Operation: %s, ID: %s",
				invalid.ResourceName, invalid.OperationType, invalid.OperationID)
		}
	} else {
		t.Logf("All openapi-operation-id values in resources yamls are valid")
	}

	// Report missing operation IDs
	if len(missingSuggestions) > 0 {
		t.Errorf("Found %d resources that should have an openapi-operation-id but don't:", len(missingSuggestions))
		for _, missing := range missingSuggestions {
			t.Errorf("  - Resource: %s, Operation: %s, Path: %s, Method: %s, Suggested ID: %s",
				missing.ResourceName, missing.OperationType, missing.Path, missing.Method, missing.SuggestedID)
		}
	} else {
		t.Logf("All resources that should have an openapi-operation-id have one")
	}

	// Find unreferenced operation IDs
	var unreferencedOperations []struct {
		OperationID string
		Method      string
		Path        string
		SpecName    string
	}

	for opID, opInfo := range allOperationIDs {
		if !referencedOperationIDs[opID] {
			unreferencedOperations = append(unreferencedOperations, struct {
				OperationID string
				Method      string
				Path        string
				SpecName    string
			}{
				OperationID: opID,
				Method:      opInfo.Method,
				Path:        opInfo.Path,
				SpecName:    opInfo.SpecName,
			})
		}
	}

	// Report unreferenced operation IDs (informational, not a failure)
	if len(unreferencedOperations) > 0 {
		t.Logf("Found %d operation IDs in OpenAPI specs that are not referenced by any resource:", len(unreferencedOperations))
		for _, unreferenced := range unreferencedOperations {
			t.Logf("  - ID: %s, Method: %s, Path: %s, Spec: %s",
				unreferenced.OperationID, unreferenced.Method, unreferenced.Path, unreferenced.SpecName)
		}
	} else {
		t.Logf("All operation IDs in OpenAPI specs are referenced by resources")
	}
}

// TestQueryParametersMatchOpenAPI validates that all query parameters defined in OpenAPI specs
// are also defined in the corresponding resource yaml file operations
func TestQueryParametersMatchOpenAPI(t *testing.T) {
	// Get all resources
	allResources := resources.GetPluralResources()

	// Track validation results
	var missingQueryParams []struct {
		ResourceName     string
		OperationType    string
		OperationID      string
		MissingParam     string
		OpenAPIParamName string
	}

	// Check each resource operation
	for resourceName, resource := range allResources {
		// Check operations that might have openapi-operation-id
		operations := map[string]*resources.CrudEntityInfo{
			"get-collection": resource.GetCollectionInfo,
			"get-entity":     resource.GetEntityInfo,
			"create-entity":  resource.CreateEntityInfo,
			"update-entity":  resource.UpdateEntityInfo,
			"delete-entity":  resource.DeleteEntityInfo,
		}

		for opType, opInfo := range operations {
			if opInfo == nil || opInfo.OpenApiOperationId == "" {
				// Skip operations without OpenAPI operation ID
				continue
			}

			// Get query parameters from OpenAPI spec
			openAPIQueryParams, err := openapi.GetQueryParametersForOperation(opInfo.OpenApiOperationId)
			if err != nil {
				t.Logf("Warning: Could not find OpenAPI operation %s: %v", opInfo.OpenApiOperationId, err)
				continue
			}

			// Convert resource query parameters to a set for easy lookup
			resourceQueryParams := make(map[string]bool)
			for _, param := range opInfo.QueryParameters {
				resourceQueryParams[param.Name] = true
			}

			// Check if all OpenAPI query parameters are defined in resources yaml files
			for _, openAPIParam := range openAPIQueryParams {
				if !resourceQueryParams[openAPIParam] {
					missingQueryParams = append(missingQueryParams, struct {
						ResourceName     string
						OperationType    string
						OperationID      string
						MissingParam     string
						OpenAPIParamName string
					}{
						ResourceName:     resourceName,
						OperationType:    opType,
						OperationID:      opInfo.OpenApiOperationId,
						MissingParam:     openAPIParam,
						OpenAPIParamName: openAPIParam,
					})
				}
			}
		}
	}

	// Report missing query parameters
	if len(missingQueryParams) > 0 {
		t.Errorf("Found %d query parameters in OpenAPI specs that are missing from resource yaml files:", len(missingQueryParams))
		for _, missing := range missingQueryParams {
			t.Errorf("  - Resource: %s, Operation: %s, OpenAPI ID: %s, Missing Parameter: %s",
				missing.ResourceName, missing.OperationType, missing.OperationID, missing.MissingParam)
		}
	} else {
		t.Logf("All query parameters from OpenAPI specs are properly defined in resource yaml files")
	}
}

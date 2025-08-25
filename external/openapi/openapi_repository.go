package openapi

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

//go:embed specs/*.yaml
var openapiFS embed.FS

// GetOpenAPISpec returns the content of an OpenAPI spec file by name
func GetOpenAPISpec(name string) ([]byte, error) {
	return openapiFS.ReadFile("specs/" + name)
}

// SpecModel represents an OpenAPI spec model with its metadata
type SpecModel struct {
	Name     string
	Document libopenapi.Document
	V3Model  v3.Document
}

// GetAllSpecModels returns all available OpenAPI specs as libopenapi models
func GetAllSpecModels() (map[string]*SpecModel, error) {
	specModels := make(map[string]*SpecModel)

	// List all YAML files in the specs directory
	entries, err := fs.ReadDir(openapiFS, "specs")
	if err != nil {
		return nil, fmt.Errorf("failed to read specs directory: %w", err)
	}

	// Process each YAML file
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		// Get the spec content
		specBytes, err := GetOpenAPISpec(entry.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read spec %s: %w", entry.Name(), err)
		}

		// Create a new document from specification bytes
		document, err := libopenapi.NewDocument(specBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to create document for %s: %w", entry.Name(), err)
		}

		// Build the V3 model
		v3ModelResult, errors := document.BuildV3Model()
		if len(errors) > 0 {
			errorMsgs := make([]string, len(errors))
			for i, e := range errors {
				errorMsgs[i] = e.Error()
			}
			return nil, fmt.Errorf("failed to build V3 model for %s: %s", entry.Name(), strings.Join(errorMsgs, "; "))
		}

		// Store the model
		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		specModels[name] = &SpecModel{
			Name:     name,
			Document: document,
			V3Model:  v3ModelResult.Model,
		}
	}

	return specModels, nil
}

// GetSpecModel returns a specific OpenAPI spec as a libopenapi model
func GetSpecModel(name string) (*SpecModel, error) {
	// If the name doesn't end with .yaml, add the extension
	if !strings.HasSuffix(name, ".yaml") {
		name = name + ".yaml"
	}

	// Get the spec content
	specBytes, err := GetOpenAPISpec(name)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec %s: %w", name, err)
	}

	// Create a new document from specification bytes
	document, err := libopenapi.NewDocument(specBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create document for %s: %w", name, err)
	}

	// Build the V3 model
	v3ModelResult, errors := document.BuildV3Model()
	if len(errors) > 0 {
		errorMsgs := make([]string, len(errors))
		for i, e := range errors {
			errorMsgs[i] = e.Error()
		}
		return nil, fmt.Errorf("failed to build V3 model for %s: %s", name, strings.Join(errorMsgs, "; "))
	}

	// Return the model
	baseName := strings.TrimSuffix(name, filepath.Ext(name))
	return &SpecModel{
		Name:     baseName,
		Document: document,
		V3Model:  v3ModelResult.Model,
	}, nil
}

// OperationInfo represents information about an operation in an OpenAPI spec
type OperationInfo struct {
	SpecName    string        // Name of the OpenAPI spec file (without extension)
	Path        string        // Path in the OpenAPI spec (e.g., "/v2/products/{id}")
	Method      string        // HTTP method (e.g., "GET", "POST", etc.)
	OperationID string        // The operationId from the OpenAPI spec
	Summary     string        // Summary description of the operation
	Operation   *v3.Operation // The actual OpenAPI v3 Operation object
}

// FindOperationByID searches all OpenAPI specs for an operation with the given ID
// and returns information about the operation if found.
//
// Example:
//
//	opInfo, err := FindOperationByID("getProduct")
//	if err != nil {
//	    log.Fatalf("Operation not found: %v", err)
//	}
//	fmt.Printf("Found operation in %s at path %s using method %s\n", 
//	    opInfo.SpecName, opInfo.Path, opInfo.Method)
func FindOperationByID(operationID string) (*OperationInfo, error) {
	// Get all spec models
	specModels, err := GetAllSpecModels()
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenAPI specs: %w", err)
	}

	// Search for the operation ID in each spec
	for specName, spec := range specModels {
		// Check each path
		paths := spec.V3Model.Paths
		if paths == nil || paths.PathItems == nil {
			continue
		}

		// Use a more idiomatic approach with a found flag
		var found bool
		var result *OperationInfo

		// Get the iterator for paths
		iter := paths.PathItems.FromOldest()

		for path, pathItem := range iter {
			// Skip if we already found a match or if path is invalid
			if found || path == "" {
				continue
			}

			// Check each HTTP method for the operation ID
			methods := []struct {
				Method    string
				Operation *v3.Operation
			}{
				{"GET", pathItem.Get},
				{"POST", pathItem.Post},
				{"PUT", pathItem.Put},
				{"DELETE", pathItem.Delete},
				{"OPTIONS", pathItem.Options},
				{"HEAD", pathItem.Head},
				{"PATCH", pathItem.Patch},
				{"TRACE", pathItem.Trace},
			}

			for _, m := range methods {
				if m.Operation != nil && m.Operation.OperationId == operationID {
					// We found a match
					found = true
					result = &OperationInfo{
						SpecName:    specName,
						Path:        path,
						Method:      m.Method,
						OperationID: operationID,
						Summary:     m.Operation.Summary,
						Operation:   m.Operation,
					}
					break
				}
			}

			// If we found a match, no need to check more paths
			if found {
				break
			}
		}

		// If we found a match, return it
		if found {
			return result, nil
		}
	}

	return nil, fmt.Errorf("operation ID '%s' not found in any OpenAPI spec", operationID)
}

// GetQueryParametersForOperation returns the query parameter names for a given operation ID
func GetQueryParametersForOperation(operationID string) ([]string, error) {
	// Find the operation using the existing function
	opInfo, err := FindOperationByID(operationID)
	if err != nil {
		return nil, err
	}

	// Extract query parameters directly from the embedded Operation
	var queryParams []string
	if opInfo.Operation != nil && opInfo.Operation.Parameters != nil {
		for _, param := range opInfo.Operation.Parameters {
			if param != nil && param.In == "query" {
				queryParams = append(queryParams, param.Name)
			}
		}
	}
	
	return queryParams, nil
}

// OperationIDInfo contains information about an operation ID and where it's defined
type OperationIDInfo struct {
	SpecName    string // Name of the OpenAPI spec file (without extension)
	Path        string // Path in the OpenAPI spec (e.g., "/v2/products/{id}")
	Method      string // HTTP method (e.g., "GET", "POST", etc.)
	Summary     string // Summary description of the operation
}

// GetAllOperationIDs returns a map of all operation IDs found in all OpenAPI specs.
// The map key is the operation ID and the value is information about where it's defined.
// This is useful for validating operation IDs referenced in other parts of the codebase.
//
// Example:
//
//	allOperationIDs, err := GetAllOperationIDs()
//	if err != nil {
//	    log.Fatalf("Failed to get operation IDs: %v", err)
//	}
//	
//	// Check if a specific operation ID exists
//	if opInfo, exists := allOperationIDs["createProduct"]; exists {
//	    fmt.Printf("Operation found in %s at path %s using method %s\n", 
//	        opInfo.SpecName, opInfo.Path, opInfo.Method)
//	}
//	
//	// Print all operation IDs
//	for opID, opInfo := range allOperationIDs {
//	    fmt.Printf("%s: %s %s in %s\n", opID, opInfo.Method, opInfo.Path, opInfo.SpecName)
//	}
func GetAllOperationIDs() (map[string]OperationIDInfo, error) {
	// Get all spec models
	specModels, err := GetAllSpecModels()
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenAPI specs: %w", err)
	}

	// Map to store all operation IDs
	operationIDs := make(map[string]OperationIDInfo)

	// Search for operation IDs in each spec
	for specName, spec := range specModels {
		// Check each path
		paths := spec.V3Model.Paths
		if paths == nil || paths.PathItems == nil {
			continue
		}

		// Get the iterator for paths
		iter := paths.PathItems.FromOldest()

		for path, pathItem := range iter {
			// Skip if path is invalid
			if path == "" {
				continue
			}

			// Check each HTTP method for operation IDs
			methods := []struct {
				Method    string
				Operation *v3.Operation
			}{
				{"GET", pathItem.Get},
				{"POST", pathItem.Post},
				{"PUT", pathItem.Put},
				{"DELETE", pathItem.Delete},
				{"OPTIONS", pathItem.Options},
				{"HEAD", pathItem.Head},
				{"PATCH", pathItem.Patch},
				{"TRACE", pathItem.Trace},
			}

			for _, m := range methods {
				if m.Operation != nil && m.Operation.OperationId != "" {
					// Store the operation ID with its information
					operationIDs[m.Operation.OperationId] = OperationIDInfo{
						SpecName: specName,
						Path:     path,
						Method:   m.Method,
						Summary:  m.Operation.Summary,
					}
				}
			}
		}
	}

	return operationIDs, nil
}

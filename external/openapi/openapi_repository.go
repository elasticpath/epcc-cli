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
	SpecName    string
	Path        string
	Method      string
	OperationID string
	Summary     string
}

// FindOperationByID searches all OpenAPI specs for an operation with the given ID
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
			// Get the current key and value

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

package resources__test

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/yosida95/uritemplate/v3"
)

// normalizeURLTemplate replaces all variable names in a URL template with a standard placeholder.
// This allows comparing URL templates that use different variable names but represent the same path structure.
// For example, "/v2/products/{id}" and "/v2/products/{product_id}" would both normalize to "/v2/products/{var}"
func normalizeURLTemplate(urlTemplate string) (string, error) {
	// Parse the URL template

	sanitizedPath := sanitizePath(urlTemplate)
	template, err := uritemplate.New(sanitizedPath)
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

// pathVarSanitizer is a regex that matches variable names in OpenAPI path templates
var pathVarSanitizer = regexp.MustCompile(`\{([^}]*)\}`)

// illegalCharReplacer is a regex that replaces non-alphanumeric/underscore characters with underscores
var illegalCharReplacer = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// sanitizePath replaces illegal characters in URI template variable names with underscores
// This is needed because the uritemplate package has stricter variable name requirements than OpenAPI
func sanitizePath(path string) string {
	return pathVarSanitizer.ReplaceAllStringFunc(path, func(match string) string {
		// Extract the variable name without the braces
		varName := match[1 : len(match)-1]
		// Replace any non-alphanumeric/underscore characters with underscores
		sanitizedVarName := illegalCharReplacer.ReplaceAllString(varName, "_")
		// Return the sanitized variable name with braces
		return "{" + sanitizedVarName + "}"
	})
}

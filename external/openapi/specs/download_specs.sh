#!/bin/bash

# Script to download OpenAPI specs to the specs directory
# This script will work regardless of the directory from which it is invoked

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Function to download a spec file
download_spec() {
    local url="$1"
    local filename="$2"
    local output_path="${SCRIPT_DIR}/${filename}"

    echo "Downloading ${filename} from ${url}..."

    # Download the file using curl
    if curl -s --fail "${url}" -o "${output_path}"; then
        echo "✅ Successfully downloaded ${filename}"
    else
        echo "❌ Failed to download ${filename} from ${url}"
        return 1
    fi

    return 0
}

# Create a log header
echo "=== OpenAPI Spec Downloader ==="
echo "Downloading specs to: ${SCRIPT_DIR}"
echo "Started at: $(date)"
echo "==============================="

# Add your OpenAPI spec URLs here
# Format: download_spec "URL" "filename.yaml"

# Example URL provided
download_spec "https://elasticpath.dev/assets/openapispecs/carts/OpenAPISpec.yaml" "carts-and-orders.yaml"
download_spec "https://elasticpath.dev/assets/openapispecs/accounts/OpenAPISpec.yaml" "account-management.yaml"
download_spec "https://elasticpath.dev/assets/openapispecs/promotions-builder/OpenAPISpec.yaml" "promotions-builder.yaml"

# Integrations doesn't have matching URLs.
#download_spec "https://elasticpath.dev/assets/openapispecs/integrations/openapi.yaml" "integrations.yaml"

# Add more URLs as needed:
# download_spec "https://api.example.org/openapi.yaml" "example-org-api.yaml"
# download_spec "https://another-api.com/spec.yaml" "another-api.yaml"

echo "==============================="
echo "Download process completed at: $(date)"

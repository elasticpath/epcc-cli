#!/usr/bin/env bash

# Round-trip smoke test for the get-all command
# 1. Create resources from different families
# 2. Export them with a single get-all --output-format epcc-cli
# 3. Delete the resources
# 4. Run the exported script to recreate them
# 5. Verify resources were recreated

set -e

TEMP_DIR=$(mktemp -d)

cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

echo "=== get-all Round-Trip Smoke Test ==="
epcc reset-store .+

# Step 1: Create resources from different families
echo "=== Step 1: Creating test resources ==="

# Account and sub-resource (account-address)
epcc create account name "get-all-test-account" legal_name "Test Account for get-all"
epcc create account-address account/name=get-all-test-account name "Test Address" first_name "John" last_name "Doe" line_1 "123 Test St" city "Test City" postcode "H0H 0H0" county "Test County" country "US"

# Customer
epcc create customer name "get-all-test-customer" email "get-all-test@example.com"


# Custom API with fields and entry
epcc create custom-api name "smoke-test-api" slug "smoke-test-api" api_type "smoke_test_ext" description "Smoke test API"
epcc create custom-field custom_api/slug=smoke-test-api name "test_string" description blah slug "test_string" field_type "string"
epcc create custom-field custom_api/slug=smoke-test-api name "test_int" description blah slug "test_int" field_type "integer"
epcc create custom-field custom_api/slug=smoke-test-api name "test_bool" description blah slug "test_bool" field_type "boolean"
epcc create custom-api-settings-entry custom_api/slug=smoke-test-api data.test_string "hello world" data.test_int 42 data.test_bool true data.type "smoke_test_ext"

# PCM resources with hierarchy and node-product relationships
epcc create pcm-product --auto-fill name "Smoke Test Product"
epcc create pcm-hierarchy name "smoke-test-hierarchy"
epcc create pcm-node name=smoke-test-hierarchy name "Parent Node" --auto-fill
epcc create pcm-node name=smoke-test-hierarchy name "Child Node" --auto-fill relationships.parent.data.id name=Parent_Node
epcc create pcm-node-product name=smoke-test-hierarchy name=Child_Node data\[0\].id last_read=entity

echo "=== Step 2: Export all resources with single get-all ==="

epcc get-all accounts account-addresses customers custom-apis custom-fields custom-api-settings-entries \
    pcm-products pcm-hierarchies pcm-nodes pcm-node-products \
    --output-file "$TEMP_DIR/export.sh" --output-format epcc-cli --truncate-output

echo "=== Step 3: Delete resources ==="

# Delete in reverse dependency order
epcc delete-all custom-api-settings-entries
epcc delete-all custom-fields
epcc delete-all custom-apis
epcc delete-all customers
epcc delete-all account-addresses
epcc delete-all accounts
epcc delete-all pcm-nodes
epcc delete-all pcm-hierarchies
epcc delete-all pcm-products

echo "=== Step 4: Run exported script to recreate resources ==="

"$TEMP_DIR/export.sh"

echo "=== Step 5: Verify resources were recreated ==="

# Check account exists (use collection query since aliases aren't saved with --skip-alias-processing)
# Note: --output-jq returns JSON-formatted values (strings have quotes), so we strip them with tr
ACCOUNT_ID=$(epcc get accounts --output-jq '.data[] | select(.name == "get-all-test-account") | .id' 2>/dev/null | tr -d '"' || echo "")
if [ -z "$ACCOUNT_ID" ]; then
    echo "FAIL: Account get-all-test-account not found after recreation"
    exit 1
fi
echo "PASS: Account recreated (id: $ACCOUNT_ID)"

# Check account-address exists (use ID since aliases aren't saved with --skip-alias-processing)
ADDRESS_COUNT=$(epcc get account-addresses "$ACCOUNT_ID" --output-jq '.meta.results.total' 2>/dev/null | tr -d '"' || echo "0")
if [ "$ADDRESS_COUNT" -lt 1 ]; then
    echo "FAIL: No account-addresses found after recreation"
    exit 1
fi
echo "PASS: Account-addresses recreated ($ADDRESS_COUNT found)"

# Check customer exists
CUSTOMER_COUNT=$(epcc get customers --output-jq '.meta.results.total' | tr -d '"')
if [ "$CUSTOMER_COUNT" -lt 1 ]; then
    echo "FAIL: No customers found after recreation"
    exit 1
fi
echo "PASS: Customers recreated ($CUSTOMER_COUNT found)"

# Check custom-api exists (use collection query since aliases aren't saved with --skip-alias-processing)
CUSTOM_API_ID=$(epcc get custom-apis --output-jq '.data[] | select(.slug == "smoke-test-api") | .id' 2>/dev/null | tr -d '"' || echo "")
if [ -z "$CUSTOM_API_ID" ]; then
    echo "FAIL: Custom API smoke-test-api not found after recreation"
    epcc get custom-apis
    exit 1
fi
echo "PASS: Custom API recreated (id: $CUSTOM_API_ID)"

# Check custom-fields exist (use ID since aliases aren't saved with --skip-alias-processing)
FIELD_COUNT=$(epcc get custom-fields "$CUSTOM_API_ID" --output-jq '.meta.results.total' 2>/dev/null | tr -d '"' || echo "0")
if [ "$FIELD_COUNT" -lt 3 ]; then
    echo "FAIL: Expected at least 3 custom-fields, found $FIELD_COUNT"
    exit 1
fi
echo "PASS: Custom fields recreated ($FIELD_COUNT found)"

# Check custom-api-settings-entry exists (use ID since aliases aren't saved with --skip-alias-processing)
ENTRY_COUNT=$(epcc get custom-api-settings-entries "$CUSTOM_API_ID" --output-jq '.meta.results.total' 2>/dev/null | tr -d '"' || echo "0")
if [ "$ENTRY_COUNT" -lt 1 ]; then
    echo "FAIL: No custom-api-settings-entries found after recreation"
    exit 1
fi
echo "PASS: Custom API entries recreated ($ENTRY_COUNT found)"

# Check pcm-product exists
PRODUCT_ID=$(epcc get pcm-products --output-jq '.data[] | select(.attributes.name == "Smoke Test Product") | .id' 2>/dev/null | tr -d '"' || echo "")
if [ -z "$PRODUCT_ID" ]; then
    echo "FAIL: PCM product 'Smoke Test Product' not found after recreation"
    exit 1
fi
echo "PASS: PCM product recreated (id: $PRODUCT_ID)"

# Check pcm-hierarchy exists
HIERARCHY_ID=$(epcc get pcm-hierarchies --output-jq '.data[] | select(.attributes.name == "smoke-test-hierarchy") | .id' 2>/dev/null | tr -d '"' || echo "")
if [ -z "$HIERARCHY_ID" ]; then
    echo "FAIL: PCM hierarchy 'smoke-test-hierarchy' not found after recreation"
    exit 1
fi
echo "PASS: PCM hierarchy recreated (id: $HIERARCHY_ID)"

# Check pcm-nodes exist
NODE_COUNT=$(epcc get pcm-nodes "$HIERARCHY_ID" --output-jq '.data | length' 2>/dev/null | tr -d '"' || echo "0")
if [ "$NODE_COUNT" -lt 2 ]; then
    echo "FAIL: Expected at least 2 pcm-nodes, found $NODE_COUNT"
    exit 1
fi
echo "PASS: PCM nodes recreated ($NODE_COUNT found)"

# Check pcm-node-products exist (get child node ID first)
CHILD_NODE_ID=$(epcc get pcm-nodes "$HIERARCHY_ID" --output-jq '.data[] | select(.attributes.name == "Child Node") | .id' 2>/dev/null | tr -d '"' || echo "")
if [ -n "$CHILD_NODE_ID" ]; then
    NODE_PRODUCT_COUNT=$(epcc get pcm-node-products "$HIERARCHY_ID" "$CHILD_NODE_ID" --output-jq '.data | length' 2>/dev/null | tr -d '"' || echo "0")
    if [ "$NODE_PRODUCT_COUNT" -lt 1 ]; then
        echo "FAIL: No pcm-node-products found after recreation"
        exit 1
    fi
    echo "PASS: PCM node-products recreated ($NODE_PRODUCT_COUNT found)"
else
    echo "FAIL: Child node not found for pcm-node-products verification"
    exit 1
fi

echo ""
echo "=== get-all Round-Trip Smoke Test PASSED ==="

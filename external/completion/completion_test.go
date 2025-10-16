package completion

import (
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"testing"
)

func init() {
	resources.PublicInit()
}

func TestCompletePluralResourcesCompletesWithNoVerb(t *testing.T) {
	// Fixture Setup
	toComplete := "cus"
	request := Request{
		Type:       CompletePluralResource,
		ToComplete: toComplete,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "customers")
}

func TestCompletePluralResourcesCompletesWithGet(t *testing.T) {
	// Fixture Setup
	toComplete := "cus"
	request := Request{
		Type:       CompletePluralResource,
		ToComplete: toComplete,
		Verb:       Get,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "customers")
	// There is no get method defined for this one
	require.NotContains(t, completions, "customer-tokens")
}

func TestCompletePluralResourcesCompletesWithDelete(t *testing.T) {
	// Fixture Setup
	toComplete := "cus"
	request := Request{
		Type:       CompletePluralResource,
		ToComplete: toComplete,
		Verb:       Delete,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "customers")
	// There is no delete method defined for this one
	require.NotContains(t, completions, "merchant-realm-mappings")
}

func TestCompletePluralResourcesCompletesWithDeleteAll(t *testing.T) {
	// Fixture Setup
	toComplete := "cus"
	request := Request{
		Type:       CompletePluralResource,
		ToComplete: toComplete,
		Verb:       DeleteAll,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "customers")
	// There is no get method defined for this one
	require.NotContains(t, completions, "merchant-realm-mappings")
	require.NotContains(t, completions, "customer-tokens")
}

func TestCompleteSingularResourcesCompletes(t *testing.T) {
	// Fixture Setup
	toComplete := "cus"
	request := Request{
		Type:       CompleteSingularResource,
		ToComplete: toComplete,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "customer")
	require.Contains(t, completions, "merchant-realm-mapping")
	require.Contains(t, completions, "customer-token")
}

func TestCompleteSingularResourcesCompletesWithCreate(t *testing.T) {
	// Fixture Setup
	toComplete := "cus"
	request := Request{
		Type:       CompleteSingularResource,
		ToComplete: toComplete,
		Verb:       Create,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "customer")
	// No create defined for merchant realm mappings
	require.NotContains(t, completions, "merchant-realm-mapping")
	require.Contains(t, completions, "customer-token")
}

func TestCompleteSingularResourcesCompletesWithUpdate(t *testing.T) {
	// Fixture Setup
	toComplete := "cus"
	request := Request{
		Type:       CompleteSingularResource,
		ToComplete: toComplete,
		Verb:       Update,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "customer")
	require.Contains(t, completions, "merchant-realm-mapping")
	// No update defined for merchant realm mappings
	require.NotContains(t, completions, "customer-token")
}

func TestCompleteSingularResourcesCompletesWithDelete(t *testing.T) {
	// Fixture Setup
	toComplete := "cus"
	request := Request{
		Type:       CompleteSingularResource,
		ToComplete: toComplete,
		Verb:       Delete,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "customer")
	// No update defined for merchant realm mappings
	require.NotContains(t, completions, "merchant-realm-mapping")
	require.NotContains(t, completions, "customer-token")
}

func TestCompleteSingularResourcesCompletesWithGet(t *testing.T) {
	// Fixture Setup
	toComplete := "cus"
	request := Request{
		Type:       CompleteSingularResource,
		ToComplete: toComplete,
		Verb:       Get,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "customer")
	require.Contains(t, completions, "merchant-realm-mapping")
	// No get defined for merchant realm mappings
	require.NotContains(t, completions, "customer-token")
}

func TestCompleteCrudActions(t *testing.T) {
	// Fixture Setup
	toComplete := "cus"
	request := Request{
		Type:       CompleteCrudAction,
		ToComplete: toComplete,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "create")
	require.Contains(t, completions, "update")
	require.Contains(t, completions, "delete")
	require.Contains(t, completions, "get")
	require.Len(t, completions, 4)
}

func TestCompleteLoginLogoutApiActions(t *testing.T) {
	// Fixture Setup
	toComplete := "cus"
	request := Request{
		Type:       CompleteLoginLogoutAPI,
		ToComplete: toComplete,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "api")
	require.Len(t, completions, 1)
}

func TestCompleteBool(t *testing.T) {
	// Fixture Setup
	toComplete := "cus"
	request := Request{
		Type:       CompleteBool,
		ToComplete: toComplete,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "true")
	require.Contains(t, completions, "false")
	require.Len(t, completions, 2)
}

func TestCompleteLoginClientID(t *testing.T) {
	// Fixture Setup
	toComplete := "cus"
	request := Request{
		Type:       CompleteLoginClientID,
		ToComplete: toComplete,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "client_id")
	require.Len(t, completions, 1)
}

func TestCompleteLoginClientSecret(t *testing.T) {
	// Fixture Setup
	toComplete := "cus"
	request := Request{
		Type:       CompleteLoginClientSecret,
		ToComplete: toComplete,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "client_secret")
	require.Len(t, completions, 1)
}

func TestCompleteQueryParamKey(t *testing.T) {
	// Fixture Setup
	toComplete := "inc"
	resource, _ := resources.GetResourceByName("carts")
	request := Request{
		Type:       CompleteQueryParamKey,
		ToComplete: toComplete,
		Resource:   resource,
		Verb:       Get,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "include")
}

func TestCompleteQueryParamKeyWithGetAll(t *testing.T) {
	// Fixture Setup
	toComplete := "so"
	resource, _ := resources.GetResourceByName("customers")
	request := Request{
		Type:       CompleteQueryParamKey,
		ToComplete: toComplete,
		Resource:   resource,
		Verb:       GetAll,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "sort")
	require.Contains(t, completions, "filter")
	require.Contains(t, completions, "include")
	require.Contains(t, completions, "page[limit]")
	require.Contains(t, completions, "page[offset]")
	require.Contains(t, completions, "page[total_method]")
}

func TestCompleteAlias(t *testing.T) {
	// Fixture Setup
	toComplete := "cus"
	resource, _ := resources.GetResourceByName("customers")
	request := Request{
		Type:       CompleteAlias,
		ToComplete: toComplete,
		Resource:   resource,
	}

	err := aliases.ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	aliases.SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "customer",
		"name":  "John Smith"
		}
}`)

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "name=John_Smith")
}

func TestCompleteQueryParamValue(t *testing.T) {
	// Fixture Setup
	toComplete := "na"
	resource, _ := resources.GetResourceByName("customers")
	request := Request{
		Type:       CompleteQueryParamValue,
		ToComplete: toComplete,
		Resource:   resource,
		Verb:       GetAll,
		QueryParam: "sort",
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "name")
	require.Contains(t, completions, "-name")
	require.Contains(t, completions, "updated_at")
	require.Contains(t, completions, "-updated_at")
	require.Contains(t, completions, "created_at")
	require.Contains(t, completions, "-created_at")
}

func TestCompleteQueryParamValueWithPageTotalMethod(t *testing.T) {
	// Fixture Setup
	toComplete := "ex"
	resource, _ := resources.GetResourceByName("customers")
	request := Request{
		Type:       CompleteQueryParamValue,
		ToComplete: toComplete,
		Resource:   resource,
		Verb:       GetAll,
		QueryParam: "page[total_method]",
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "exact")
	require.Contains(t, completions, "estimate")
	require.Contains(t, completions, "lower_bound")
	require.Contains(t, completions, "observed")
	require.Contains(t, completions, "cached")
	require.Contains(t, completions, "none")
}

func TestCompleteQueryParamValueWithFilter(t *testing.T) {
	// Fixture Setup
	toComplete := "eq"
	resource, _ := resources.GetResourceByName("customers")
	request := Request{
		Type:       CompleteQueryParamValue,
		ToComplete: toComplete,
		Resource:   resource,
		Verb:       GetAll,
		QueryParam: "filter",
	}

	// Exercise SUT
	_, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp|cobra.ShellCompDirectiveNoSpace)
	// Filter completions are handled by GetFilterCompletion function
	// This test verifies the completion type works and sets NoSpace directive
}

func TestCompleteLoginAccountManagementKey(t *testing.T) {
	// Fixture Setup
	toComplete := "acc"
	request := Request{
		Type:       CompleteLoginAccountManagementKey,
		ToComplete: toComplete,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "account_id")
	require.Contains(t, completions, "account_name")
	require.Len(t, completions, 2)
}

func TestHeaderKeyWithNilValueCompletes(t *testing.T) {
	// Fixture Setup
	toComplete := "EP-"
	request := Request{
		Type:       CompleteHeaderKey,
		ToComplete: toComplete,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "EP-Beta-Features")
}

func TestHeaderKeyWithNonNilValueCompletes(t *testing.T) {
	// Fixture Setup
	toComplete := "X-Moltin"
	request := Request{
		Type:       CompleteHeaderKey,
		ToComplete: toComplete,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "X-Moltin-Currency")
}

func TestHeaderValueWithNilValueCompletesWithoutPanicing(t *testing.T) {
	// Fixture Setup
	toComplete := "ac"
	request := Request{
		Type:       CompleteHeaderValue,
		ToComplete: toComplete,
		Header:     "EP-Beta-Features",
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Empty(t, completions)

}

func TestHeaderValueWithNonNilValueCompletes(t *testing.T) {
	// Fixture Setup
	toComplete := "U"
	request := Request{
		Type:       CompleteHeaderValue,
		ToComplete: toComplete,
		Header:     "X-Moltin-Currency",
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "USD")
}

func TestAttributeValueWithNoTemplating(t *testing.T) {
	// Fixture Setup
	toComplete := ""
	acct := resources.MustGetResourceByName("password-profiles")
	request := Request{
		Type:       CompleteAttributeValue,
		Verb:       Create,
		ToComplete: toComplete,
		Attribute:  "username_format",
		Resource:   acct,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "any")
	require.Contains(t, completions, "email")
	require.Equal(t, 2, len(completions))
}

func TestAttributeValueWithTemplating(t *testing.T) {
	// Fixture Setup
	toComplete := ""
	acct := resources.MustGetResourceByName("password-profiles")
	request := Request{
		Type:           CompleteAttributeValue,
		Verb:           Create,
		ToComplete:     toComplete,
		Attribute:      "username_format",
		Resource:       acct,
		AllowTemplates: true,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "any")
	require.Contains(t, completions, "email")
	require.Contains(t, completions, `{{\ randAlphaNum\ |`)
	require.Contains(t, completions, `{{\ randAlphaNum\ }}`)
}

func TestAttributeValueWithTemplatingAndPipe(t *testing.T) {
	// Fixture Setup
	toComplete := "{{ randAlphaNum 3 | "
	acct := resources.MustGetResourceByName("password-profiles")
	request := Request{
		Type:           CompleteAttributeValue,
		Verb:           Create,
		ToComplete:     toComplete,
		Attribute:      "username_format",
		Resource:       acct,
		AllowTemplates: true,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "any")
	require.Contains(t, completions, "email")
	require.Contains(t, completions, `{{\ randAlphaNum\ 3\ |\ upper\ |`)
	require.Contains(t, completions, `{{\ randAlphaNum\ 3\ |\ lower\ }}`)
}

func TestCompleteAttributeKeyWithEmptyExistingValuesReturnsAll(t *testing.T) {
	// Fixture Setup
	toComplete := ""
	acct := resources.MustGetResourceByName("accounts")
	request := Request{
		Type:       CompleteAttributeKey,
		Verb:       Create,
		ToComplete: toComplete,
		Resource:   acct,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "name")
	require.Contains(t, completions, "legal_name")
	require.Contains(t, completions, "registration_id")
	require.Contains(t, completions, "parent_id")
	require.Len(t, completions, 4)
}

func TestCompleteAttributeKeyWithTwoUsedValuesExistingValuesReturnsRemaining(t *testing.T) {
	// Fixture Setup
	toComplete := ""
	acct := resources.MustGetResourceByName("accounts")
	request := Request{
		Type:       CompleteAttributeKey,
		Verb:       Create,
		ToComplete: toComplete,
		Resource:   acct,
		Attributes: map[string]struct{}{
			"name":       {},
			"legal_name": {},
		},
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "registration_id")
	require.Contains(t, completions, "parent_id")
	require.Len(t, completions, 2)
}

func TestCompleteAttributeKeyWithWildcardReturnsCompletedAdjacentValues(t *testing.T) {
	// Fixture Setup
	toComplete := ""
	acct := resources.MustGetResourceByName("pcm-products")
	request := Request{
		Type:       CompleteAttributeKey,
		Verb:       Create,
		ToComplete: toComplete,
		Resource:   acct,
		Attributes: map[string]struct{}{
			"name":                        {},
			"sku":                         {},
			"custom_inputs.foo.name":      {},
			"components.bar.options.type": {},
		},
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "custom_inputs.foo.validation_rules[0].required")
	require.Contains(t, completions, "components.bar.min")
}

func TestCompleteAttributeKeyWithWildcardReturnsIncrementedArrayIndexes(t *testing.T) {
	// Fixture Setup
	toComplete := ""
	acct := resources.MustGetResourceByName("pcm-products")
	request := Request{
		Type:       CompleteAttributeKey,
		Verb:       Create,
		ToComplete: toComplete,
		Resource:   acct,
		Attributes: map[string]struct{}{
			"name":                         {},
			"sku":                          {},
			"components.bar.options.id[0]": {},
		},
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "components.bar.options.id[1]")
}

func TestCompleteAttributeKeyWithWithMultipleArrayIndexesIncrementsAppropriately(t *testing.T) {
	// Fixture Setup
	toComplete := ""
	acct := resources.MustGetResourceByName("rule-promotions")
	request := Request{
		Type:       CompleteAttributeKey,
		Verb:       Create,
		ToComplete: toComplete,
		Resource:   acct,
		Attributes: map[string]struct{}{
			"rule_set.rules.children[0].args[0]": {},
		},
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "rule_set.rules.children[0].args[1]")
	require.Contains(t, completions, "rule_set.rules.children[1].args[0]")

}

func TestCompleteQueryParamKeyGetCollectionWithExplicitParams(t *testing.T) {
	// Fixture Setup
	toComplete := "pa"
	resource, _ := resources.GetResourceByName("account-members")
	request := Request{
		Type:       CompleteQueryParamKey,
		ToComplete: toComplete,
		Resource:   resource,
		Verb:       GetAll,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	// Should only contain explicitly defined parameters, not hardcoded fallbacks
	require.Contains(t, completions, "page[limit]")
	require.Contains(t, completions, "page[offset]")
	require.Contains(t, completions, "sort")
	require.Contains(t, completions, "filter")
	// Should NOT contain hardcoded fallbacks like page[total_method] or include
	require.NotContains(t, completions, "page[total_method]")
	require.NotContains(t, completions, "include")
}

func TestCompleteQueryParamKeyGetCollectionWithFallbackParams(t *testing.T) {
	// Fixture Setup
	toComplete := "pa"
	resource, _ := resources.GetResourceByName("customers")
	request := Request{
		Type:       CompleteQueryParamKey,
		ToComplete: toComplete,
		Resource:   resource,
		Verb:       GetAll,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	// Should contain hardcoded fallback parameters since no explicit params defined
	require.Contains(t, completions, "sort")
	require.Contains(t, completions, "filter")
	require.Contains(t, completions, "include")
	require.Contains(t, completions, "page[limit]")
	require.Contains(t, completions, "page[offset]")
	require.Contains(t, completions, "page[total_method]")
}

func TestCompleteQueryParamKeyGetEntityWithExplicitParams(t *testing.T) {
	// Fixture Setup
	toComplete := "inc"
	resource, _ := resources.GetResourceByName("account-memberships")
	request := Request{
		Type:       CompleteQueryParamKey,
		ToComplete: toComplete,
		Resource:   resource,
		Verb:       Get,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	// Should only contain explicitly defined parameter
	require.Contains(t, completions, "include")
	// Should be exactly 1 completion since only "include" is explicitly defined
	require.Len(t, completions, 1)
}

func TestCompleteQueryParamKeyGetEntityWithFallbackParams(t *testing.T) {
	// Fixture Setup
	toComplete := "inc"
	resource, _ := resources.GetResourceByName("customers")
	request := Request{
		Type:       CompleteQueryParamKey,
		ToComplete: toComplete,
		Resource:   resource,
		Verb:       Get,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	// Should contain hardcoded fallback parameter since no explicit params defined
	require.Contains(t, completions, "include")
	// Should be exactly 1 completion since only "include" is the fallback for get-entity
	require.Len(t, completions, 1)
}

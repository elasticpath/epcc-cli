package completion

import (
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/stretchr/testify/require"
	"testing"
)

type filteringTestCase struct {
	title               string
	filter              string
	expectedCompletions []string
}

func TestFilteringCompletion(t *testing.T) {
	testCases := []filteringTestCase{
		{"Empty Filter", "", []string{"eq(", "lt(", "gt(", "ge(", "le(", "in(", "like("}},
		{"Prefix of operator", "e", []string{"eq(", "lt(", "gt(", "ge(", "le(", "in(", "like("}},
		{"Full binary operator match", "eq(", []string{"eq(name,", "eq(legal_name,", "eq(registration_id,", "eq(parent_id,", "eq(id,", "eq(updated_at,", "eq(created_at,"}},
		{"Full variable operator match", "in(", []string{"in(name,", "in(legal_name,", "in(registration_id,", "in(parent_id,", "in(id,", "in(updated_at,", "in(created_at,"}},
		// The shell is responsible for filtering out things that don't match
		{"Full binary operator match with attribute prefix", "eq(n", []string{"eq(name,", "eq(legal_name,", "eq(registration_id,", "eq(parent_id,", "eq(id,", "eq(updated_at,", "eq(created_at,"}},
		{"Full variable operator match with attribute prefix", "in(n", []string{"in(name,", "in(legal_name,", "in(registration_id,", "in(parent_id,", "in(id,", "in(updated_at,", "in(created_at,"}},
		{"Full binary operator match with full attribute", "eq(name", []string{"eq(name,", "eq(legal_name,", "eq(registration_id,", "eq(parent_id,", "eq(id,", "eq(updated_at,", "eq(created_at,"}},
		{"Full vararg operator match with full attribute", "in(name", []string{"in(name,", "in(legal_name,", "in(registration_id,", "in(parent_id,", "in(id,", "in(updated_at,", "in(created_at,"}},
		{"Full binary operator match with full attribute and comma has no completions", "eq(name,", []string{}},
		{"Full vararg operator match with full attribute and comma has no completions", "in(name,", []string{}},
		{"Open double quoted string is closed with binary operator", `eq(status,"paid`, []string{`eq(status,"paid")`}},
		{"Open single quoted string is closed with binary operator", `eq(status,'paid`, []string{`eq(status,'paid')`}},
		{"Open double quoted string is closed with vararg operator", `in(status,"paid`, []string{`in(status,"paid")`, `in(status,"paid",`}},
		{"Open single quoted string is closed with vararg operator", `in(status,'paid`, []string{`in(status,'paid')`, `in(status,'paid',`}},
		{"Finished double quoted string is closed with binary operator", `eq(status,"paid"`, []string{`eq(status,"paid")`}},
		{"Finished single quoted string is closed with binary operator", `eq(status,'paid'`, []string{`eq(status,'paid')`}},
		{"Finished double quoted string is closed with vararg operator", `in(status,"paid"`, []string{`in(status,"paid")`, `in(status,"paid",`}},
		{"Finished single quoted string is closed with vararg operator", `in(status,'paid'`, []string{`in(status,'paid')`, `in(status,'paid',`}},
		{"Finished raw string is closed with binary operator", `eq(status,paid`, []string{`eq(status,paid)`}},
		{"Finished raw string is closed with vararg operator", `in(status,paid`, []string{`in(status,paid)`, `in(status,paid,`}},
		{"Open double quoted string is closed with vararg operator as 3rd arg", `in(status,"incomplete","paid`, []string{`in(status,"incomplete","paid")`, `in(status,"incomplete","paid",`}},
		{"Open single quoted string is closed with vararg operator as 3rd arg", `in(status,"incomplete",'paid`, []string{`in(status,"incomplete",'paid')`, `in(status,"incomplete",'paid',`}},
		{"Finished double quoted string is closed with vararg operator as 3rd arg", `in(status,'incomplete',"paid"`, []string{`in(status,'incomplete',"paid")`, `in(status,'incomplete',"paid",`}},
		{"Finished single quoted string is closed with vararg operator as 3rd arg", `in(status,'incomplete','paid'`, []string{`in(status,'incomplete','paid')`, `in(status,'incomplete','paid',`}},
		{"Closed operator auto completes with chain", "eq(status,paid)", []string{"eq(status,paid):"}},
		{"Chain operator completes with arguments", "eq(status,paid):", []string{"eq(status,paid):eq(", "eq(status,paid):lt(", "eq(status,paid):gt(", "eq(status,paid):ge(", "eq(status,paid):le(", "eq(status,paid):in(", "eq(status,paid):like("}},
		{"Chain operator completes with arguments with partial match", "eq(status,paid):e", []string{"eq(status,paid):eq(", "eq(status,paid):lt(", "eq(status,paid):gt(", "eq(status,paid):ge(", "eq(status,paid):le(", "eq(status,paid):in(", "eq(status,paid):like("}},
		{"Longer Smoke Test", "eq(status,paid):like(name,'foo*'):in(", []string{
			"eq(status,paid):like(name,'foo*'):in(name,",
			"eq(status,paid):like(name,'foo*'):in(id,",
			"eq(status,paid):like(name,'foo*'):in(created_at,",
			"eq(status,paid):like(name,'foo*'):in(updated_at,",
			"eq(status,paid):like(name,'foo*'):in(parent_id,",
			"eq(status,paid):like(name,'foo*'):in(legal_name,",
			"eq(status,paid):like(name,'foo*'):in(registration_id,"}},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			// Fixture Setup
			filter := tc.filter
			resource, _ := resources.GetResourceByName("account")

			expectedCompletions := tc.expectedCompletions
			// Execute SUT
			completions := GetFilterCompletion(filter, resource)

			// Verification
			require.ElementsMatch(t, expectedCompletions, completions, "Elements should have matched")
		})
	}

}

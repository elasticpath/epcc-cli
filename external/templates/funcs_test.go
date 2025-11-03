package templates

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNRandIntReturnsEmptyArrayWhenNNegative(t *testing.T) {

	// Fixture Setup
	n := -1
	minInt := 0
	maxInt := 100

	// Execute SUT
	res := NRandInt(n, minInt, maxInt)

	// Verification
	require.Empty(t, res)
}

func TestNRandIntReturnsEmptyArrayWhenNIsBiggerThanRange(t *testing.T) {

	// Fixture Setup
	n := 200
	minInt := 0
	maxInt := 100

	// Execute SUT
	res := NRandInt(n, minInt, maxInt)

	// Verification
	require.Empty(t, res)
}

func TestNRandIntReturnsEmptyArrayWhenIntervalIsEmpty(t *testing.T) {

	// Fixture Setup
	n := 1
	minInt := 0
	maxInt := 0

	// Execute SUT
	res := NRandInt(n, minInt, maxInt)

	// Verification
	require.Empty(t, res)
}

func TestNRandIntReturnsValidValueInSimpleCase(t *testing.T) {

	// Fixture Setup
	n := 10
	minInt := 0
	maxInt := 100

	// Execute SUT
	res := NRandInt(n, minInt, maxInt)

	// Verification
	validateAllUnique(t, res)
	require.Equal(t, n, len(res))
	validateAllInRange(t, res, minInt, maxInt)
}

func TestNRandIntReturnsValidValueInSimpleCaseWithNonZeroMin(t *testing.T) {

	// Fixture Setup
	n := 10
	minInt := 100
	maxInt := 200

	// Execute SUT
	res := NRandInt(n, minInt, maxInt)

	// Verification
	validateAllUnique(t, res)
	require.Equal(t, n, len(res))
	validateAllInRange(t, res, minInt, maxInt)
}

func TestNRandIntReturnsDifferentValuesWhenCalledTwice(t *testing.T) {

	// Fixture Setup
	n := 10
	minInt := 100
	maxInt := 200

	// Execute SUT
	res := NRandInt(n, minInt, maxInt)
	res2 := NRandInt(n, minInt, maxInt)

	// Verification
	validateAllUnique(t, res)
	require.Equal(t, n, len(res))
	validateAllInRange(t, res, minInt, maxInt)

	validateAllUnique(t, res2)
	require.Equal(t, n, len(res2))
	validateAllInRange(t, res2, minInt, maxInt)

	require.NotEqual(t, res, res2)
}

func TestNRandIntReturnsSameValuesWhenSeeded(t *testing.T) {

	// Fixture Setup
	n := 10
	minInt := 100
	maxInt := 200
	seed := time.Now().Unix() % 86400

	// Execute SUT
	Seed(seed)
	res := NRandInt(n, minInt, maxInt)
	Seed(seed)
	res2 := NRandInt(n, minInt, maxInt)

	// Verification
	validateAllUnique(t, res)
	require.Equal(t, n, len(res))
	validateAllInRange(t, res, minInt, maxInt)

	validateAllUnique(t, res2)
	require.Equal(t, n, len(res2))
	validateAllInRange(t, res2, minInt, maxInt)

	require.Equal(t, res, res2)
}

func TestNRandIntReturnsDifferentValuesWhenCalledTwiceWithLargeValues(t *testing.T) {

	// Fixture Setup
	n := 4096
	minInt := 1
	maxInt := 65535

	// Execute SUT
	res := NRandInt(n, minInt, maxInt)
	res2 := NRandInt(n, minInt, maxInt)

	// Verification
	validateAllUnique(t, res)
	require.Equal(t, n, len(res))
	validateAllInRange(t, res, minInt, maxInt)

	validateAllUnique(t, res2)
	require.Equal(t, n, len(res2))
	validateAllInRange(t, res2, minInt, maxInt)

	require.NotEqual(t, res, res2)
}

func validateAllUnique(t *testing.T, values []int) {
	m := map[int]struct{}{}

	for _, v := range values {
		m[v] = struct{}{}
	}

	require.Equal(t, len(m), len(values), "All values should be unique")
}

func validateAllInRange(t *testing.T, values []int, minInt, maxInt int) {
	// Check that all values are within the expected range
	for _, v := range values {
		require.GreaterOrEqual(t, v, minInt, "Value should be >= minInt")
		require.Less(t, v, maxInt, "Value should be < maxInt")
	}
}

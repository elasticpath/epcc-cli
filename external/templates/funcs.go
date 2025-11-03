package templates

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/elasticpath/epcc-cli/external/faker"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

// toFloat64 converts 64-bit floats
func toFloat64(v any) float64 {
	return cast.ToFloat64(v)
}

func toInt(v any) int {
	return cast.ToInt(v)
}

// toInt64 converts integer types to 64-bit integers
func toInt64(v any) int64 {
	return cast.ToInt64(v)
}

// RandString is the internal function that generates a random string. It takes the length of the string and a string of allowed characters as parameters.
func RandString(letters string, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

// RandAlphaNum generates a string consisting of characters in the range 0-9, a-z, and A-Z.
func RandAlphaNum(n int) string {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return RandString(letters, n)
}

// RandAlpha generates a string consisting of characters in the range a-z and A-Z.
func RandAlpha(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return RandString(letters, n)
}

// RandNumeric generates a string consisting of characters in the range 0-9.
func RandNumeric(n int) string {
	const digits = "0123456789"
	return RandString(digits, n)
}

// RandInt returns a value between the min (inclusive) and max (exclusive)
func RandInt(minA, maxA any) int {

	minN := toInt(minA)
	maxN := toInt(maxA)

	return r.Intn(maxN-minN) + minN
}

// RandNorm returns a normal sample with set mean and std deviation.
func RandNorm(meanA any, stdDevA any) float64 {
	mean := toFloat64(meanA)
	stdDev := toFloat64(stdDevA)

	return r.NormFloat64()*stdDev + mean
}

var mutex sync.Mutex

var sampler = make(map[string][]time.Time)

func Fake(string string) string {
	return faker.CallFakeFunc(string)
}

func Seed(x any) string {
	n := toInt64(x)

	faker.Seed(n)
	r = rand.New(rand.NewSource(n))
	return ""
}

// NRandInt returns n unique values between the min (inclusive) and max (exclusive)
func NRandInt(nAny, minAny, maxAny any) []int {

	n := toInt(nAny)
	minInt := toInt(minAny)
	maxInt := toInt(maxAny)

	if n > (maxInt - minInt) {
		return []int{}
	} else if n < 0 {
		return []int{}
	}

	// If the range is small enough, lets just generate a permutation
	if n < 2048 && (maxInt-minInt) < 4096 {
		nRange := maxInt - minInt
		perm := r.Perm(nRange)

		v := make([]int, n)

		for i := 0; i < min(n, nRange); i++ {
			v[i] = perm[i] + minInt
		}
		return v
	} else {
		// Otherwise, lets just generate a set of random numbers
		// This algorithm uses rejection sampling and is maybe terrible when n~=(max-min)
		v := map[int]struct{}{}

		for len(v) < n {
			v[RandInt(minInt, maxInt)] = struct{}{}
		}

		results := make([]int, 0, n)
		for k := range v {
			results = append(results, k)
		}

		sort.Ints(results)
		return results
	}
}

func FormatPrice(currency string, pAny any) string {

	p := toInt64(pAny)

	symbol := "£"
	if currency == "USD" {
		symbol = "$"
	}

	return fmt.Sprintf("%s%d.%02d", symbol, p/100, p%100)
}

func WeightedDateTimeSampler(start string, end string) string {

	key := fmt.Sprintf("%s_%s", start, end)
	mutex.Lock()
	defer mutex.Unlock()

	if computedTable, ok := sampler[key]; ok {
		t := computedTable[r.Intn(len(computedTable))]

		t = t.Add(time.Duration(RandInt(0, 3600)) * time.Second)
		return t.Format(time.RFC3339)
	}

	startTime, err := parseAny(start)
	if err != nil {
		log.Fatalf("failed to parse start time: %v", err)
	}

	endTime, err := parseAny(end)
	if err != nil {
		log.Fatalf("failed to parse start time: %v", err)
	}

	if !startTime.Before(endTime) {
		log.Fatalf("start %s must be before end, %s", startTime, endTime)
	}

	hours := endTime.Sub(startTime).Hours()

	weightedHours := make([]int, int(hours))

	// Actually compute the day of the week
	startDayOfWeek := 0.0

	for idx := range weightedHours {
		// Each hour has uniform weight

		hour := float64(idx % 24)
		day := float64((idx / 24) % 7)
		hourlyBias := 3 * math.Cos(2*math.Pi*((hour+4)/24))               // daily cycle
		dayOfWeekBias := 2 * math.Cos(2*math.Pi*((day+startDayOfWeek)/7)) // weekly cycle

		//hourlyBias := ((idx % 24) / 2)
		//dayOfWeekBias := (idx/24)%7 + startDayOfWeek
		weekBias := idx * 4 / int(hours)

		weightedHours[idx] = int(math.Max(hourlyBias+dayOfWeekBias+float64(weekBias), 1))
	}

	sum := 0

	lookup := make([]time.Time, 0)

	currentBlock := startTime
	for _, v := range weightedHours {
		sum += v
		for j := 0; j < v; j++ {
			lookup = append(lookup, currentBlock)
		}

		currentBlock = currentBlock.Add(time.Duration(1) * time.Hour)
	}

	sampler[key] = lookup

	t := lookup[r.Intn(len(lookup))]
	t = t.Add(time.Duration(RandInt(0, 3600)) * time.Second)
	return t.Format(time.RFC3339)
}

func parseAny(s string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("unsupported time format: %q", s)
}

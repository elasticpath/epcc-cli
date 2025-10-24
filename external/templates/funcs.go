package templates

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// randString is the internal function that generates a random string.
// It takes the length of the string and a string of allowed characters as parameters.
func RandString(letters string, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// randAlphaNum generates a string consisting of characters in the range 0-9, a-z, and A-Z.
func RandAlphaNum(n int) string {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return RandString(letters, n)
}

// randAlpha generates a string consisting of characters in the range a-z and A-Z.
func RandAlpha(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return RandString(letters, n)
}

// randNumeric generates a string consisting of characters in the range 0-9.
func RandNumeric(n int) string {
	const digits = "0123456789"
	return RandString(digits, n)
}

func RandInt(min, max int) int {
	return rand.Intn(max-min) + min
}

var mutex sync.Mutex

var sampler = make(map[string][]time.Time)

func WeightedDateTimeSampler(start string, end string) string {

	key := fmt.Sprintf("%s_%s", start, end)
	mutex.Lock()
	defer mutex.Unlock()

	if computedTable, ok := sampler[key]; ok {
		t := computedTable[rand.Intn(len(computedTable))]

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
	for i, v := range weightedHours {

		sum += v

		fmt.Printf("%d %d\n", i, v)
		for j := 0; j < v; j++ {
			lookup = append(lookup, currentBlock)
		}

		currentBlock = currentBlock.Add(time.Duration(1) * time.Hour)
	}

	//fmt.Printf("Sum: %d", sum)

	sampler[key] = lookup

	t := lookup[rand.Intn(len(lookup))]
	t = t.Add(time.Duration(RandInt(0, 3600)) * time.Second)
	return t.Format(time.RFC3339)
}

func parseAny(s string) (time.Time, error) {
	// Try a few common layouts; add more if you need.
	//if t, err := time.Parse(time.RFC3339, s); err == nil {
	//	return t, nil
	//}
	//if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
	//	return t, nil
	//}
	//if t, err := time.Parse("2006-01-02 15:04", s); err == nil {
	//	return t, nil
	//}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("unsupported time format: %q", s)
}

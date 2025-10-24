package templates

import (
	"sort"
	"testing"
)

func TestBiasedDateSampler(t *testing.T) {

	times := []string{}
	for i := 0; i < 1000; i++ {
		times = append(times, WeightedDateTimeSampler("2025-07-01", "2025-09-30"))
	}

	sort.Strings(times)

	for i := 0; i < len(times); i++ {
		//fmt.Println(times[i])
	}

}

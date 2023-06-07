package cmd

import (
	"testing"
)

func BenchmarkCmdCompletion(b *testing.B) {
	for n := 0; n < b.N; n++ {
		foo := generateRunbookCmd()

		if foo == nil {
			return
		}
	}

}

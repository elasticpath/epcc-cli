package cmd

import (
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/elasticpath/epcc-cli/external/runbooks"
	"testing"
)

func BenchmarkRootCommandBootstrap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runbooks.Reset()
		RootCmd = GetRootCommand()
		resources.PublicInit()
		InitializeCmd()
	}
}

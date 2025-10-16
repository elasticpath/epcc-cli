package resources__test

import (
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/resources"
)

func init() {
	aliases.InitializeAliasDirectoryForTesting()
	resources.PublicInit()
}

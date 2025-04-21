package resources

import "github.com/elasticpath/epcc-cli/external/aliases"

func init() {
	aliases.InitializeAliasDirectoryForTesting()
	PublicInit()
}

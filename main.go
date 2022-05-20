package main

import (
	"github.com/elasticpath/epcc-cli/cmd"
	"github.com/elasticpath/epcc-cli/external/httpclient"
)

func main() {
	cmd.Execute()
	httpclient.LogStats()
}

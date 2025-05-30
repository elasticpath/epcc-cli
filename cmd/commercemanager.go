package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/browser"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
)

var ApiHostToUrlMap = map[string]string{
	"api.moltin.com":             "https://euwest.cm.elasticpath.com/",
	"euwest.api.elasticpath.com": "https://euwest.cm.elasticpath.com/",
	"useast.api.elasticpath.com": "https://useast.cm.elasticpath.com/",
}
var cmCommand = &cobra.Command{
	Use:   "commerce-manager",
	Short: "Open commerce manager",
	RunE: func(cmd *cobra.Command, args []string) error {

		env := config.GetEnv()
		u, err := url.Parse(env.EPCC_API_BASE_URL)
		if err != nil {
			fmt.Println(err)
			return err
		}

		cmUrl, ok := ApiHostToUrlMap[u.Host]

		if !ok {
			return fmt.Errorf("Don't know where Commerce Manager is for $EPCC_API_BASE_URL=%s \n", u)
		}

		err = browser.OpenUrl(cmUrl)
		if err != nil {
			return err
		}

		log.Tracef("Opening browser to %s", cmUrl)

		return nil
	},
}

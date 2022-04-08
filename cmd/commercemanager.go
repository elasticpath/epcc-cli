package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
)

var cmCommand = &cobra.Command{
	Use:   "commerce-manager",
	Short: "Open commerce manager",
	RunE: func(cmd *cobra.Command, args []string) error {
		logToDisk("POST", "/v2/accounts", nil, nil, 200)
		u, err := url.Parse(config.Envs.EPCC_API_BASE_URL)
		if err != nil {
			fmt.Println(err)
			return err
		}
		var cmUrl string
		switch u.Host {
		case "api.moltin.com":
			cmUrl = "https://euwest.cm.elasticpath.com/"
		case "useast.api.elasticpath.com":
			cmUrl = "https://useast.cm.elasticpath.com/"
		}

		if cmUrl == "" {
			return fmt.Errorf("Don't know where Commerce Manager is for $EPCC_API_BASE_URL=%s \n", u)
		}
		err = OpenUrl(cmUrl)
		if err != nil {
			return err
		}

		log.Tracef("Opening browser to %s", cmUrl)

		return nil
	},
}

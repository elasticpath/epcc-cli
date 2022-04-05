package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/url"
)

var cmCommand = &cobra.Command{
	Use:   "commerce-manager",
	Short: "Open commerce manager",
	RunE: func(cmd *cobra.Command, args []string) error {
		u, err := url.Parse(Envs.EPCC_API_BASE_URL)
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
			fmt.Printf("Don't know where Commerce Manager is for $EPCC_API_BASE_URL=%s \n", u)
			return err
		}
		err = OpenUrl(cmUrl)
		if err != nil {
			fmt.Println(err)
			return err
		}

		fmt.Printf("Opening browser to %s", cmUrl)

		return nil
	},
}

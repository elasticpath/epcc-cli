package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
	"os/exec"
	"runtime"
)

var cmCommand = &cobra.Command{
	Use:   "commerce-manager",
	Short: "Open commerce manager",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		switch runtime.GOOS {
		case "linux":
			err = exec.Command("xdg-open", cmUrl).Start()
		case "windows":
			err = exec.Command("rundll32", "url.dll,FileProtocolHandler", cmUrl).Start()
		case "darwin":
			err = exec.Command("open", cmUrl).Start()
		default:
			err = fmt.Errorf("unsupported platform")
		}

		if err != nil {
			return err
		}

		log.Trace("Opening browser to %s", cmUrl)

		return nil
	},
}

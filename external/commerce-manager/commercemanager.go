package commercemanager

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/command"
	"net/url"
	"os/exec"
	"runtime"
)

var Command = command.Command{
	Keyword:     "commerce-manager",
	Description: "Open commerce manager",
	Execute: func(cmds map[string]command.Command, cmd string, args []string, envs config.Env) int {
		u, err := url.Parse(envs.EPCC_API_BASE_URL)
		if err != nil {
			fmt.Println(err)
			return 1
		}
		var cmUrl string
		switch u.Host {
		case "api.moltin.com":
			cmUrl = "https://euwest.cm.elasticpath.com/"
		case "useast.api.elasticpath.com":
			cmUrl = "https://useast.cm.elasticpath.com/"
		}

		if cmUrl == "" {
			fmt.Printf("Don't know where Commerce Manager is for $EPCC_API_BASE_URL=%s \n", envs.EPCC_API_BASE_URL)
			return 1
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
			fmt.Println(err)
			return 1
		}

		return 0
	},
}

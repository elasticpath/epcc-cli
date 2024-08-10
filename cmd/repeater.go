package cmd

import (
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/shutdown"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"
)

func repeater(c func(*cobra.Command, []string) error, repeat, repeatDelay uint32, cmd *cobra.Command, args []string, ignoreErrors bool) error {
	for i := 0; i < int(repeat); i++ {
		err := c(cmd, args)

		if err != nil {

			if ignoreErrors {
				log.Debugf("Ignored error %v", ignoreErrors)
			} else {
				if repeat > 1 && !ignoreErrors && httpclient.RetryAllErrors {
					log.Infof("If you want to continue repeating even if the requests gets a 4xx you should use `--ignore-errors.`")
				}
				return err
			}
		}

		if i < int(repeat)-1 {
			time.Sleep(time.Duration(repeatDelay) * time.Millisecond)
		}

		if shutdown.ShutdownFlag.Load() {
			return nil
		}
	}

	return nil
}

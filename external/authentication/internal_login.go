package authentication

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/config"
	log "github.com/sirupsen/logrus"
	"net/url"
	"time"
)

func InternalImplicitAuthentication(args []string) error {
	values := url.Values{}
	values.Set("grant_type", "implicit")

	env := config.GetEnv()
	if len(args) == 0 {
		log.Debug("Arguments have been passed, not using profile EPCC_CLIENT_ID")
		values.Set("client_id", env.EPCC_CLIENT_ID)
	}

	if len(args)%2 != 0 {
		return fmt.Errorf("invalid number of arguments supplied to login command, must be multiple of 2, not %v", len(args))
	}

	for i := 0; i < len(args); i += 2 {
		k := args[i]
		values.Set(k, args[i+1])
	}

	token, err := GetAuthenticationToken(false, &values, true)

	if err != nil {
		return err
	}

	if token != nil {
		log.Infof("Successfully authenticated with implicit token, session expires %s", time.Unix(token.Expires, 0).Format(time.RFC1123Z))
	} else {
		log.Warn("Did not successfully authenticate against the API")
	}

	return nil
}

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var get = &cobra.Command{
	Use:   "get [resource]",
	Short: "Retrieves a single resource.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := &http.Client{
			Timeout: time.Second * 10,
		}
		req, err := http.NewRequest("GET", Envs.EPCC_API_BASE_URL+"/v2/"+args[0], nil)
		if err != nil {
			return fmt.Errorf("Got error %s", err.Error())
		}
		req.Header.Set("user-agent", "golang application")
		req.Header.Set("Authorization", "Bearer: Md8e3f817e7975c8d3e81ba2dd3b242472d1c91d6")
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("Got error %s", err.Error())
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 && resp.StatusCode <= 600 {
			return fmt.Errorf(resp.Status)
		}

		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(body))

		return nil
	},
}

package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/profiles"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
	"strconv"
	"strings"
)

// LogsCmd represents the logs command
var LogsClear = &cobra.Command{
	Use:   "clear",
	Short: "Clears all HTTP request and response logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		return profiles.ClearAllRequestLogs()
	},
}

var LogsList = &cobra.Command{
	Use:   "list",
	Short: "List all HTTP logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		files, err := profiles.GetAllRequestLogTitles()
		if err != nil {
			return err
		}

		for idx, name := range files {
			fmt.Printf("%d %s\n", idx, name)
		}
		return nil
	},
}

var LogsShow = &cobra.Command{
	Use:   "show <NUMBER>",
	Short: "Show HTTP logs for specific number, negative values are from the last value",
	Args:  cobra.MinimumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {

		i, err := strconv.Atoi(args[0])

		if err != nil {
			return fmt.Errorf("could not get the %s entry => %w", args[0], err)
		}

		content, err := profiles.GetNthRequestLog(i)

		if err != nil {
			return fmt.Errorf("couldn't print logs: %v", err)
		}

		fmt.Println(content)

		return nil
	},
}

var CurlInlineAuth = false

var LogsCurlReplay = &cobra.Command{
	Use:   "curl-replay <NUMBER>",
	Short: "Generate a curl command that replays the request from a specific log entry",
	Args:  cobra.MinimumNArgs(1),

	// Allows running `epcc logs curl-replay -1` without error
	//DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		i, err := strconv.Atoi(args[0])

		if err != nil {
			return fmt.Errorf("could not get the %s entry => %w", args[0], err)
		}

		content, err := profiles.GetNthRequestLog(i)

		if err != nil {
			return fmt.Errorf("couldn't get log: %v", err)
		}

		// Split the content by lines
		lines := strings.Split(content, "\n")
		if len(lines) == 0 {
			return fmt.Errorf("log entry has no content")
		}

		// Strip out stray whitespace at the end of the line (e.g., \r)
		for i := range lines {
			lines[i] = strings.TrimSpace(lines[i])
		}

		// First line should be the request line (METHOD PATH HTTP/VERSION)
		requestLine := lines[0]
		parts := strings.Fields(requestLine)
		if len(parts) < 2 {
			return fmt.Errorf("invalid request line: %s", requestLine)
		}

		method := parts[0]
		path := parts[1]

		// Extract headers - they appear after the request line until an empty line
		var host string
		var headers []string
		lineIndex := 1

		// Process headers
		for ; lineIndex < len(lines); lineIndex++ {
			line := lines[lineIndex]

			// Empty line marks the end of headers
			if line == "" {
				break
			}

			// Process each header
			if strings.Contains(line, ": ") {
				headerParts := strings.SplitN(line, ": ", 2)
				if len(headerParts) == 2 {
					name := headerParts[0]
					value := headerParts[1]

					if name == "Host" {
						host = value
					}

					// Skip headers that curl adds automatically or might cause issues
					if name != "Content-Length" && name != "Accept-Encoding" {
						headers = append(headers, fmt.Sprintf("%s: %s", name, value))
					}

				}
			}
		}

		if host == "" {
			return fmt.Errorf("could not find Host header in request")
		}

		env := config.GetEnv()

		var template = "https://%s%s"

		var authUrl = ""

		reqURL, err := url.Parse(env.EPCC_API_BASE_URL)
		if err != nil {
			log.Debugf("Could not get base url defaulting to https")
			authUrl = fmt.Sprintf("https://%s/oauth/access_token", host)
		} else if reqURL.Scheme != "" {
			// TODO maybe handle other ports
			template = reqURL.Scheme + "://%s%s"
			authUrl = fmt.Sprintf("%s://%s/oauth/access_token", reqURL.Scheme, host)
		} else {
			authUrl = fmt.Sprintf("https://%s/oauth/access_token", host)
		}

		// Build the full URL
		url := fmt.Sprintf(template, host, path)

		// Look for request body - it would be after the headers and before the response
		var body string
		bodyStartIndex := lineIndex + 1
		responseStartIndex := -1

		// Find where the response starts (with HTTP/)
		for j := bodyStartIndex; j < len(lines); j++ {
			if strings.HasPrefix(lines[j], "HTTP/") {
				responseStartIndex = j
				break
			}
		}

		// Extract body if present
		if responseStartIndex > bodyStartIndex {
			bodyLines := lines[bodyStartIndex:responseStartIndex] // -1 to skip the empty line before response
			if len(bodyLines) > 0 {
				body = strings.Join(bodyLines, "\n")
			}
		}

		// Build the curl command
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("curl -X %s \\\n  '%s'", method, url))

		// Add headers
		for _, header := range headers {

			if strings.HasPrefix(header, "Authorization") && CurlInlineAuth {
				if env.EPCC_CLIENT_SECRET != "" {
					sb.WriteString(fmt.Sprintf(" \\\n  -H \"Authorization: Bearer $(curl -s -X POST '%s' -d 'client_id=%s' -d 'client_secret=%s' -d 'grant_type=client_credentials' | jq -r .access_token)\"", authUrl, env.EPCC_CLIENT_ID, env.EPCC_CLIENT_SECRET))
				} else {
					sb.WriteString(fmt.Sprintf(" \\\n  -H \"Authorization: Bearer $(curl -s -X POST '%s' -d 'client_id=%s' -d 'grant_type=implicit' | jq -r .access_token)\"", authUrl, env.EPCC_CLIENT_ID))
				}
			} else {
				sb.WriteString(fmt.Sprintf(" \\\n  -H '%s'", header))
			}

		}

		// Add body if present
		if strings.TrimSpace(body) != "" {
			// Escape single quotes in the body
			escapedBody := strings.ReplaceAll(body, "'", "'\\''")
			sb.WriteString(fmt.Sprintf(" \\\n  -d '%s'", escapedBody))
		}

		fmt.Println(sb.String())
		return nil
	},
}

var Logs = &cobra.Command{Use: "logs", Short: "Retrieve information about previous requests"}

package oidc

import (
	"bytes"
	"embed"
	"fmt"

	"github.com/elasticpath/epcc-cli/external/browser"
	"github.com/elasticpath/epcc-cli/external/clictx"

	"path/filepath"
	"time"

	"net/http"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
)

//go:embed site/*
var EmbedFS embed.FS

func StartOIDCServer(port uint16) error {

	// Parse all templates at startup
	var err error
	templates, err := template.ParseFS(EmbedFS, "site/*.gohtml")
	if err != nil {
		panic(fmt.Sprintf("Error parsing templates: %v", err))
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		ctx := clictx.Ctx
		log.Tracef("Handling request %s %s", r.Method, r.URL.Path)

		// Remove leading "/" and sanitize the path
		requestPath := strings.TrimPrefix(r.URL.Path, "/")
		if requestPath == "" {
			requestPath = "index"
		}

		// Check if it's a .html request for template rendering
		if filepath.Ext(requestPath) == "" {

			templateName := requestPath + ".gohtml"

			var data any
			var error error

			switch requestPath {
			case "index":
				data, error = GetIndexData(ctx, port)
			case "callback":
				data, error = GetCallbackData(ctx, port, r)
			case "get_token":
				data, error = GetTokenData(ctx, port, r)
			}

			if error != nil {
				log.Errorf("Request %s %s => %s", r.Method, r.URL.Path, "500")
				http.Error(w, fmt.Sprintf("Error getting data: %v", error), http.StatusInternalServerError)
				return
			}

			// Render the template by its base name
			if tmpl := templates.Lookup(templateName); tmpl != nil {
				if err := tmpl.Execute(w, data); err != nil {
					log.Errorf("Request %s %s => %s", r.Method, r.URL.Path, "500")
					http.Error(w, fmt.Sprintf("Error rendering template: %v", err), http.StatusInternalServerError)
				}

				log.Warnf("Request %s %s => %s", r.Method, r.URL.Path, "200")
				return
			}

			log.Infof("Request %s %s => %s", r.Method, r.URL.Path, "404")
			http.NotFound(w, r)
			return
		}

		file, err := EmbedFS.ReadFile("site/" + requestPath)
		if err != nil {
			log.Warnf("Request %s %s => %s", r.Method, r.URL.Path, "404")
			http.NotFound(w, r)
			return
		}

		if filepath.Ext(requestPath) == ".css" {
			w.Header().Set("Content-Type", "text/css")
		}

		http.ServeContent(w, r, requestPath, time.Now(), bytes.NewReader(file))
		log.Infof("Request %s %s => %s", r.Method, r.URL.Path, "200")
	})

	log.Infof("Starting server on port %d", port)

	go func() {
		log.Warnf("Waiting for server to start")
		time.Sleep(1 * time.Second)
		browser.OpenUrl(fmt.Sprintf("http://localhost:%d", port))

	}()

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	return err

}

type OidcProfileInfo struct {
	Name              string `mapstructure:"name"`
	AuthorizationLink string `mapstructure:"authorization_link"`
	Idp               string
}

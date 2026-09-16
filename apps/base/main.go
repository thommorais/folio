// Command base is the folio backend: a PocketBase instance that also serves
// the folio REST API. PocketBase owns auth, the admin UI and storage; folio
// adds its collections and the /api/folio routes driven by folio-core.
package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/router"

	folio "folio/folio-core"
	"folio/folio-core/adapters/httpapi"
)

func main() {
	app := pocketbase.New()

	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: isGoRun,
	})

	app.RootCmd.AddCommand(newSeedCommand(app))

	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		if err := trustProxyIPHeader(e.App); err != nil {
			return err
		}
		return folio.Migrate(e.App)
	})

	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		useCases := folio.New(e.App, nil)
		httpapi.New(useCases.Deps()).Mount(e)

		// Serve the built SPA when a directory is configured. Registered last
		// so /api and /_ win, and with indexFallback so a deep link like
		// /folio/todos reaches the client router instead of 404ing.
		if dir := publicDir(); dir != "" {
			e.Router.GET("/{path...}", spa(os.DirFS(dir)))
		}

		return e.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

// Railway's edge IP varies per request, which breaks realtime subscriptions.
func trustProxyIPHeader(app core.App) error {
	settings := app.Settings()
	if len(settings.TrustedProxy.Headers) > 0 {
		return nil
	}
	settings.TrustedProxy.Headers = []string{"X-Forwarded-For"}
	settings.TrustedProxy.UseLeftmostIP = true
	return app.Save(settings)
}

// apis.Static would answer an unmatched /api path with 200 index.html.
func spa(fsys fs.FS) func(*core.RequestEvent) error {
	static := apis.Static(fsys, true)

	return func(e *core.RequestEvent) error {
		path := "/" + strings.TrimPrefix(e.Request.URL.Path, "/")
		for _, prefix := range []string{"/api/", "/_/"} {
			if strings.HasPrefix(path, prefix) {
				return router.NewNotFoundError("Missing or invalid route.", nil)
			}
		}
		return static(e)
	}
}

// publicDir returns the directory the static SPA is served from. In production
// the nina-spa build is copied next to the binary as pb_public; PB_PUBLIC_DIR
// overrides this (e.g. a mounted path on Railway).
func publicDir() string {
	if dir := os.Getenv("PB_PUBLIC_DIR"); dir != "" {
		return dir
	}
	// Under `go run`, os.Args[0] is a temp binary, so resolve relative to the
	// working directory instead of the (temp) binary location.
	dir := filepath.Join(filepath.Dir(os.Args[0]), "pb_public")
	if strings.HasPrefix(os.Args[0], os.TempDir()) {
		dir = "./pb_public"
	}

	// An absent directory means the API is being run without a frontend
	// build. Returning "" leaves the catch-all route unregistered, rather
	// than answering every unmatched path with a 404 from an empty FS.
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return ""
	}
	return dir
}

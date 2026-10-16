package demo_test

import (
	"context"
	"fmt"
	"github.com/Alps86/shop/internal/runtime"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"time"
)

func (deployment *deployment) prepareNative(data map[string]string) error {
	public, err := ephemeralAddress()
	if err != nil {
		return err
	}

	deployment.public = "http://" + public
	deployment.config.Listen, err = ephemeralAddress()
	if err != nil {
		return err
	}

	deployment.config.Frontend.BackendListen, err = ephemeralAddress()
	if err != nil {
		return err
	}

	return deployment.configureNative(data)
}

func (deployment *deployment) configureNative(data map[string]string) error {
	deployment.backend = "http://" + deployment.config.Frontend.BackendListen
	deployment.config.Frontend.PublicOrigin = deployment.public
	deployment.config.Frontend.ContentDirectory = deployment.directory + "/public"
	deployment.config.Frontend.CacheEndpoint = deployment.public
	deployment.config.Frontend.CacheKey = data["CacheKey"]
	deployment.config.Frontend.CSRFKey = data["CSRFKey"]
	if err := deployment.startSearchProxy(); err != nil {
		return err
	}

	app, err := runtime.Open(context.Background(), deployment.config)
	deployment.app = app
	if err != nil {
		return err
	}

	deployment.store, deployment.connection = app.Store, app.Client

	return deployment.startEngines(data)
}

func (deployment *deployment) startSearchProxy() error {
	target, err := url.Parse(deployment.config.SearchEndpoint)
	if err != nil {
		return err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	deployment.search = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		deployment.mutex.RLock()
		unavailable := deployment.unavailable
		deployment.mutex.RUnlock()
		if unavailable {
			http.Error(writer, "Suchdienst nicht verfügbar", http.StatusServiceUnavailable)
			return
		}

		proxy.ServeHTTP(writer, request)
	}))
	deployment.config.SearchEndpoint = deployment.search.URL

	return nil
}

func (deployment *deployment) startEngines(data map[string]string) error {
	if err := deployment.prepareContent(); err != nil {
		return err
	}

	if err := deployment.buildHugo(); err != nil {
		return err
	}

	if err := deployment.launchEngines(data); err != nil {
		return err
	}

	return deployment.startApp()
}

func (deployment *deployment) prepareContent() error {
	for _, name := range []string{"layouts", "static", "content"} {
		if err := os.CopyFS(deployment.directory+"/source/"+name, os.DirFS("../../web/"+name)); err != nil {
			return err
		}
	}

	data, err := os.ReadFile("../../web/hugo.toml")
	if err != nil {
		return err
	}

	// #nosec G703 -- Eigenes temporäres Testverzeichnis.
	if err := os.WriteFile(deployment.directory+"/source/hugo.toml", data, 0600); err != nil {
		return err
	}

	return deployment.app.ExportContent(context.Background(), deployment.directory+"/source/content/generated")
}

func (deployment *deployment) buildHugo() error {
	// #nosec G204 -- Explizit konfiguriertes Standardbinary ohne Shell.
	command := exec.Command(binary("SHOP_HUGO", "hugo"), "--source", deployment.directory+"/source", "--destination", deployment.directory+"/public")
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Hugo: %w: %s", err, output)
	}

	return nil
}

func (deployment *deployment) startApp() error {
	ctx, cancel := context.WithCancel(context.Background())
	deployment.cancel = cancel
	deployment.finished = make(chan error, 1)
	go func() {
		deployment.finished <- deployment.app.Run(ctx)
	}()

	return deployment.ready(30 * time.Second)
}

package http_test

import (
	"context"
	"fmt"
	shophttp "github.com/Alps86/shop/internal/adapter/http"
	"github.com/cucumber/godog"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

func (current *world) startComposition(table *godog.Table) error {
	directory, err := os.MkdirTemp("", "shop-composition-")
	if err != nil {
		return err
	}

	engine := &compositionEngine{directory: directory}
	current.engine = engine
	if err := current.engineOrigin(engine); err != nil {
		return err
	}

	return engine.start(read(table)[0])
}

func (engine *compositionEngine) start(fields map[string]string) error {
	if err := engine.buildHugo(); err != nil {
		return err
	}

	port, err := engine.startNginx(fields)
	if err != nil {
		return err
	}

	return engine.startVarnish(port, fields)
}

func (engine *compositionEngine) nginxConfig(port string, fields map[string]string) (string, error) {
	config, err := os.ReadFile("../../" + fields["Nginx"])
	if err != nil {
		return "", err
	}

	replacements := strings.NewReplacer("include mime.types;", "types { text/html html; text/css css; application/javascript js; }", "/tmp/shop-nginx.pid", engine.directory+"/nginx.pid", "/tmp/shop-nginx-", engine.directory+"/nginx-", "listen 8082;", "listen 127.0.0.1:"+port+";", "/srv/shop/web", engine.directory+"/public", "http://shop:8080", engine.origin.URL)
	path := engine.directory + "/nginx.conf"

	// #nosec G703 -- Pfad ausschließlich unter eigenem os.MkdirTemp-Verzeichnis.
	return path, os.WriteFile(path, []byte(replacements.Replace(string(config))), 0600)
}

func (engine *compositionEngine) startVarnish(nginxPort string, fields map[string]string) error {
	port, err := ephemeralPort()
	if err != nil {
		return err
	}

	path, err := engine.varnishConfig(nginxPort, fields)
	if err != nil {
		return err
	}

	if err := engine.launchVarnish(port, path); err != nil {
		return err
	}

	engine.url = "http://127.0.0.1:" + port

	return engine.ready()
}

func (engine *compositionEngine) ready() error {
	client := &http.Client{Timeout: 100 * time.Millisecond}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		response, err := client.Get(engine.url + "/de-DE/")
		if err == nil {
			_, _ = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			if response.StatusCode == 200 {
				return nil
			}
		}

		time.Sleep(20 * time.Millisecond)
	}

	return fmt.Errorf("Standardengines nicht bereit: %s", engine.output.String())
}

func (current *world) composedRequest(table *godog.Table) error {
	fields := read(table)[0]
	request, err := http.NewRequest("GET", current.engine.url+fields["URL"], nil)
	if err != nil {
		return err
	}

	request.Header.Set("Cookie", fields["Cookie"])
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	defer func() {
		_ = response.Body.Close()
	}()

	return current.acceptResponse(response)
}

func (current *world) stopComposition(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
	if current.engine == nil {
		return ctx, err
	}

	if err != nil {
		err = fmt.Errorf("%w; engines: %s", err, current.engine.output.String())
	}

	current.engine.stop()

	return ctx, err
}

func (engine *compositionEngine) stop() {
	for _, command := range []*exec.Cmd{engine.varnish, engine.nginx} {
		if command == nil {
			continue
		}

		if command.Process != nil {
			_ = syscall.Kill(-command.Process.Pid, syscall.SIGTERM)
			_ = command.Wait()
		}
	}

	if engine.origin != nil {
		engine.origin.Close()
	}

	_ = os.RemoveAll(engine.directory)
}

func binary(variable, name string) string {
	if path := os.Getenv(variable); path != "" {
		return path
	}

	return name
}

func ephemeralPort() (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}

	defer func() {
		_ = listener.Close()
	}()
	_, port, err := net.SplitHostPort(listener.Addr().String())

	return port, err
}

func (output *engineLog) Write(data []byte) (int, error) {
	output.mutex.Lock()
	defer output.mutex.Unlock()

	return output.buffer.Write(data)
}

func (output *engineLog) String() string {
	output.mutex.Lock()
	defer output.mutex.Unlock()

	return output.buffer.String()
}

func (engine *compositionEngine) buildHugo() error {
	// #nosec G204 -- Explizit konfiguriertes lokales Standardbinary ohne Shell.
	command := exec.Command(binary("SHOP_HUGO", "hugo"), "--source", engine.source, "--destination", engine.directory+"/public")
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Hugo: %w: %s", err, output)
	}

	return nil
}

func (engine *compositionEngine) startNginx(fields map[string]string) (string, error) {
	port, err := ephemeralPort()
	if err != nil {
		return "", err
	}

	config, err := engine.nginxConfig(port, fields)
	if err != nil {
		return "", err
	}

	// #nosec G204 -- Explizit konfiguriertes lokales Standardbinary ohne Shell.
	engine.nginx = exec.Command(binary("SHOP_NGINX", "nginx"), "-c", config, "-g", "daemon off;")
	engine.nginx.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	engine.nginx.Stderr = &engine.output

	return port, engine.nginx.Start()
}

func (engine *compositionEngine) varnishConfig(port string, fields map[string]string) (string, error) {
	config, err := os.ReadFile("../../" + fields["Varnish"])
	if err != nil {
		return "", err
	}

	replacements := strings.NewReplacer(`"nginx"`, `"127.0.0.1"`, `"8082"`, `"`+port+`"`, `"shop"`, `"127.0.0.1"`, "SHOP_CACHE_KEY", fields["CacheKey"])
	path := engine.directory + "/default.vcl"
	// #nosec G703 -- Ausschließlich eigenes temporäres Testverzeichnis.
	return path, os.WriteFile(path, []byte(replacements.Replace(string(config))), 0600)
}

func (engine *compositionEngine) launchVarnish(port, path string) error {
	// #nosec G204 -- Explizit konfiguriertes lokales Standardbinary ohne Shell.
	engine.varnish = exec.Command(binary("SHOP_VARNISH", "varnishd"), "-F", "-j", "none", "-f", path, "-a", "127.0.0.1:"+port, "-n", engine.directory+"/varnish", "-s", "malloc,16m", "-p", "vsl_mask=-ReqHeader,-BereqHeader,-RespHeader,-BerespHeader,-ReqUnset,-BereqUnset,-RespUnset,-BerespUnset")
	engine.varnish.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	engine.varnish.Stderr = &engine.output

	return engine.varnish.Start()
}

func (current *world) acceptResponse(response *http.Response) error {
	current.response = httptest.NewRecorder()
	current.response.Code = response.StatusCode
	for name, values := range response.Header {
		current.response.Header()[name] = values
	}

	_, err := io.Copy(current.response.Body, response.Body)

	return err
}

func (current *world) engineOrigin(engine *compositionEngine) error {
	config := current.config()
	config.ContentDirectory = engine.directory + "/public"
	handler, err := shophttp.New(config)
	if err != nil {
		return err
	}

	engine.origin = httptest.NewServer(compositionOrigin{current: current, handler: handler})

	return engine.prepareSource(current.fixture)
}

package demo_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

func (deployment *deployment) launchEngines(data map[string]string) error {
	nginx, err := ephemeralAddress()
	if err != nil {
		return err
	}

	_, port, err := net.SplitHostPort(nginx)
	if err != nil {
		return err
	}

	if err := deployment.launchNginx(port); err != nil {
		return err
	}

	return deployment.launchVarnish(port, data)
}

func (deployment *deployment) launchNginx(port string) error {
	source, err := os.ReadFile("../../deploy/nginx/nginx.conf")
	if err != nil {
		return err
	}

	replace := strings.NewReplacer("include mime.types;", "types { text/html html; text/css css; application/javascript js; }", "/tmp/shop-nginx.pid", deployment.directory+"/nginx.pid", "/tmp/shop-nginx-", deployment.directory+"/nginx-", "listen 8082;", "listen 127.0.0.1:"+port+";", "/srv/shop/web", deployment.directory+"/public", "http://shop:8080", "http://"+deployment.config.Listen)
	path := deployment.directory + "/nginx.conf"
	// #nosec G703 -- Ausschließlich eigenes temporäres Testverzeichnis.
	if err := os.WriteFile(path, []byte(replace.Replace(string(source))), 0600); err != nil {
		return err
	}

	// #nosec G204 -- Konfiguriertes Standardbinary mit selbst erzeugter lokaler Konfiguration.
	deployment.nginx = exec.Command(binary("SHOP_NGINX", "nginx"), "-c", path, "-g", "daemon off;")

	return deployment.launch(deployment.nginx)
}

func (deployment *deployment) launchVarnish(port string, data map[string]string) error {
	path, err := deployment.varnishConfig(port, data)
	if err != nil {
		return err
	}

	address, err := url.Parse(deployment.public)
	if err != nil {
		return err
	}

	// #nosec G204 -- Konfiguriertes Standardbinary ohne Shell; keine Token im VSL.
	deployment.varnish = exec.Command(binary("SHOP_VARNISH", "varnishd"), "-F", "-j", "none", "-f", path, "-a", address.Host, "-n", deployment.directory+"/varnish", "-s", "malloc,16m", "-p", "vsl_mask=-ReqHeader,-BereqHeader,-RespHeader,-BerespHeader,-ReqUnset,-BereqUnset,-RespUnset,-BerespUnset")

	return deployment.launch(deployment.varnish)
}

func (deployment *deployment) varnishConfig(port string, data map[string]string) (string, error) {
	source, err := os.ReadFile("../../deploy/varnish/default.vcl")
	if err != nil {
		return "", err
	}

	replace := strings.NewReplacer(`"nginx"`, `"127.0.0.1"`, `"8082"`, `"`+port+`"`, `"shop"`, `"127.0.0.1"`, "SHOP_CACHE_KEY", data["CacheKey"])
	path := deployment.directory + "/default.vcl"
	// #nosec G703 -- Eigenes temporäres Testverzeichnis.
	return path, os.WriteFile(path, []byte(replace.Replace(string(source))), 0600)
}

func (deployment *deployment) launch(command *exec.Cmd) error {
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Stdout, command.Stderr = &deployment.output, &deployment.output

	return command.Start()
}

func (deployment *deployment) ready(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if deployment.readyOnce() {
			return nil
		}

		time.Sleep(50 * time.Millisecond)
	}

	return fmt.Errorf("Demolaufzeit nicht bereit: %s", deployment.output.String())
}

func (deployment *deployment) readyOnce() bool {
	for _, target := range []string{deployment.public + "/de-DE/", deployment.backend + "/backend"} {
		if !healthy(target) {
			return false
		}
	}

	return deployment.projectionReady()
}

func (deployment *deployment) projectionReady() bool {
	if deployment.app != nil {
		status, err := deployment.app.Snapshot(context.Background())
		return err == nil && status["CDC"] == "initialized" && status["OpenSearch"] == "ready"
	}

	client := &http.Client{Timeout: time.Second}
	response, err := client.Get(deployment.backend + "/backend")
	if err != nil {
		return false
	}

	body, err := io.ReadAll(response.Body)
	closeError := response.Body.Close()

	return err == nil && closeError == nil && strings.Contains(string(body), "<dt>CDC</dt><dd>initialized</dd>")
}

func healthy(target string) bool {
	client := &http.Client{Timeout: time.Second}
	response, err := client.Get(target)
	if err != nil {
		return false
	}

	_, copyError := io.Copy(io.Discard, response.Body)
	closeError := response.Body.Close()

	return response.StatusCode == http.StatusOK && copyError == nil && closeError == nil
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

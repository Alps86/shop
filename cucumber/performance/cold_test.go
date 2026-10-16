package performance_test

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func (current *world) cold(route, path, cookie string, data map[string]string) error {
	before, err := current.cacheCounters()
	if err != nil {
		return err
	}

	samples, err := current.curlSamples(route, path, cookie, data)
	if err != nil {
		return err
	}

	after, err := current.cacheCounters()
	if err != nil {
		return err
	}

	return current.record(route, "cold-cache-services-ready", samples, after.subtract(before))
}

func (current *world) curlSamples(route, path, cookie string, data map[string]string) (result []sample, err error) {
	count := current.expectedRequests

	directory, err := os.MkdirTemp("", "shop-cold-body-")
	if err != nil {
		return nil, err
	}

	defer func() {
		err = errors.Join(err, os.RemoveAll(directory))
	}()
	config := current.curlConfig(path, cookie, directory, count)
	command := current.curlCommand(config)
	output, err := command.Output()
	if err != nil {
		return nil, require(err, "Standardwerkzeug curl")
	}

	return current.readCold(route, directory, count, output)
}

func (current *world) curlConfig(path, cookie, directory string, count int) string {
	var config strings.Builder
	run := strings.ToLower(rand.Text())
	for index := range count {
		fmt.Fprintf(&config, "url = %q\nheader = %q\noutput = %q\nwrite-out = \"%%{json}\\n\"\nmax-time = 15\n", current.public+path, "Host: cold-"+run+"-"+strconv.Itoa(index)+".localhost", filepath.Join(directory, strconv.Itoa(index)))
		if cookie != "" {
			fmt.Fprintf(&config, "header = %q\n", "Cookie: "+cookie)
		}

		if index+1 < count {
			config.WriteString("next\n")
		}
	}

	return config.String()
}

func (current *world) readCold(route, directory string, count int, output []byte) ([]sample, error) {
	_, _, contains := current.route(route)
	samples, err := curlSamples(output)
	if err != nil {
		return nil, err
	}

	for index := range count {
		body, err := readBody(directory, strconv.Itoa(index))
		if err != nil {
			return nil, err
		}

		if err := validHTML(200, string(body), contains); err != nil {
			return nil, err
		}
	}

	err = os.WriteFile(current.reportPath(route, "cold", ".jsonl"), output, 0600)

	return samples, err
}

func curlSamples(output []byte) ([]sample, error) {
	result := []sample{}
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	for decoder.More() {
		var observed curlResult
		if err := decoder.Decode(&observed); err != nil {
			return nil, err
		}

		result = append(result, sample{seconds: observed.TimeTotal, status: observed.HTTPCode, connections: observed.NumConnects})
	}

	return result, nil
}

func readBody(directory, name string) ([]byte, error) {
	file, err := os.OpenInRoot(directory, name)
	if err != nil {
		return nil, err
	}

	body, readError := io.ReadAll(file)

	return body, errors.Join(readError, file.Close())
}

func (current *world) curlCommand(config string) *exec.Cmd {
	command := exec.Command("curl", "--disable", "--parallel", "--silent", "--show-error", "--config", "-")
	command.Args = append(command.Args, "--parallel-max", current.data["Concurrency"], "--noproxy", "*")
	configureTool(command, "SHOP_CURL")
	command.Stdin = strings.NewReader(config)

	return command
}

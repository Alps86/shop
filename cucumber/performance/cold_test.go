package performance_test

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
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

	return current.executeCold(coldBatch{route: route, path: path, cookie: cookie, directory: directory, count: count})
}

func (current *world) executeCold(batch coldBatch) ([]sample, error) {
	config := batch.config(current.public)
	command := current.curlCommand(config)
	batch.started = time.Now()
	output, commandError := command.Output()
	batch.ended = time.Now()
	samples, readError := current.readCold(batch, output)

	return samples, errors.Join(require(commandError, "Standardwerkzeug curl"), readError)
}

func (current *world) readCold(batch coldBatch, output []byte) ([]sample, error) {
	if err := current.saveColdMetrics(batch.route, output); err != nil {
		return nil, err
	}

	samples, err := curlSamples(output)
	if err != nil {
		return nil, err
	}

	return samples, current.verifyColdBodies(batch, output)
}

func (current *world) verifyColdBodies(batch coldBatch, output []byte) error {
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	_, _, contains := current.route(batch.route)
	for decoder.More() {
		var observed curlResult
		if err := decoder.Decode(&observed); err != nil {
			return err
		}

		if err := current.checkColdTransfer(batch, observed, contains); err != nil {
			return err
		}
	}

	return nil
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

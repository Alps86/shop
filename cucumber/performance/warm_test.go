package performance_test

import (
	"encoding/csv"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func (current *world) warm(route, path, cookie string, data map[string]string) error {
	if _, err := current.hey(path, cookie, data["WarmupPerRoute"]); err != nil {
		return err
	}

	return current.warmSamples(route, path, cookie)
}

func (current *world) warmSamples(route, path, cookie string) error {
	before, err := current.cacheCounters()
	if err != nil {
		return err
	}

	samples, err := current.heyResult(route, path, cookie)
	if err != nil {
		return err
	}

	after, err := current.cacheCounters()
	if err != nil {
		return err
	}

	return current.record(route, "warm", samples, after.subtract(before))
}

func (current *world) heyResult(route, path, cookie string) ([]sample, error) {
	output, err := current.hey(path, cookie, strconv.Itoa(current.expectedRequests))
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(current.reportPath(route, "warm", ".csv"), output, 0600); err != nil {
		return nil, err
	}

	return heySamples(string(output))
}

func (current *world) hey(path, cookie, count string) ([]byte, error) {
	arguments := []string{"-n", count, "-c", current.data["Concurrency"], "-o", "csv", "-disable-redirects", "-disable-compression", "-t", "15"}
	if cookie != "" {
		arguments = append(arguments, "-H", "Cookie: "+cookie)
	}

	arguments = append(arguments, current.public+path)
	command := exec.Command("hey")
	configureTool(command, "SHOP_HEY")
	command.Args = append(command.Args, arguments...)
	output, err := command.Output()

	return output, require(err, "Standardwerkzeug hey")
}

func heySamples(output string) ([]sample, error) {
	rows, err := csv.NewReader(strings.NewReader(output)).ReadAll()
	if err != nil {
		return nil, err
	}

	samples := []sample{}
	for _, row := range rows[1:] {
		observed, err := heySample(row)
		if err != nil {
			return nil, err
		}

		samples = append(samples, observed)
	}

	return samples, nil
}

func heySample(row []string) (sample, error) {
	seconds, err := strconv.ParseFloat(row[0], 64)
	if err != nil {
		return sample{}, err
	}

	status, err := strconv.Atoi(row[6])

	return sample{seconds: seconds, status: status}, err
}

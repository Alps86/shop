package performance_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cucumber/godog"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
)

func (current *world) record(route, cache string, samples []sample, delta counters) error {
	count := current.expectedRequests
	if len(samples) != count {
		return fmt.Errorf("%s/%s: %d Messproben statt %d", route, cache, len(samples), count)
	}

	report := current.decorate(summarize(samples), route, cache, delta)
	current.reports = append(current.reports, report)

	return current.verifyCache(report)
}

func summarize(samples []sample) measurement {
	times := []float64{}
	report := measurement{Requests: len(samples), CacheApplicability: "public-fragments", BrowserLCPMeasured: false, TargetP95Milliseconds: 100, Goal: "unmet", StatusCounts: map[int]int{}}
	for _, sample := range samples {
		times = append(times, sample.seconds*1000)
		report.observeSample(sample)
	}

	sort.Float64s(times)
	report.ErrorRate /= float64(len(samples))
	report.P50Milliseconds = percentile(times, 0.50)
	report.P95Milliseconds = percentile(times, 0.95)
	report.P99Milliseconds = percentile(times, 0.99)

	return report.withGoal()
}

func (report measurement) withGoal() measurement {
	if report.ErrorRate != 0 {
		report.Goal = "ineligible; measured HTTP errors"
		return report
	}

	if report.P95Milliseconds < report.TargetP95Milliseconds {
		report.Goal = "met-in-measured-profile"
	}

	return report
}

func percentile(values []float64, quantile float64) float64 {
	return values[int(math.Ceil(float64(len(values))*quantile))-1]
}

func (current *world) verifyCache(report measurement) error {
	if report.ErrorRate != 0 {
		return fmt.Errorf("%s/%s: Fehlerquote %f", report.Route, report.Cache, report.ErrorRate)
	}

	if report.Cache != "cold-cache-services-ready" {
		return current.verifyWarm(report)
	}

	if report.NewConnections >= report.Requests {
		return fmt.Errorf("kalte Requests %s ohne belegte TCPwiederverwendung", report.Route)
	}

	if report.hasWarmObjects() {
		return fmt.Errorf("kaltes Messfenster %s: Hits %d, Coalescing %d", report.Route, report.CacheHits, report.Coalesced)
	}

	return current.verifyMisses(report)
}

func (current *world) verifyMisses(report measurement) error {
	if current.privateRoute(report.Route) {
		return nil
	}

	if float64(report.CacheMisses) < float64(report.Requests) {
		return fmt.Errorf("ESI-Hostpartitionen %s: nur %d Misses für %d Proben", report.Route, report.CacheMisses, report.Requests)
	}

	return nil
}

func (current *world) privateRoute(route string) bool {
	return map[string]bool{"cart": true, "checkout": true, "order": true, "personalized-search": true}[route]
}

func (current *world) writeReport() error {
	environment, err := current.environmentReport()
	if err != nil {
		return err
	}

	output, err := json.MarshalIndent(environment, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf("Echte Messwerte ohne Browser-LCP: %s\n", output)

	return os.WriteFile(filepath.Join(current.directory, "measurement.json"), output, 0600)
}

func (current *world) environmentReport() (map[string]any, error) {
	environment := map[string]any{"CPU": runtime.NumCPU(), "OS": runtime.GOOS, "images": os.Getenv("SHOP_PERFORMANCE_IMAGES"), "dataset": current.datasetFacts(), "TCP": "hey keepalive enabled; NewConnections=-1 is unmeasured; actual curl cold num_connects per route", "compression": "disabled explicitly for hey; curl sends no Accept-Encoding", "ColdBoundary": "Only Varnish frame/fragment objects are cold; PostgreSQL/OpenSearch/Temporal services stay ready; OpenSearch request cache is not reset", "BrowserLCPMeasured": false, "CompletePageSLAAchieved": false, "measurements": current.reports, "BackendMeasurements": current.backendReports, "FragmentMeasurements": current.fragmentReports, "Failure": current.failure, "PostFailureProbes": current.diagnostics}
	memory, err := memoryKiB()
	if err != nil {
		return nil, err
	}

	environment["RAMKiB"] = memory
	osDetails, err := operatingSystem()
	if err != nil {
		return nil, err
	}

	environment["OSDetails"] = osDetails

	return environment, nil
}

func (current *world) observe(table *godog.Table) error {
	data := read(table)[0]
	if incorrectBrowserClaims(data) {
		return fmt.Errorf("HTTPmessung belegt weder Browser-LCP noch mutierende Commands")
	}

	if data["TargetP95Milliseconds"] != "100" {
		return fmt.Errorf("unverändertes Messziel benötigt 100 ms")
	}

	expected := map[string]string{"RequiredStatistics": "p50,p95,p99,errorRate", "RequiredEnvironment": "CPU,RAM,OS,images,dataset,TCP,compression", "Measurements": fmt.Sprint(len(current.reports)), "BackendMeasurements": fmt.Sprint(len(current.backendReports)), "FragmentMeasurements": fmt.Sprint(len(current.fragmentReports))}
	for key, actual := range expected {
		if data[key] != actual {
			return fmt.Errorf("Messbericht %s=%s statt %s", key, actual, data[key])
		}
	}

	return nil
}

func (report measurement) hasWarmObjects() bool {
	return report.CacheHits != 0 || report.Coalesced != 0
}

func incorrectBrowserClaims(data map[string]string) bool {
	return data["ProvesBrowserLCP"] != "false" || data["MutatingCommandsIncluded"] != "false"
}

func (current *world) datasetFacts() map[string]string {
	facts := map[string]string{}
	for _, key := range []string{"Dataset", "Products", "Variants", "Categories", "Profiles", "ProfileCodes", "Orders", "Concurrency", "Network", "CounterFlushMilliseconds"} {
		facts[key] = current.data[key]
	}

	return facts
}

func (current *world) decorate(report measurement, route, cache string, delta counters) measurement {
	report.measurementProfile = current.profile
	report.ComposedHTMLMeasured = true
	report.Route = route
	report.Cache = cache
	report.CacheHits = delta["MAIN.cache_hit"]
	report.CacheMisses = delta["MAIN.cache_miss"]
	report.Coalesced = delta["MAIN.busy_sleep"]
	if cache == "warm" {
		report.NewConnections = -1
	}

	if current.privateRoute(route) {
		report.CacheApplicability = "private,no-store; cache N/A"
	}

	return report
}

func (current *world) verifyWarm(report measurement) error {
	if current.privateRoute(report.Route) {
		return nil
	}

	if report.CacheHits == 0 {
		return fmt.Errorf("warmer öffentlicher Pfad %s ohne tatsächlichen Cachehit", report.Route)
	}

	return nil
}

func (report *measurement) observeSample(sample sample) {
	report.NewConnections += sample.connections
	report.StatusCounts[sample.status]++
	if sample.status != 200 {
		report.ErrorRate++
	}
}

func (current *world) finishReport(cause error) error {
	if cause != nil {
		current.failure = cause.Error()
	}

	return errors.Join(cause, current.writeReport())
}

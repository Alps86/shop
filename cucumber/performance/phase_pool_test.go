package performance_test

import (
	"context"
	"slices"
	"time"
)

func (current *world) phaseOutput(input poolPhaseInput, action poolAction) ([]byte, error) {
	if !current.observesPool(input) {
		return action()
	}

	phase := poolPhase{Route: input.route, Cache: input.cache, Boundary: current.profile.Boundary, LifetimeEvidence: "unproven", Note: "Outside load samples; /ready adds Diagnostics SQL and Temporal/Search checks. Background acquires and the end observer are included; no request-specific SQL spans or peak utilization."}
	phase.Start = current.phasePoolSnapshot()
	output, err := action()
	phase.End = current.phasePoolSnapshot()
	phase.compare()
	current.poolPhases = append(current.poolPhases, phase)

	return output, err
}

func (current *world) observesPool(input poolPhaseInput) bool {
	if current.profile.Boundary != "requestStart-to-last-composed-HTML-byte" {
		return false
	}

	if !slices.Contains([]string{"category", "personalized-search"}, input.route) {
		return false
	}

	return input.cache == "warm" || input.cache == "cold-cache-services-ready"
}

func (current *world) phasePoolSnapshot() snapshotPool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := current.readSnapshotPool(ctx)
	result.Note = "Single phase-boundary /ready; includes its own Diagnostics SQL; also checks Temporal and Search; no per-request spans"

	return result
}

func poolCumulativeKeys() []string {
	return []string{"PoolAcquireCount", "PoolAcquireDurationNanoseconds", "PoolEmptyAcquireCount", "PoolEmptyAcquireWaitTimeNanoseconds", "PoolCanceledAcquireCount"}
}

func (phase *poolPhase) compare() {
	phase.Comparison = phase.comparisonFailure()
	if phase.Comparison != "" {
		return
	}

	phase.Comparison = "conditional"
	phase.ConditionalOnSamePoolLifetime = true
	phase.ConditionalDeltas = map[string]uint64{}
	for _, key := range poolCumulativeKeys() {
		phase.ConditionalDeltas[key] = phase.End.Values[key] - phase.Start.Values[key]
	}
}

func (phase poolPhase) comparisonFailure() string {
	if !phase.validSnapshots() {
		return "not-comparable: snapshot status missing, not 200, incomplete or failed"
	}

	for _, key := range []string{"PoolMaxConns", "PoolNumCPU"} {
		if phase.Start.Values[key] != phase.End.Values[key] {
			return "not-comparable: pool configuration changed"
		}
	}

	return phase.counterFailure()
}

func (phase poolPhase) validSnapshots() bool {
	return phase.Start.Status == 200 && phase.End.Status == 200 && phase.Start.Error == "" && phase.End.Error == ""
}

func (phase poolPhase) counterFailure() string {
	for _, key := range poolCumulativeKeys() {
		if phase.End.Values[key] < phase.Start.Values[key] {
			return "not-comparable: cumulative counter decreased"
		}
	}

	return ""
}

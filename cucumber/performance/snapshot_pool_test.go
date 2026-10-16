package performance_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

func (current *world) readSnapshotPool(ctx context.Context) snapshotPool {
	proof := snapshotPool{StartedUTC: time.Now().UTC(), Note: "Subsequent single /ready observer; Diagnostics SQL counts itself; /ready also queries Temporal and Search; not request-specific SQL or acquire durations"}
	response, err := snapshotReady(ctx, current.internal)
	proof.EndedUTC = time.Now().UTC()
	proof.Status = response.status
	if err != nil {
		proof.Error = snapshotError(err)
		return proof
	}

	proof.Values, err = parseSnapshotPool(response.body)
	if err != nil {
		proof.Error = err.Error()
	}

	return proof
}

func snapshotReady(ctx context.Context, origin string) (probeResponse, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, origin+"/ready", nil)
	if err != nil {
		return probeResponse{}, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return probeResponse{}, err
	}

	body, readError := io.ReadAll(io.LimitReader(response.Body, 1024*1024))

	return probeResponse{status: response.StatusCode, body: string(body)}, errors.Join(readError, response.Body.Close())
}

func parseSnapshotPool(body string) (map[string]uint64, error) {
	values := map[string]uint64{}
	for _, field := range strings.Fields(body) {
		key, raw, _ := strings.Cut(strings.Trim(strings.TrimPrefix(field, "map["), "]"), ":")
		if !snapshotPoolKey(key) {
			continue
		}

		parsed, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return values, err
		}

		values[key] = parsed
	}

	return values, completeSnapshotPool(values)
}

func snapshotPoolKey(key string) bool {
	return slices.Contains(strings.Split("PoolMaxConns|PoolAcquiredConns|PoolIdleConns|PoolConstructingConns|PoolTotalConns|PoolAcquireCount|PoolAcquireDurationNanoseconds|PoolEmptyAcquireCount|PoolEmptyAcquireWaitTimeNanoseconds|PoolCanceledAcquireCount|PoolNumCPU", "|"), key)
}

func completeSnapshotPool(values map[string]uint64) error {
	if len(values) != 11 {
		return errors.New("Pool statistics incomplete; absent keys are unproven, not zero")
	}

	if values["PoolMaxConns"] == 0 {
		return errors.New("PoolMaxConns ist unbelegt oder ungültig")
	}

	if values["PoolNumCPU"] == 0 {
		return errors.New("PoolNumCPU ist unbelegt oder ungültig")
	}

	return nil
}

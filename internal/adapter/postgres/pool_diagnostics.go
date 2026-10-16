package postgres

import (
	"runtime"
	"strconv"
)

func (store *Store) poolDiagnostics() map[string]string {
	stat := store.pool.Stat()

	return map[string]string{
		"PoolMaxConns":                        strconv.FormatInt(int64(store.pool.Config().MaxConns), 10),
		"PoolAcquiredConns":                   strconv.FormatInt(int64(stat.AcquiredConns()), 10),
		"PoolIdleConns":                       strconv.FormatInt(int64(stat.IdleConns()), 10),
		"PoolConstructingConns":               strconv.FormatInt(int64(stat.ConstructingConns()), 10),
		"PoolTotalConns":                      strconv.FormatInt(int64(stat.TotalConns()), 10),
		"PoolAcquireCount":                    strconv.FormatInt(stat.AcquireCount(), 10),
		"PoolAcquireDurationNanoseconds":      strconv.FormatInt(stat.AcquireDuration().Nanoseconds(), 10),
		"PoolEmptyAcquireCount":               strconv.FormatInt(stat.EmptyAcquireCount(), 10),
		"PoolEmptyAcquireWaitTimeNanoseconds": strconv.FormatInt(stat.EmptyAcquireWaitTime().Nanoseconds(), 10),
		"PoolCanceledAcquireCount":            strconv.FormatInt(stat.CanceledAcquireCount(), 10),
		"PoolNumCPU":                          strconv.Itoa(runtime.NumCPU()),
	}
}

package postgres

import "github.com/Alps86/shop/internal/domain/value"

func (state ReplicationState) Resume(checkpoint uint64) (uint64, error) {
	if !state.Exists {
		return 0, value.NewError("ReindexRequired", "ReplicationSlot")
	}

	if state.Status == "lost" {
		return 0, value.NewError("ReindexRequired", "WAL")
	}

	if checkpoint == 0 {
		checkpoint = state.Confirmed
	}

	if checkpoint < state.Restart {
		return 0, value.NewError("ReindexRequired", "WAL")
	}

	return checkpoint, nil
}

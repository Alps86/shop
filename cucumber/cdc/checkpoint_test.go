package cdc_test

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func (current *world) readStreamResult() {
	if !current.streamRead {
		current.streamErr = <-current.stopped
		current.streamRead = true
	}
}

func (current *world) streamResult() error {
	current.readStreamResult()

	return current.streamErr
}

func (current *world) streamFailure() error {
	if errors.Is(current.streamErr, context.Canceled) {
		return nil
	}

	return current.streamErr
}

func (current *world) waitPosition(previous uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	return current.waitCheckpoint(ctx, ticker.C, previous)
}

func (current *world) waitCheckpoint(ctx context.Context, ticks <-chan time.Time, previous uint64) error {
	for {
		if finished, err := current.checkpointStatus(ctx, previous); finished {
			return err
		}

		if finished, err := current.nextCheckpoint(ctx, ticks, previous); finished {
			return err
		}
	}
}

func (current *world) checkpointStatus(ctx context.Context, previous uint64) (bool, error) {
	lsn, err := current.store.Checkpoint(ctx, current.config.Slot)
	if ctx.Err() != nil {
		return true, errors.Join(err, current.checkpointDeadline())
	}

	if err != nil {
		return true, err
	}

	return lsn > previous, nil
}

func (current *world) nextCheckpoint(ctx context.Context, ticks <-chan time.Time, previous uint64) (bool, error) {
	select {
	case <-current.done:
		return true, current.endedCheckpoint(ctx, previous)
	case <-ctx.Done():
		return true, current.checkpointDeadline()
	case <-ticks:
		return false, nil
	}
}

func (current *world) endedCheckpoint(ctx context.Context, previous uint64) error {
	current.readStreamResult()
	if reached, err := current.checkpointStatus(ctx, previous); reached {
		return errors.Join(err, current.streamFailure())
	}

	return errors.Join(fmt.Errorf("CDC-Verbraucher vor Checkpointfortschritt beendet"), current.streamErr)
}

func (current *world) checkpointDeadline() error {
	diagnosis := current.streamDeadline()
	closed := current.closeStream()

	return errors.Join(diagnosis, closed, current.streamFailure())
}

func (current *world) unreportedStreamFailure() error {
	if errors.Is(current.err, current.streamErr) {
		return nil
	}

	return current.streamFailure()
}

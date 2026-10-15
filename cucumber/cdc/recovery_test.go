package cdc_test

import (
	"context"
	"errors"
	"github.com/Alps86/shop/internal/adapter/cdc"
	"github.com/cucumber/godog"
	"time"
)

func (current *world) streamSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein eigener dauerhafter pgoutput-Verbraucher vorbereitet ist$`, current.startStream)
	ctx.Step(`^ein echter WAL-Verbraucher mit technischem Backoff läuft$`, current.startRunner)
	ctx.Step(`^das Ziel nach einem beobachteten Fehlversuch wieder erreichbar wird$`, current.restoreAfterAttempt)
	ctx.Step(`^das Projektionsziel vorübergehend ausfällt$`, current.failTarget)
	ctx.Step(`^der Verbraucher wegen des Zielausfalls beendet wird$`, current.failedStream)
	ctx.Step(`^der Verbraucher seinen Diagnosefehler meldet$`, current.failedStream)
	ctx.Step(`^derselbe Slot mit wieder verfügbarem Ziel neu startet$`, current.restartStream)
	ctx.Step(`^das Ereignis verarbeitet und der Verbraucher angehalten wird$`, current.stopStream)
	ctx.Step(`^die technische Wiederholung während ihres Backoffs gekündigt wird$`, current.cancelBackoff)
}

func (current *world) startStream() error {
	if current.consumer != nil {
		if err := current.closeStream(); err != nil {
			return err
		}
	}

	current.consumer = cdc.New(current.config)
	if err := current.consumer.Prepare(context.Background()); err != nil {
		return err
	}

	current.projected = make(chan struct{}, 10)
	current.stopped = make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	current.cancel = cancel
	go func() {
		current.stopped <- current.consumer.Stream(ctx)
	}()

	return nil
}

func (current *world) failTarget() {
	current.setAvailable(false)
}

func (current *world) failedStream() error {
	select {
	case err := <-current.stopped:
		current.err = err
	case <-time.After(10 * time.Second):
		return context.DeadlineExceeded
	}

	current.cancel()

	return current.closeStream()
}

func (current *world) restartStream() error {
	current.setAvailable(true)
	return current.startStream()
}

func (current *world) stopStream() error {
	select {
	case <-current.projected:
	case <-time.After(10 * time.Second):
		current.cancel()
		return context.DeadlineExceeded
	}

	if err := current.waitAcknowledged(); err != nil {
		current.cancel()
		return err
	}

	current.cancel()
	current.err = <-current.stopped
	if errors.Is(current.err, context.Canceled) {
		current.err = nil
	}

	return errors.Join(current.err, current.closeStream())
}

func (current *world) setAvailable(available bool) {
	current.mutex.Lock()
	defer current.mutex.Unlock()
	current.available = available
}

func (current *world) waitAcknowledged() error {
	return current.waitPosition(0)
}

func (current *world) waitPosition(previous uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		lsn, err := current.store.Checkpoint(ctx, current.config.Slot)
		if err != nil {
			return err
		}

		if lsn > previous {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (current *world) startRunner() error {
	current.consumer = cdc.New(current.config)
	if err := current.consumer.Prepare(context.Background()); err != nil {
		return err
	}

	current.projected = make(chan struct{}, 10)
	current.attempted = make(chan struct{}, 10)
	current.stopped = make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	current.cancel = cancel
	go func() {
		current.stopped <- current.consumer.Run(ctx)
	}()

	return nil
}

func (current *world) restoreAfterAttempt() error {
	select {
	case <-current.attempted:
		current.setAvailable(true)
		return nil
	case <-time.After(10 * time.Second):
		return context.DeadlineExceeded
	}
}

func (current *world) closeStream() error {
	if err := current.consumer.Close(context.Background()); err != nil {
		return err
	}

	return current.waitInactive()
}

func (current *world) waitInactive() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var active bool
		if err := current.admin.QueryRow(ctx, "SELECT COALESCE((SELECT active FROM pg_replication_slots WHERE slot_name=$1),false)", current.config.Slot).Scan(&active); err != nil {
			return err
		}

		if !active {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (current *world) cancelBackoff() error {
	select {
	case <-current.attempted:
	case <-time.After(10 * time.Second):
		return context.DeadlineExceeded
	}

	if err := current.waitInactive(); err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond)
	current.cancel()
	current.err = <-current.stopped
	if !errors.Is(current.err, context.Canceled) {
		return current.err
	}

	return current.closeStream()
}

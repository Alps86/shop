package runtime_test

import (
	"context"
	"github.com/cucumber/godog"
	"github.com/jackc/pglogrepl"
	"net"
	"strconv"
	"time"
)

func (current *world) failureSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^die geöffnete Fachverbindung geschlossen wird$`, current.closeDatabase)
	ctx.Step(`^ein gespeicherter WAL-Checkpoint ohne seinen Slot besteht$`, current.lostSlot)
	ctx.Step(`^die Bereitschaft nach dem technischen Retryintervall erneut abgefragt wird$`, current.afterRetry)
	ctx.Step(`^der HTTP-Dienst einen bereits geschlossenen TCP-Listener erhält$`, current.closedListener)
}

func (current *world) closedListener() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}

	if err := listener.Close(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	current.err = current.app.Serve(ctx, listener)

	return nil
}

func (current *world) closeDatabase() {
	current.app.Store.Close()
}

func (current *world) lostSlot() error {
	lsn, err := current.app.Store.WALBarrier(context.Background())
	if err != nil {
		return err
	}

	return current.app.Store.Acknowledge(context.Background(), current.config.Slot, pglogrepl.LSN(lsn).String(), nil)
}

func (current *world) afterRetry(table *godog.Table) error {
	milliseconds, err := strconv.Atoi(read(table)[0]["Milliseconds"])
	if err != nil {
		return err
	}

	time.Sleep(time.Duration(milliseconds) * time.Millisecond)

	return current.request(table)
}

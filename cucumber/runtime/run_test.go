package runtime_test

import (
	"context"
	"fmt"
	"github.com/cucumber/godog"
	"strconv"
	"strings"
	"time"
)

func (current *world) start() error {
	if current.err != nil {
		return current.err
	}

	ctx, cancel := context.WithCancel(context.Background())
	current.cancel = cancel
	current.finished = make(chan error, 1)
	go func() {
		current.finished <- current.app.Run(ctx)
	}()

	return nil
}

func (current *world) awaitReady(table *godog.Table) error {
	seconds, err := strconv.Atoi(read(table)[0]["TimeoutSeconds"])
	if err != nil {
		return err
	}

	deadline := time.Now().Add(time.Duration(seconds) * time.Second)
	for time.Now().Before(deadline) {
		if err := current.request(table); err != nil {
			current.status = 0
			current.body = err.Error()
		}

		if current.readyResponse(read(table)[0]["Contains"]) {
			return nil
		}

		time.Sleep(20 * time.Millisecond)
	}

	return fmt.Errorf("Bereitschaft: %d %s", current.status, current.body)
}

func (current *world) readyResponse(contains string) bool {
	return current.status == 200 && strings.Contains(current.body, contains)
}

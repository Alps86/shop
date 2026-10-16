package opensearch_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cucumber/godog"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

func (current *world) startTransportLog() {
	current.previousLogger = slog.Default()
	current.transportLog = &transportLog{}
	slog.SetDefault(slog.New(slog.NewJSONHandler(current.transportLog, nil)))
}

func (output *transportLog) Write(data []byte) (int, error) {
	output.mutex.Lock()
	defer output.mutex.Unlock()

	return output.buffer.Write(data)
}

func (output *transportLog) String() string {
	output.mutex.Lock()
	defer output.mutex.Unlock()

	return output.buffer.String()
}

func protocolParent(data row) (context.Context, context.CancelFunc, error) {
	if data["Parent"] == "canceled" {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		return ctx, cancel, nil
	}

	if data["ParentMilliseconds"] == "" {
		ctx, cancel := context.WithCancel(context.Background())

		return ctx, cancel, nil
	}

	duration, err := time.ParseDuration(data["ParentMilliseconds"] + "ms")
	ctx, cancel := context.WithTimeout(context.Background(), duration)

	return ctx, cancel, err
}

func (current *world) transportDiagnostic(table *godog.Table) error {
	data := read(table)[0]
	lines := strings.FieldsFunc(strings.TrimSpace(current.transportLog.String()), func(char rune) bool {
		return char == '\n'
	})
	if strconv.Itoa(len(lines)) != data["Calls"] {
		return fmt.Errorf("Transportdiagnosen %d statt %s", len(lines), data["Calls"])
	}

	if len(lines) == 0 {
		return nil
	}

	return safeDiagnostic(lines[0], data)
}

func safeDiagnostic(line string, expected row) error {
	var fields map[string]any
	if err := json.Unmarshal([]byte(line), &fields); err != nil {
		return err
	}

	for _, key := range []string{"phase", "dynamicType", "unwrappedType", "deadline", "canceled", "ctxErrCode"} {
		if fmt.Sprint(fields[key]) != expected[key] {
			return fmt.Errorf("Transportdiagnosefeld %s stimmt nicht", key)
		}
	}

	return diagnosticKeys(fields)
}

func diagnosticKeys(fields map[string]any) error {
	allowed := map[string]bool{"time": true, "level": true, "msg": true, "phase": true, "dynamicType": true, "unwrappedType": true, "deadline": true, "canceled": true, "ctxErrCode": true}
	for key := range fields {
		if !allowed[key] {
			return fmt.Errorf("unerlaubtes Diagnosefeld %s", key)
		}
	}

	return nil
}

package cdc_test

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/cdc"
	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
)

func (current *world) protocolSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^ergänzend Slotmetadatenwerte für eine Fortsetzung geprüft werden$`, current.slotMetadata)
	ctx.Step(`^ergänzend folgende Protokollframes durch denselben Decoder laufen$`, current.decodeFrames)
}

func (current *world) decodeFrames(input *godog.Table) error {
	consumer := cdc.New(current.config)
	for _, row := range input.Rows[1:] {
		frame, err := hex.DecodeString(row.Cells[0].Value)
		if err != nil {
			return err
		}

		err = consumer.DecodeFrame(context.Background(), frame)
		if value.Code(err) != row.Cells[1].Value {
			return fmt.Errorf("Frame %s: %s statt %s", row.Cells[0].Value, value.Code(err), row.Cells[1].Value)
		}
	}

	return nil
}

func (current *world) slotMetadata(input *godog.Table) error {
	for _, row := range input.Rows[1:] {
		restart, restartErr := strconv.ParseUint(row.Cells[1].Value, 10, 64)
		confirmed, confirmedErr := strconv.ParseUint(row.Cells[2].Value, 10, 64)
		checkpoint, checkpointErr := strconv.ParseUint(row.Cells[3].Value, 10, 64)
		if err := errors.Join(restartErr, confirmedErr, checkpointErr); err != nil {
			return err
		}

		state := postgres.ReplicationState{Exists: row.Cells[0].Value == "true", Restart: restart, Confirmed: confirmed, Status: row.Cells[4].Value}
		position, err := state.Resume(checkpoint)
		if value.Code(err) != row.Cells[5].Value {
			return fmt.Errorf("Slotstatus: %s statt %s", value.Code(err), row.Cells[5].Value)
		}

		if strconv.FormatUint(position, 10) != row.Cells[6].Value {
			return fmt.Errorf("ResumePosition: %d statt %s", position, row.Cells[6].Value)
		}
	}

	return nil
}

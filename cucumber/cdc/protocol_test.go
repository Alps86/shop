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
	"github.com/jackc/pgx/v5/pgproto3"
	"strconv"
)

func (current *world) protocolSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^ergänzend Slotmetadatenwerte für eine Fortsetzung geprüft werden$`, current.slotMetadata)
	ctx.Step(`^ergänzend folgende Protokollframes durch denselben Decoder laufen$`, current.decodeFrames)
	ctx.Step(`^ergänzend ein gültiger Keepalive nach geschlossener echter Verbindung Antwort verlangt$`, current.closedKeepalive)
	ctx.Step(`^ergänzend eine standardisierte Notice durch denselben Nachrichtendecoder läuft$`, current.decodeNotice)
}

func (current *world) decodeNotice() {
	consumer := cdc.New(current.config)
	current.err = consumer.DecodeMessage(context.Background(), &pgproto3.NoticeResponse{Severity: "NOTICE", Code: "00000", Message: "ergänzende Protokollfixture"})
}

func (current *world) closedKeepalive() error {
	current.consumer = cdc.New(current.config)
	if err := current.consumer.Prepare(context.Background()); err != nil {
		return err
	}

	if err := current.consumer.Close(context.Background()); err != nil {
		return err
	}

	frame, err := hex.DecodeString("6b0000000000000000000000000000000001")
	if err != nil {
		return err
	}

	current.err = current.consumer.DecodeFrame(context.Background(), frame)

	return nil
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

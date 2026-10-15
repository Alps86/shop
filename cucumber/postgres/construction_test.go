package postgres_test

import (
	"context"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/cucumber/godog"
)

func (current *world) construction(input *godog.Table) error {
	row := table(input)
	current.observed = map[string]string{}
	config := current.constructionConfig(row)

	store, err := postgres.Open(context.Background(), config)
	if err != nil {
		current.err = err
		return nil
	}
	defer store.Close()
	if row["Operation"] == "open" {
		return nil
	}

	_, current.err = store.Guests().Open(context.Background())
	current.observed = map[string]string{"EntropyUnavailable": strconv.FormatBool(errors.Is(current.err, io.EOF))}

	return nil
}

func (current *world) constructionConfig(row map[string]string) postgres.Config {
	config := postgres.Config{DSN: row["DSN"]}
	if row["Schema"] != "default" {
		config.Schema = current.schema
	}

	if row["DSN"] == "current" {
		config.DSN = os.Getenv("SHOP_TEST_POSTGRES_DSN")
	}

	if row["Entropy"] == "unavailable" {
		config.Entropy = strings.NewReader("")
	}

	return config
}

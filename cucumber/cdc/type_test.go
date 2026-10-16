package cdc_test

import (
	"context"
	"github.com/Alps86/shop/internal/adapter/cdc"
	"github.com/Alps86/shop/internal/adapter/event"
	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/application"
	"github.com/jackc/pgx/v5"
	"sync"
)

type world struct {
	restrictedRole  string
	foreignDatabase string
	priorCheckpoint uint64
	attempted       chan struct{}
	mutex           sync.Mutex
	projected       chan struct{}
	available       bool
	store           *postgres.Store
	admin           *pgx.Conn
	consumer        *cdc.Consumer
	source          *cdc.ReindexSource
	snapshot        event.ReindexSnapshot
	config          cdc.Config
	events          []application.ProjectionEvent
	cancel          context.CancelFunc
	streamRead      bool
	streamErr       error
	stopped         chan error
	done            chan struct{}
	err             error
}

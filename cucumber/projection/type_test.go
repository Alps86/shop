package projection_test

import (
	"context"
	"github.com/Alps86/shop/internal/adapter/cdc"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/application"
	"github.com/jackc/pgx/v5"
)

type world struct {
	Earlier            application.ProjectionEvent
	PreviousCheckpoint uint64
	Store              *postgres.Store
	Admin              *pgx.Conn
	Config             cdc.Config
	Search             opensearch.Adapter
	Endpoint           string
	Alias              string
	Consumer           *cdc.Consumer
	Cancel             context.CancelFunc
	Stopped            chan error
	Manifest           postgres.ImportManifest
	Source             *gatedSource
	Rebuilt            chan error
	RebuildComplete    bool
	TargetIndex        string
}

type gatedSource struct {
	Source         opensearch.ReindexSource
	SnapshotReady  chan struct{}
	SnapshotResume chan struct{}
	CatchupReady   chan struct{}
	CatchupResume  chan struct{}
}

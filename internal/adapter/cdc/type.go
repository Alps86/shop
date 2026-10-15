package cdc

import (
	"github.com/Alps86/shop/internal/adapter/event"
	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/application"
	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgconn"
	"sync"
	"time"
)

type Config struct {
	DSN         string
	Schema      string
	Publication string
	Slot        string
	Store       *postgres.Store
	Target      application.ProjectionPort
}

type Consumer struct {
	retryDelay       time.Duration
	config           Config
	connection       *pgconn.PgConn
	connectionConfig *pgconn.Config
	relations        map[uint32]*pglogrepl.RelationMessage
	pending          []event.ProjectionEnvelope
	acknowledged     pglogrepl.LSN
	stop             pglogrepl.LSN
}

type ReindexSource struct {
	mutex     sync.Mutex
	Config    Config
	snapshots map[string]*Consumer
}

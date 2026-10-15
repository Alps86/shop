package event

import (
	"time"

	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
)

type ProjectionEnvelope struct {
	EventID       string
	SchemaVersion int
	OccurredAt    time.Time
	Kind          string
	Context       value.RequestContext
	Projection    application.ProjectionEvent
}

type ProcessEnvelope struct {
	EventID string
	Kind    string
	CartID  value.CartID
	OrderID value.OrderID
	Binding application.ProcessBinding
}

type ReindexSnapshot struct {
	Events []application.ProjectionEvent
	Cursor string
}

type ProjectionEnvelopes []ProjectionEnvelope

type projectionRegistry map[string]ProjectionEnvelope

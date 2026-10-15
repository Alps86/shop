package temporal

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"go.temporal.io/sdk/client"
	"time"
)

type ExpirationPort interface {
	Expire(context.Context, value.CartID, time.Time) (sale.Cart, error)
}

type Adapter struct {
	Client     client.Client
	Service    application.Service
	Expiration ExpirationPort
	Queue      string
}

type Start struct {
	ID      string
	Binding application.ProcessBinding
}

type Status struct {
	State   string
	Binding application.ProcessBinding
}

type cartResult struct {
	View    application.CartView
	Failure value.Optional[value.Error]
}

type purchaseResult struct {
	Result  application.PurchaseResult
	Failure value.Optional[value.Error]
}

type cartProcess struct {
	Start    Start
	Status   Status
	Terminal bool
}

type orderProcess struct {
	Start    Start
	Status   Status
	Terminal bool
}

type updateHandler struct {
	Name    string
	Handler any
}

type PaymentSimulation struct {
	Approved bool
	Delay    time.Duration
}

type workflowUpdate struct {
	ID      string
	UserID  value.UserID
	Name    string
	Command any
}

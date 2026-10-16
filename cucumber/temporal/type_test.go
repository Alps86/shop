package temporal_test

import (
	"context"
	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/adapter/temporal"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/jackc/pgx/v5"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"sync"
	"time"
)

type world struct {
	timerCheckpoint sale.Cart
	lastPlace       application.PlaceCommand
	commitFault     *purchaseFault
	clockNow        time.Time
	activityBase    time.Time
	activityCommand application.ChangeCommand
	newCartID       value.CartID
	gate            *gatedPayment
	processStatus   temporal.Status
	queryErr        error
	authorization   application.PaymentAuthorization
	initialPurchase sale.Order
	service         application.Service
	guest           application.GuestView
	databaseStore   *postgres.Store
	schema          string
	admin           *pgx.Conn
	testEnvironment *testsuite.TestWorkflowEnvironment
	started         time.Time
	store           *cartStore
	runner          worker.Worker
	view            application.CartView
	purchase        application.PurchaseResult
	err             error
	connection      client.Client
	adapter         temporal.Adapter
	id              string
	runID           string
}

type cartStore struct {
	writeDelay         time.Duration
	expirationFailures int
	expirationError    error
	expirationCalls    int
	mutex              sync.Mutex
	cart               sale.Cart
}

type purchaseFault struct {
	Committed   chan struct{}
	Transaction application.PurchaseTransaction
	Remaining   int
	Mutex       sync.Mutex
}

type gatedPayment struct {
	Payment application.PaymentPort
	Started chan struct{}
	Resume  chan struct{}
}

type registrationInterceptor struct {
	interceptor.WorkerInterceptorBase
	Operation string
}

type registrationInbound struct {
	interceptor.WorkflowInboundInterceptorBase
	Operation string
}

type registrationOutbound struct {
	interceptor.WorkflowOutboundInterceptorBase
	Operation string
}

type failedFulfillment struct {
	Failure error
}

type fixedClock struct {
	At time.Time
}

type restartReadiness struct {
	ctx      context.Context
	interval time.Duration
	probes   int
	canceled int
	first    string
}

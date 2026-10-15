package postgres_test

import (
	"context"
	"errors"
	"github.com/Alps86/shop/internal/adapter/event"
	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"maps"
	"strconv"
)

func (current *world) cancellationSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein bereits gekündigter Kontext die öffentliche SQL-Grenze erreicht$`, current.cancelledBoundary)
	ctx.Step(`^die öffentliche SQL-Grenze nach dem Datenbankfehler aufgerufen wird$`, current.failedBoundary)
	ctx.Step(`^eine echte konkurrierende Schreibsperre das Requestbudget überschreitet$`, current.blockedWrite)
}

func (current *world) cancelledBoundary(input *godog.Table) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := current.writeBoundaries()
	maps.Copy(calls, current.readBoundaries())
	current.err = calls[table(input)["Boundary"]](ctx)
	current.observed = map[string]string{"Cancelled": strconv.FormatBool(errors.Is(current.err, context.Canceled))}
}

func (current *world) writeBoundaries() map[string]func(context.Context) error {
	return map[string]func(context.Context) error{"Migrate": current.store.Migrate, "Import": current.cancelImport, "GuestOpen": current.cancelGuestOpen, "StartCart": current.cancelStart, "SaveCart": current.cancelSave, "Inventory": current.cancelInventory, "Purchase": current.cancelPurchase, "FulfillmentSave": current.cancelDelivery, "Expire": current.cancelExpire, "Profile": current.cancelProfile, "Snapshot": current.cancelSnapshot, "ProcessDispatch": current.cancelDispatch}
}

func (current *world) readBoundaries() map[string]func(context.Context) error {
	return map[string]func(context.Context) error{"Resolve": current.cancelResolve, "CartGet": current.cancelCartGet, "Product": current.cancelProduct, "AdminOrders": current.cancelAdminOrders, "Diagnostics": current.cancelDiagnostics, "Checkpoint": current.cancelCheckpoint, "Slot": current.cancelSlot, "Seen": current.cancelSeen, "ReplayChange": current.cancelReplayChange, "ReplayConfirm": current.cancelReplayConfirm, "Evaluation": current.cancelEvaluation, "Variant": current.cancelVariant}
}

func (current *world) cancelImport(ctx context.Context) error {
	_, err := current.store.Import(ctx, current.manifest)
	return err
}

func (current *world) cancelGuestOpen(ctx context.Context) error {
	_, err := current.store.Guests().Open(ctx)
	return err
}

func (current *world) cancelStart(ctx context.Context) error {
	_, err := current.store.Carts().Start(ctx, current.guest.UserID, "cancel-a", current.Now())
	return err
}

func (current *world) cancelSave(ctx context.Context) error {
	_, err := current.store.Carts().Save(ctx, application.CartWrite{})
	return err
}

func (current *world) cancelInventory(ctx context.Context) error {
	_, err := current.store.Inventory().Change(ctx, application.InventoryCommand{SKU: "PHONE", CommandID: "cancel-a"})
	return err
}

func (current *world) cancelPurchase(ctx context.Context) error {
	_, err := current.store.Purchases().Commit(ctx, application.PurchaseCommand{CartID: "c-cancel", CommandID: "cancel-a"})
	return err
}

func (current *world) cancelDelivery(ctx context.Context) error {
	fulfillment := sale.Fulfillment{Updates: []sale.FulfillmentUpdate{{OrderID: "o-a", CommandID: "cancel-a"}}}
	return current.store.Fulfillment().Save(ctx, fulfillment, "cancel-a")
}

func (current *world) cancelExpire(ctx context.Context) error {
	_, err := current.store.Carts().Expire(ctx, "c-cancel", current.Now())
	return err
}

func (current *world) cancelProfile(ctx context.Context) error {
	_, err := current.store.Guests().Profile(ctx, current.guest.UserID, "neutral")
	return err
}

func (current *world) cancelSnapshot(ctx context.Context) error {
	_, err := current.store.ExportedSnapshot(ctx, "cancelled")
	return err
}

func (current *world) cancelDispatch(ctx context.Context) error {
	return (postgres.ProcessDispatcher{Store: current.store}).Dispatch(ctx)
}

func (current *world) cancelResolve(ctx context.Context) error {
	_, err := current.store.Guests().Resolve(ctx, current.guest.Session)
	return err
}

func (current *world) cancelCartGet(ctx context.Context) error {
	_, err := current.store.Carts().Get(ctx, "c-cancel")
	return err
}

func (current *world) cancelProduct(ctx context.Context) error {
	_, err := current.store.Catalog().Product(ctx, "p-phone", value.Optional[value.SKU]{}, value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"})
	return err
}

func (current *world) cancelAdminOrders(ctx context.Context) error {
	_, err := current.store.AdminOrders().List(ctx, 1)
	return err
}

func (current *world) cancelDiagnostics(ctx context.Context) error {
	_, err := current.store.Diagnostics(ctx)
	return err
}

func (current *world) cancelCheckpoint(ctx context.Context) error {
	_, err := current.store.Checkpoint(ctx, "cancel-consumer")
	return err
}

func (current *world) cancelSlot(ctx context.Context) error {
	_, err := current.store.ReplicationSlot(ctx, "cancel-consumer")
	return err
}

func (current *world) cancelSeen(ctx context.Context) error {
	_, err := current.store.Seen(ctx, "cancel-consumer", event.ProjectionEnvelope{EventID: "e-cancel"})
	return err
}

func (current *world) cancelReplayChange(ctx context.Context) error {
	_, _, err := current.store.Carts().ReplayChange(ctx, application.ChangeCommand{CartID: "c-cancel", CommandID: "cancel-a"})
	return err
}

func (current *world) cancelReplayConfirm(ctx context.Context) error {
	_, _, err := current.store.Carts().ReplayConfirm(ctx, application.ConfirmCommand{CartID: "c-cancel", CommandID: "cancel-a"})
	return err
}

func (current *world) failedBoundary(input *godog.Table) {
	calls := current.writeBoundaries()
	maps.Copy(calls, current.readBoundaries())
	current.err = calls[table(input)["Boundary"]](context.Background())
	current.observed = map[string]string{}
}

func (current *world) cancelEvaluation(ctx context.Context) error {
	_, err := current.store.Evaluator().Evaluate(ctx, current.cart.Cart)
	return err
}

func (current *world) cancelVariant(ctx context.Context) error {
	_, err := current.store.Catalog().ProductForSKU(ctx, "PHONE", value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"})
	return err
}

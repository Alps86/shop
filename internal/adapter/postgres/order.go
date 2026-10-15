package postgres

import (
	"context"

	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/jackc/pgx/v5"
)

func (repository OrderRepository) Get(ctx context.Context, id value.OrderID) (sale.Order, error) {
	var order sale.Order
	err := repository.store.pool.QueryRow(ctx, "SELECT payload FROM orders WHERE id=$1", id).Scan(&order)

	return order, databaseError(err)
}

func (repository OrderRepository) ForCart(ctx context.Context, cart value.CartID) (value.OrderID, error) {
	var id value.OrderID
	err := repository.store.pool.QueryRow(ctx, "SELECT id FROM orders WHERE cart_id=$1", cart).Scan(&id)

	return id, databaseError(err)
}

func (repository AdminOrderRepository) Get(ctx context.Context, id value.OrderID) (application.OrderView, error) {
	order, err := repository.store.Orders().Get(ctx, id)
	if err != nil {
		return application.OrderView{}, err
	}

	fulfillment, err := repository.store.Fulfillment().Get(ctx, id)

	return application.OrderView{Purchase: order, Fulfillment: application.NewFulfillmentView(fulfillment)}, err
}

func (repository AdminOrderRepository) List(ctx context.Context, page int) (application.AdminOrderPage, error) {
	if page < 1 {
		return application.AdminOrderPage{}, value.NewError("InvalidInput", "Page")
	}

	result := application.AdminOrderPage{Page: page, PageSize: 12}
	if err := repository.store.pool.QueryRow(ctx, "SELECT count(*) FROM orders").Scan(&result.Total); err != nil {
		return application.AdminOrderPage{}, databaseError(err)
	}

	rows, err := repository.store.pool.Query(ctx, "SELECT payload,fulfillment FROM orders ORDER BY id LIMIT 12 OFFSET $1", (page-1)*12)
	if err != nil {
		return application.AdminOrderPage{}, databaseError(err)
	}

	defer rows.Close()

	return repository.orderRows(rows, result)
}

func (repository AdminOrderRepository) orderRows(rows pgx.Rows, result application.AdminOrderPage) (application.AdminOrderPage, error) {
	for rows.Next() {
		var order sale.Order
		var fulfillment sale.Fulfillment
		if err := rows.Scan(&order, &fulfillment); err != nil {
			return application.AdminOrderPage{}, databaseError(err)
		}

		result.Orders = append(result.Orders, application.OrderView{Purchase: order, Fulfillment: application.NewFulfillmentView(fulfillment)})
	}

	return result, databaseError(rows.Err())
}

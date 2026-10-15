package commerce_test

import (
	"context"
	"github.com/Alps86/shop/internal/adapter/temporal"
	"time"
)

func (current *world) orderProcess() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	for {
		encoded, err := current.app.Client.QueryWorkflow(ctx, "order/"+string(current.purchase.OrderID), "", "ProcessStatus")
		if err == nil {
			var status temporal.Status
			if err := encoded.Get(&status); err != nil {
				return "", err
			}

			return status.State, nil
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
}

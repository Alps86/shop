package temporal_test

import (
	"context"
	"github.com/Alps86/shop/internal/application"
)

func (fault *purchaseFault) Commit(ctx context.Context, command application.PurchaseCommand) (application.PurchaseResult, error) {
	fault.Mutex.Lock()
	defer fault.Mutex.Unlock()
	result, err := fault.Transaction.Commit(ctx, command)
	if err != nil {
		return result, err
	}

	if fault.Remaining > 0 {
		fault.Remaining--
		return fault.lostResponse(ctx)
	}

	return result, nil
}

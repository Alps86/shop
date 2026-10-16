package opensearch

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

func logTransportFailure(ctx context.Context, err error) {
	slog.ErrorContext(ctx, "OpenSearch transport failure",
		"phase", "transport",
		"dynamicType", fmt.Sprintf("%T", err),
		"unwrappedType", fmt.Sprintf("%T", errors.Unwrap(err)),
		"deadline", errors.Is(err, context.DeadlineExceeded),
		"canceled", errors.Is(err, context.Canceled),
		"ctxErrCode", contextErrorCode(ctx.Err()))
}

func contextErrorCode(err error) string {
	if err == nil {
		return "none"
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return "deadline"
	}

	return "canceled"
}

package runtime

import (
	"context"
	"errors"
	"github.com/Alps86/shop/internal/domain/value"
	"io"
	"net"
	"net/http"
	"time"
)

func (config Config) Execute(ctx context.Context, arguments []string) error {
	if len(arguments) == 0 {
		return config.Run(ctx)
	}

	if len(arguments) != 1 {
		return value.NewError("InvalidInput", "Arguments")
	}

	if arguments[0] != "healthcheck" {
		return value.NewError("InvalidInput", "Command")
	}

	return config.Check(ctx)
}

func (config Config) Check(ctx context.Context) error {
	_, port, err := net.SplitHostPort(config.Listen)
	if err != nil {
		return err
	}

	return config.probe(ctx, "http://"+net.JoinHostPort("127.0.0.1", port)+"/ready")
}

func (config Config) probe(ctx context.Context, endpoint string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	_, readError := io.Copy(io.Discard, response.Body)
	if response.StatusCode != http.StatusOK {
		return errors.Join(value.NewError("DependencyUnavailable", "Readiness"), readError, response.Body.Close())
	}

	return errors.Join(readError, response.Body.Close())
}

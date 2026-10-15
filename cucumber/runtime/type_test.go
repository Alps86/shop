package runtime_test

import (
	"context"
	"github.com/Alps86/shop/internal/runtime"
	"net/http/httptest"
	"os/exec"
)

type world struct {
	app             *runtime.App
	config          runtime.Config
	server          *httptest.Server
	status          int
	body            string
	err             error
	cancel          context.CancelFunc
	finished        chan error
	binary          string
	command         *exec.Cmd
	processFinished chan error
	baseURL         string
	exitCode        int
}

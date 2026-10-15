package runtime_test

import (
	"errors"
	"fmt"
	"github.com/cucumber/godog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

func (current *world) commandSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^die ausführbare Shopdatei gebaut wird$`, current.build)
	ctx.Step(`^die Shopdatei als eigener Dienst gestartet wird$`, current.startCommand)
	ctx.Step(`^die Shopdatei mit Argumenten beendet wird$`, current.execute)
	ctx.Step(`^lautet der Prozessstatus$`, current.commandObserved)
}

func (current *world) build(table *godog.Table) error {
	if read(table)[0]["Package"] != "./cmd/shop" {
		return fmt.Errorf("nur der eigene Shopbefehl darf gebaut werden")
	}

	directory, err := os.MkdirTemp("", "shop-command-")
	if err != nil {
		return err
	}

	current.binary = filepath.Join(directory, "shop")

	return current.buildBinary()
}

func (current *world) buildBinary() error {
	command := exec.Command("go", "build")
	command.Args = append(command.Args, "-o", current.binary)
	if os.Getenv("SHOP_PROCESS_COVERDIR") != "" {
		command.Args = append(command.Args, "-cover", "-coverpkg=./...", "-covermode=atomic")
	}

	command.Args = append(command.Args, "./cmd/shop")
	command.Dir = "../.."
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Shopbuild: %w: %s", err, output)
	}

	return nil
}

func (current *world) startCommand() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}

	current.config.Listen = listener.Addr().String()
	if err := listener.Close(); err != nil {
		return err
	}

	current.baseURL = "http://" + current.config.Listen

	return current.launchCommand()
}

func (current *world) launchCommand() error {
	current.command = exec.Command("./shop")
	current.command.Dir = filepath.Dir(current.binary)
	current.command.Env = current.environment()
	if err := current.command.Start(); err != nil {
		return err
	}

	current.processFinished = make(chan error, 1)
	go func() {
		current.processFinished <- current.command.Wait()
	}()

	return nil
}

func (current *world) environment() []string {
	config := current.config
	environment := append(os.Environ(), "SHOP_POSTGRES_DSN="+config.DSN, "SHOP_SCHEMA="+config.Schema, "SHOP_TEMPORAL_ADDRESS="+config.TemporalAddress, "SHOP_TEMPORAL_NAMESPACE="+config.TemporalNamespace, "SHOP_TASK_QUEUE="+config.TaskQueue, "SHOP_OPENSEARCH_URL="+config.SearchEndpoint, "SHOP_SEARCH_INDEX="+config.SearchIndex, "SHOP_CDC_PUBLICATION="+config.Publication, "SHOP_CDC_SLOT="+config.Slot, "SHOP_LISTEN="+config.Listen, "SHOP_SEED_FILE="+config.SeedFile)
	if directory := os.Getenv("SHOP_PROCESS_COVERDIR"); directory != "" {
		environment = append(environment, "GOCOVERDIR="+directory)
	}

	return environment
}

func (current *world) execute(table *godog.Table) error {
	command, err := current.requestedCommand(read(table)[0]["Arguments"])
	if err != nil {
		return err
	}

	command.Dir = filepath.Dir(current.binary)
	command.Env = current.environment()
	err = command.Run()
	current.exitCode = 0
	var exited *exec.ExitError
	if errors.As(err, &exited) {
		current.exitCode = exited.ExitCode()
		return nil
	}

	return err
}

func (current *world) requestedCommand(arguments string) (*exec.Cmd, error) {
	commands := map[string]*exec.Cmd{
		"healthcheck":       exec.Command("./shop", "healthcheck"),
		"unknown":           exec.Command("./shop", "unknown"),
		"healthcheck,extra": exec.Command("./shop", "healthcheck", "extra"),
	}
	command, exists := commands[arguments]
	if !exists {
		return nil, fmt.Errorf("nicht erlaubter Testbefehl: %s", arguments)
	}

	return command, nil
}

func (current *world) commandObserved(table *godog.Table) error {
	if strconv.Itoa(current.exitCode) != read(table)[0]["ExitCode"] {
		return fmt.Errorf("Prozessstatus: %d", current.exitCode)
	}

	return nil
}

func (current *world) stopCommand() error {
	if current.command == nil {
		return current.removeBinary()
	}

	if err := current.command.Process.Signal(syscall.SIGTERM); err != nil {
		return err
	}

	select {
	case err := <-current.processFinished:
		return errors.Join(err, current.removeBinary())
	case <-time.After(10 * time.Second):
		return errors.Join(fmt.Errorf("Shopdienst stoppt nicht"), current.command.Process.Kill())
	}
}

func (current *world) removeBinary() error {
	if current.binary == "" {
		return nil
	}

	return os.RemoveAll(filepath.Dir(current.binary))
}

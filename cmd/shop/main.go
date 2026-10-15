package main

import (
	"context"
	"github.com/Alps86/shop/internal/runtime"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := runtime.NewEnvironmentConfig().Execute(ctx, os.Args[1:])
	stop()
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

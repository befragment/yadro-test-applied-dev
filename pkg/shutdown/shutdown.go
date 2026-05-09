package shutdown

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func WaitForShutdown() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel()
	}()

	return ctx
}

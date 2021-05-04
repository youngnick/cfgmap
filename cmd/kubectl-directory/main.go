package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/youngnick/directory/cmd/kubectl-directory/cli"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // required for GKE
)

func main() {

	stopCh := setupSignalHandler()
	ctx := contextWithStopCh(context.Background(), stopCh)
	cmds := cli.NewRootCmd(ctx, os.Stdin, os.Stdout, os.Stderr)

	cobra.CheckErr(cmds.Execute())
}

// ContextWithStopCh will wrap a context with a stop channel.
// When the provided stopCh closes, the cancel() will be called on the context.
// This provides a convenient way to represent a stop channel as a context.
func contextWithStopCh(ctx context.Context, stopCh <-chan struct{}) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()
		select {
		case <-ctx.Done():
		case <-stopCh:
		}
	}()
	return ctx
}

// SetupSignalHandler registered for SIGTERM and SIGINT. A stop channel is returned
// which is closed on one of these signals. If a second signal is caught, the program
// is terminated with exit code 1.
func setupSignalHandler() <-chan struct{} {

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	return stop
}

package contexthelp

import "context"

// StopCh will wrap a context with a stop channel.
// When the provided stopCh closes, the cancel() will be called on the context.
// This provides a convenient way to represent a stop channel as a context.
func WithStopCh(ctx context.Context, stopCh <-chan struct{}) context.Context {
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

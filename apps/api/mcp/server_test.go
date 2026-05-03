package mcp

import (
	"context"
	"testing"
	"time"

	"4ks/apps/api/app"
)

func TestStartReturnsWhenContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan error, 1)
	go func() {
		done <- New(nil, app.Services{}).Start(ctx)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Start returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Start did not return after context cancellation")
	}
}

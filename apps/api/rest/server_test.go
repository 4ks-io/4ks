package rest

import (
	"context"
	"testing"
	"time"

	"4ks/apps/api/app"
	controllers "4ks/apps/api/controllers"
	"4ks/apps/api/utils"
)

func TestServerStartStopsOnContextCancel(t *testing.T) {
	ginMode := utils.MinimalRuntimeConfig()
	ginMode.Routes.Port = "0"

	srv, err := New(ginMode, app.Services{}, Deps{
		Version: "test-version",
		System: controllers.SystemControllerDeps{
			DB:        testProber{},
			Search:    testProber{},
			Messaging: testProber{},
			Storage:   testProber{},
		},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() {
		done <- srv.Start(ctx)
	}()

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Start returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after context cancellation")
	}
}

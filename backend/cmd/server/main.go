// Command server runs the All-in-One Tools application: a single binary
// that serves both the REST API and the embedded Vue frontend.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/naufalzaid17/all-in-one-tools/backend/internal/api"
	"github.com/naufalzaid17/all-in-one-tools/backend/internal/core"
	"github.com/naufalzaid17/all-in-one-tools/backend/internal/tools/jsonutils"
	"github.com/naufalzaid17/all-in-one-tools/backend/internal/tools/qrcode"
	"github.com/naufalzaid17/all-in-one-tools/backend/internal/tools/security"
	"github.com/naufalzaid17/all-in-one-tools/backend/public"
)

const (
	defaultAddr     = ":8080"
	shutdownTimeout = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func run() error {
	registry := core.NewRegistry()
	registry.MustRegister(jsonutils.New())
	registry.MustRegister(security.New())
	registry.MustRegister(qrcode.New())

	addr := defaultAddr
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           api.NewRouter(registry, public.FS()),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Serve until interrupted, then drain in-flight requests.
	errCh := make(chan error, 1)
	go func() {
		log.Printf("all-in-one-tools listening on http://localhost%s", addr)
		errCh <- srv.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-stop:
		log.Printf("received %s, shutting down", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return err
	}
	if err := <-errCh; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

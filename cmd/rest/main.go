package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/spacesedan/go-sequence/internal/handlers"
	"github.com/spacesedan/go-sequence/internal/lobby"
)

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func main() {

	errC, err := run()
	if err != nil {
		log.Fatalf("Error when starting server: %v", err)
	}

	if err := <-errC; err != nil {
		log.Fatalf("Error while running: %v", err)
	}

}

type ServerConfig struct {
	address string
	logger  *slog.Logger
}

func run() (<-chan error, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("Logger Active")

	errC := make(chan error, 1)

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGKILL)

	serverConfig := ServerConfig{
		address: ":42069",
		logger:  logger,
	}

	srv, _ := newServer(serverConfig)

	go func() {
		<-ctx.Done()

		ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*5)

		defer func() {
			cancel()
			stop()
			close(errC)
		}()

		srv.SetKeepAlivesEnabled(false)

		if err := srv.Shutdown(ctxTimeout); err != nil {
			errC <- err
		}

		logger.Info("Shutdown Complete")
	}()

	go func() {
		logger.Info("Listening and serving to:", slog.String("addr", serverConfig.address))

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errC <- err
		}
	}()

	return errC, nil

}

func newServer(sc ServerConfig) (*http.Server, error) {
	r := chi.NewRouter()

	// start services
    lm := lobby.NewLobbyManager(sc.logger)

	// Register handlers
    handlers.NewLobbyHandler(lm, sc.logger).Register(r)
    handlers.NewViewHandler().Register(r)


	// handler static files
	fs := http.FileServer(http.Dir("assets"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	return &http.Server{
		Handler:           r,
		Addr:              sc.address,
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       1 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
	}, nil
}

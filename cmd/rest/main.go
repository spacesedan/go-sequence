package main

import (
	"context"
	"encoding/gob"
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
	"github.com/go-redis/redis/v8"

	"github.com/spacesedan/go-sequence/cmd/internal"
	"github.com/spacesedan/go-sequence/internal/client"
	"github.com/spacesedan/go-sequence/internal/handlers"
	"github.com/spacesedan/go-sequence/internal/lobby"
	"github.com/spacesedan/go-sequence/internal/services"
)

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func main() {
	gob.Register(client.WsClient{})
	gob.Register(lobby.WsPayload{})
	gob.Register(lobby.WsResponse{})

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
	redis   *redis.Client
}

func run() (<-chan error, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	rdb, err := internal.NewRedis(logger)
	if err != nil {
		return nil, services.WrapErrorf(err, services.ErrorCodeUnknown, "internal.NewRedis")
	}

	errC := make(chan error, 1)

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGKILL)

	serverConfig := ServerConfig{
		address: ":42069",
		logger:  logger,
		redis:   rdb,
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
	lm := lobby.NewLobbyManager(sc.redis, sc.logger)
	go lm.Run()

	// Register handlers
	handlers.NewLobbyHandler(sc.redis, lm, sc.logger).Register(r)
	handlers.NewViewHandler(sc.redis, lm).Register(r)

	// handler static files
	fs := http.FileServer(http.Dir("assets"))
	bundle := http.FileServer(http.Dir("dist"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))
	r.Handle("/bundle/*", http.StripPrefix("/bundle/", bundle))

	return &http.Server{
		Handler:           r,
		Addr:              sc.address,
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       1 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
	}, nil
}

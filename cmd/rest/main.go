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

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/spacesedan/go-sequence/cmd/internal"
	"github.com/spacesedan/go-sequence/internal/handlers"
	"github.com/spacesedan/go-sequence/internal/lobby"
	"github.com/spacesedan/go-sequence/internal/services"
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
	address   string
	logger    *slog.Logger
	redisPool *redis.Pool
}

func run() (<-chan error, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("Logger Active")

	redisPool, err := internal.NewRedis()
	if err != nil {
		return nil, services.WrapErrorf(err, services.ErrorCodeUnknown, "internal.NewRedis")
	}

	errC := make(chan error, 1)

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGKILL)

	serverConfig := ServerConfig{
		address:   ":42069",
		logger:    logger,
		redisPool: redisPool,
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
	sessionManager := scs.New()
	sessionManager.Store = redisstore.New(sc.redisPool)
	sessionManager.Lifetime = 24 * time.Hour

	// start services
	lm := lobby.NewLobbyManager(sc.logger)

	// Register handlers
	handlers.NewLobbyHandler(lm, sc.logger, sessionManager).Register(r)
	handlers.NewViewHandler(sessionManager, lm).Register(r)

	// handler static files
	fs := http.FileServer(http.Dir("assets"))
	bundle := http.FileServer(http.Dir("dist"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))
	r.Handle("/bundle/*", http.StripPrefix("/bundle/", bundle))

	return &http.Server{
		Handler:           sessionManager.LoadAndSave(r),
		Addr:              sc.address,
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       1 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
	}, nil
}

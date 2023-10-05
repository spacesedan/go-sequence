package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/spacesedan/go-sequence/internal/game"
)

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func main() {
	gs := game.NewGameService()
	deck := gs.GetDeck()
	board := gs.GetBoard()
	fmt.Println(deck)

    for i:= 0; i < 10; i++ {
        for j:= 0; j< 10; j++ {
            fmt.Println(board[i][j])
        }
    }
}

func _main() {

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

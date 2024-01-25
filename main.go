package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/theleeeo/docs-server/provider"
	"github.com/theleeeo/docs-server/server"
	"github.com/theleeeo/leolog"
)

func loadConfig() (*server.Config, error) {
	pollInterval, err := parseInterval(os.Getenv("POLL_INTERVAL"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse poll interval: %w", err)
	}

	return &server.Config{
		Owner:        os.Getenv("OWNER"),
		Repo:         os.Getenv("REPO"),
		CompanyName:  os.Getenv("COMPANY_NAME"),
		PathPrefix:   os.Getenv("PATH_PREFIX"),
		FileSuffix:   os.Getenv("FILE_SUFFIX"),
		PollInterval: pollInterval,
	}, nil
}

func parseInterval(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}

	return d, nil
}

func main() {
	logger := slog.New(leolog.NewHandler(nil))
	slog.SetDefault(logger)

	serverConfig, err := loadConfig()
	if err != nil {
		panic(err)
	}

	ghClient := provider.NewGithub(serverConfig.Owner, serverConfig.Repo)

	s, err := server.New(serverConfig, ghClient)
	if err != nil {
		panic(err)
	}

	app := NewApp(s)

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "localhost:4444"
	}

	var termChan = make(chan os.Signal, 1)
	var appErrChan = make(chan error, 1)
	var serverErrChan = make(chan error, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	wg := &sync.WaitGroup{}

	startApp(ctx, app, addr, wg, appErrChan)
	startServer(ctx, s, wg, serverErrChan)

	select {
	case <-termChan:
		slog.Info("shutting down...")

		go func() {
			<-termChan
			slog.Info("force killing...")
			os.Exit(1)
		}()

		cancel()

	case err := <-appErrChan:
		slog.Error("the app encountered an error", "error", err.Error())
		cancel()
	case err := <-serverErrChan:
		slog.Error("the server encountered an error", "error", err.Error())
		cancel()
	}

	wg.Wait()
}

func startApp(ctx context.Context, app *fiber.App, addr string, wg *sync.WaitGroup, errChan chan<- error) {
	go func() {
		wg.Add(1)

		slog.Info("starting app", "addr", addr)
		if err := app.Listen(addr); err != nil {
			errChan <- err
		}
	}()

	go func() {
		// The waitgroup is freed here because the app.Listen() can finish before the app.Shutdown() is completed.
		defer wg.Done()

		<-ctx.Done()

		if err := app.Shutdown(); err != nil {
			slog.Error("failed to shutdown app", "error", err)
		}
	}()
}

func startServer(ctx context.Context, s *server.Server, wg *sync.WaitGroup, errChan chan<- error) {
	go func() {
		wg.Add(1)
		defer wg.Done()

		slog.Info("starting server")
		if err := s.Run(ctx); err != nil {
			errChan <- err
		}
	}()
}

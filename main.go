package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/theleeeo/docs-server/app"
	"github.com/theleeeo/docs-server/provider"
	"github.com/theleeeo/docs-server/server"
	"github.com/theleeeo/leolog"
)

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
	logger := slog.New(leolog.NewHandler(&slog.HandlerOptions{Level: slog.LevelDebug}))
	log.SetLevel(log.LevelDebug)
	slog.SetDefault(logger)

	cfg, err := loadConfig()
	if err != nil {
		color.Red("failed to load config: %s", err)
		return
	}

	ghClient, err := provider.NewGithub(cfg.Provider.Github.Owner, cfg.Provider.Github.Repo, cfg.Provider.Github.Token)
	if err != nil {
		color.Red("failed to create github provider: %s", err)
		return
	}

	serverConfig := &server.Config{
		PathPrefix: cfg.Server.PathPrefix,
		FileSuffix: cfg.Server.FileSuffix,
	}
	interval, err := parseInterval(cfg.Server.PollInterval)
	if err != nil {
		color.Red("failed to parse poll interval: %s", err)
		return
	}
	serverConfig.PollInterval = interval

	s, err := server.New(serverConfig, ghClient)
	if err != nil {
		color.Red("failed to create server: %s", err)
		return
	}

	app, err := app.New(&app.Config{
		CompanyName: cfg.Design.CompanyName,
		CompanyLogo: cfg.Design.CompanyLogo,
	}, s)
	if err != nil {
		color.Red("failed to create app: %s", err)
		return
	}

	var termChan = make(chan os.Signal, 1)
	var appErrChan = make(chan error, 1)
	var serverErrChan = make(chan error, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	wg := &sync.WaitGroup{}

	startApp(ctx, app, cfg.Addr, wg, appErrChan)
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

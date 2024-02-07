package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/color"
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

func parseLogLevel(s string) (slog.Level, error) {
	if s == "" {
		return slog.LevelInfo, nil
	}

	s = strings.ToLower(s)
	switch s {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelDebug, fmt.Errorf("invalid log level: %s", s)
	}
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		color.Red("failed to load config: %s", err)
		return
	}

	logLevel, err := parseLogLevel(cfg.LogLevel)
	if err != nil {
		color.Red("failed to parse log level: %s", err)
		return
	}

	logger := slog.New(leolog.NewHandler(&slog.HandlerOptions{Level: logLevel}))
	log.SetLevel(log.LevelDebug)
	slog.SetDefault(logger)

	ghClient, err := setupProvider(cfg)
	if err != nil {
		color.Red("failed to setup provider: %s", err)
		return
	}

	s, err := setupServer(cfg, ghClient)
	if err != nil {
		color.Red("failed to setup server: %s", err)
		return
	}

	app, err := app.New(&app.Config{
		Address:     cfg.App.Address,
		HeaderTitle: cfg.Design.HeaderTitle,
		HeaderImage: cfg.Design.HeaderImage,
		Favicon:     cfg.Design.Favicon,
		PathPrefix:  cfg.App.PathPrefix,
	}, s)
	if err != nil {
		color.Red("failed to create app: %s", err)
		return
	}

	var termChan = make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	var appErrChan = make(chan error, 1)
	var serverErrChan = make(chan error, 1)

	ctx, cancel := context.WithCancel(context.Background())

	wg := &sync.WaitGroup{}

	go func() {
		wg.Add(1)
		defer wg.Done()

		if err := app.Run(ctx); err != nil {
			appErrChan <- err
		}
	}()

	go func() {
		wg.Add(1)
		defer wg.Done()

		if err := s.Run(ctx); err != nil {
			serverErrChan <- err
		}
	}()

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

func setupProvider(cfg *Config) (p server.Provider, err error) {
	// This pattern is possible extensibility, even if it's not used atm.
	if cfg.Provider.Github != nil {
		ghConfig := &provider.GithubConfig{
			Owner:      cfg.Provider.Github.Owner,
			Repo:       cfg.Provider.Github.Repo,
			PathPrefix: cfg.Provider.Github.PathPrefix,
			FileSuffix: cfg.Provider.Github.FileSuffix,
			MaxTags:    cfg.Provider.Github.MaxTags,
			AuthToken:  cfg.Provider.Github.AuthToken,
		}

		p, err = provider.NewGithub(ghConfig)
		if err != nil {
			return nil, err
		}
	}

	if p == nil {
		return nil, fmt.Errorf("no provider configured")
	}

	return p, nil
}

func setupServer(cfg *Config, p server.Provider) (s *server.Server, err error) {
	interval, err := parseInterval(cfg.Server.PollInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to parse poll interval: %w", err)
	}

	serverConfig := &server.Config{
		PollInterval: interval,
	}

	s, err = server.New(serverConfig, p)
	if err != nil {
		return nil, err
	}

	return s, nil
}

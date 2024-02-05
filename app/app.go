package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/template/html/v2"
	"github.com/theleeeo/docs-server/server"
)

const (
	publicFilesPath = "public"
	defaultAddress  = "localhost:4444"
	defaultLogo     = "/favicon.ico"
)

type App struct {
	fiberApp *fiber.App

	cfg  *Config
	serv *server.Server

	files struct {
		script string
		style  string
	}
}

func New(cfg *Config, s *server.Server) (*App, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	a := &App{
		fiberApp: fiber.New(fiber.Config{
			Views:   html.New("./views", ".html"),
			GETOnly: true,
		}),

		cfg:  cfg,
		serv: s,
	}

	// Load header logo
	logo, err := getHeaderLogo(a.cfg.HeaderLogo)
	if err != nil {
		return nil, err
	}

	a.fiberApp.Use(favicon.New(favicon.Config{
		Data: logo,
	}))

	if err := a.loadScript(); err != nil {
		return nil, err
	}

	registerHandlers(a)

	return a, nil
}

func (a *App) Run(ctx context.Context) error {
	errChan := make(chan error)
	go func() {
		slog.Info("starting app", "addr", a.cfg.Address)
		if err := a.fiberApp.Listen(a.cfg.Address); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		if err := a.fiberApp.Shutdown(); err != nil {
			slog.Error("failed to shutdown app", "error", err)
		}
		return nil
	case err := <-errChan:
		return err
	}
}

func registerHandlers(a *App) {
	a.fiberApp.Get(fmt.Sprint(a.cfg.PathPrefix, "/"), a.getIndexHandler)

	a.fiberApp.Get(fmt.Sprint(a.cfg.PathPrefix, "/script.js"), a.getScriptHandler)
	a.fiberApp.Get(fmt.Sprint(a.cfg.PathPrefix, "/style.css"), a.getStyleHandler)

	a.fiberApp.Get(fmt.Sprint(a.cfg.PathPrefix, "/:version/:role"), a.renderDocHandler)

	a.fiberApp.Get(fmt.Sprint(a.cfg.PathPrefix, "/versions"), a.getVersionsHandler)
	a.fiberApp.Get(fmt.Sprint(a.cfg.PathPrefix, "/version/:version/roles"), a.getRolesHandler)
}

func validateConfig(cfg *Config) error {
	if cfg.Address == "" {
		slog.Info("no address set, using default", "default", defaultAddress)
		cfg.Address = defaultAddress
	}

	if cfg.RootUrl == "" {
		return fmt.Errorf("root url is required")
	}

	rootUrl, err := url.Parse(cfg.RootUrl)
	if err != nil {
		return fmt.Errorf("invalid root url: %w", err)
	}
	rootUrl.Scheme = "https"
	if cfg.DocsUseHttp {
		rootUrl.Scheme = "http"
	}

	cfg.RootUrl = rootUrl.String()

	if cfg.HeaderLogo == "" {
		slog.Info("no header logo set, using default", "default", defaultLogo)
		cfg.HeaderLogo = defaultLogo
	}

	if !strings.HasPrefix(cfg.PathPrefix, "/") {
		cfg.PathPrefix = fmt.Sprint("/", cfg.PathPrefix)
	}

	return nil
}

func getHeaderLogo(location string) ([]byte, error) {
	// Check if HeaderLogo is a file
	slog.Debug("checking if HeaderLogo is a file")
	// Prepend "public/" to the path because that's where the static files are
	file, err := os.Open(filepath.Join(publicFilesPath, location))
	if err == nil {
		slog.Info("header logo loaded from file")
		defer file.Close()
		return io.ReadAll(file)
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	slog.Debug("headerLogo is not a file")

	// If HeaderLogo is not a file, assume it's a URL and make an HTTP request
	slog.Debug("checking if HeaderLogo is a URL")
	resp, err := http.Get(location)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	slog.Info("header logo loaded from URL")
	return io.ReadAll(resp.Body)
}

func (a *App) loadScript() error {
	b, err := os.ReadFile(filepath.Join(publicFilesPath, "script.js"))
	if err != nil {
		return err
	}

	t, err := template.New("script").Parse(string(b))
	if err != nil {
		return err
	}

	fmt.Println(a.cfg.PathPrefix)

	vars := map[string]string{
		"PathPrefix": a.cfg.PathPrefix,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, vars); err != nil {
		return err
	}

	a.files.script = buf.String()
	fmt.Println(a.files.script)

	return nil
}

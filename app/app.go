package app

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/theleeeo/docs-server/server"
)

const (
	publicFilesPath = "public"
	defaultAddress  = "localhost:4444"
)

type App struct {
	fiberApp *fiber.App

	cfg  *Config
	serv *server.Server

	files struct {
		headerImage image
		favicon     image
		script      string
		style       string
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

	slog.Info("loading favicon")
	icon, err := loadImage(a.cfg.Favicon)
	if err != nil {
		return nil, err
	}
	a.files.favicon = *icon

	slog.Info("loading header image")
	headerImage, err := loadImage(a.cfg.HeaderImage)
	if err != nil {
		return nil, err
	}
	a.files.headerImage = *headerImage

	if err := a.loadScript(); err != nil {
		return nil, err
	}

	if err := a.loadStyle(); err != nil {
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
	a.fiberApp.Get(fmt.Sprint(a.cfg.PathPrefix, "/header-image"), a.getHeaderImageHandler)
	a.fiberApp.Get(fmt.Sprint(a.cfg.PathPrefix, "/favicon.ico"), a.getFaviconHandler)

	a.fiberApp.Get(fmt.Sprint(a.cfg.PathPrefix, "/"), a.getIndexHandler)

	a.fiberApp.Get(fmt.Sprint(a.cfg.PathPrefix, "/script.js"), a.getScriptHandler)
	a.fiberApp.Get(fmt.Sprint(a.cfg.PathPrefix, "/style.css"), a.getStyleHandler)

	a.fiberApp.Get(fmt.Sprint(a.cfg.PathPrefix, "/:version/:role"), a.renderDocHandler)

	a.fiberApp.Get(fmt.Sprint(a.cfg.PathPrefix, "/versions"), a.getVersionsHandler)
	a.fiberApp.Get(fmt.Sprint(a.cfg.PathPrefix, "/version/:version/roles"), a.getRolesHandler)

	a.fiberApp.Get(fmt.Sprint(a.cfg.PathPrefix, "/proxy/:version/:file"), a.proxyHandler)
}

func validateConfig(cfg *Config) error {
	if cfg.Address == "" {
		slog.Info("no address set, using default", "default", defaultAddress)
		cfg.Address = defaultAddress
	}

	if cfg.Favicon == "" {
		return fmt.Errorf("favicon is required")
	}

	if cfg.HeaderImage == "" {
		slog.Info("no header image set, using favicon", "favicon", cfg.Favicon)
		cfg.HeaderImage = cfg.Favicon
	}

	if cfg.PathPrefix != "" && !strings.HasPrefix(cfg.PathPrefix, "/") {
		cfg.PathPrefix = fmt.Sprint("/", cfg.PathPrefix)
	}

	if cfg.HeaderColor == "" {
		cfg.HeaderColor = "none"
	}

	return nil
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

	vars := map[string]string{
		"PathPrefix": a.cfg.PathPrefix,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, vars); err != nil {
		return err
	}

	a.files.script = buf.String()

	return nil
}

func (a *App) loadStyle() error {
	b, err := os.ReadFile(filepath.Join(publicFilesPath, "style.css"))
	if err != nil {
		return err
	}

	t, err := template.New("style").Parse(string(b))
	if err != nil {
		return err
	}

	vars := map[string]string{
		"HeaderColor": a.cfg.HeaderColor,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, vars); err != nil {
		return err
	}

	a.files.style = buf.String()

	return nil
}

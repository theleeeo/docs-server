package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/theleeeo/docs-server/server"
)

const (
	// staticPath is the path to the static files
	staticFilesPath = "./public"
	defaultAddress  = "localhost:4444"
	defaultLogo     = "/favicon.ico"
)

func getHeaderLogo(location string) ([]byte, error) {
	// Check if HeaderLogo is a file
	slog.Debug("checking if HeaderLogo is a file")
	// Prepend "public/" to the path because that's where the static files are
	file, err := os.Open(fmt.Sprint(staticFilesPath, "/", location))
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

func registerHandlers(a *App) {
	a.fiberApp.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.Send(a.logo)
	})

	// Serve static files
	a.fiberApp.Static("/", staticFilesPath)

	a.fiberApp.Get("/", a.getIndexHandler)
	a.fiberApp.Get("/:version/:role", a.renderDocHandler)

	a.fiberApp.Get("/versions", a.getVersionsHandler)
	a.fiberApp.Get("/version/:version/roles", a.getRolesHandler)
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

	return nil
}

type App struct {
	fiberApp *fiber.App

	cfg  *Config
	serv *server.Server

	logo []byte
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
	a.logo = logo

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

func (a *App) getIndexHandler(c *fiber.Ctx) error {
	return c.Render("version-select", fiber.Map{
		"HeaderTitle": a.cfg.HeaderTitle,
		"HeaderLogo":  a.cfg.HeaderLogo,
	}, "layouts/main")
}

func (a *App) renderDocHandler(c *fiber.Ctx) error {
	version := c.Params("version")
	role := c.Params("role")

	return c.Render("doc", fiber.Map{
		"RootUrl":     a.cfg.RootUrl,
		"Path":        fmt.Sprintf("%s%s%s", a.serv.Path(), role, a.serv.FileSuffix()),
		"Ref":         version,
		"HeaderTitle": a.cfg.HeaderTitle,
		"HeaderLogo":  a.cfg.HeaderLogo,
	}, "layouts/main")
}

func (a *App) getVersionsHandler(c *fiber.Ctx) error {
	return c.JSON(a.serv.GetVersions())
}

func (a *App) getRolesHandler(c *fiber.Ctx) error {
	version := c.Params("version")

	doc := a.serv.GetVersion(version)
	if doc == nil {
		return c.Status(fiber.StatusNotFound).SendString("404 Version Not Found")
	}

	return c.JSON(doc.Files)
}

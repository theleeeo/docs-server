package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/theleeeo/docs-server/server"
)

const (
	publicFilesPath = "public"
	defaultAddress  = "localhost:4444"
)

type App struct {
	httpServer *http.Server
	templates  map[string]*template.Template

	cfg  *Config
	serv *server.Server

	files struct {
		headerImage *image
		favicon     *image
		script      string
		style       string
	}
}

func New(cfg *Config, s *server.Server) (*App, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	a := &App{
		templates: make(map[string]*template.Template),
		cfg:       cfg,
		serv:      s,
	}

	if cfg.Favicon != "" {
		slog.Info("loading favicon")
		icon, err := loadImage(a.cfg.Favicon)
		if err != nil {
			return nil, err
		}
		a.files.favicon = icon
	}

	if cfg.HeaderImage != "" {
		slog.Info("loading header image")
		headerImage, err := loadImage(a.cfg.HeaderImage)
		if err != nil {
			return nil, err
		}
		a.files.headerImage = headerImage
	}

	if err := a.loadScript(); err != nil {
		return nil, err
	}

	if err := a.loadStyle(); err != nil {
		return nil, err
	}

	if err := a.loadTemplates(); err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	registerHandlers(mux, a)

	a.httpServer = &http.Server{
		Addr:    a.cfg.Address,
		Handler: mux,
	}

	return a, nil
}

func (a *App) Run(ctx context.Context) error {
	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)

		slog.Info("starting app", "addr", a.cfg.Address)
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("failed to shutdown app", "error", err)
		}

		if err, ok := <-errChan; ok && err != nil {
			return err
		}

		return nil
	case err, ok := <-errChan:
		if !ok {
			return nil
		}

		return err
	}
}

func registerHandlers(mux *http.ServeMux, a *App) {
	if a.cfg.PathPrefix != "" {
		mux.HandleFunc("GET "+a.cfg.PathPrefix, a.redirectToRootHandler)
	}

	mux.HandleFunc("GET "+a.route("/header-image"), a.getHeaderImageHandler)
	mux.HandleFunc("GET "+a.route("/favicon.ico"), a.getFaviconHandler)
	mux.HandleFunc("GET "+a.rootRoute(), a.getIndexHandler)
	mux.HandleFunc("GET "+a.route("/script.js"), a.getScriptHandler)
	mux.HandleFunc("GET "+a.route("/style.css"), a.getStyleHandler)
	mux.HandleFunc("GET "+a.route("/versions"), a.getVersionsHandler)
	mux.HandleFunc("GET "+a.route("/version/{version}/roles"), a.getRolesHandler)
	mux.HandleFunc("GET "+a.route("/{version}/{role...}"), a.renderDocHandler)
	mux.HandleFunc("GET "+a.route("/proxy/{version}/{file...}"), a.proxyHandler)
}

func validateConfig(cfg *Config) error {
	if cfg.Address == "" {
		slog.Info("no address set, using default", "default", defaultAddress)
		cfg.Address = defaultAddress
	}

	if cfg.Favicon == "" {
		slog.Info("no favicon set")
	}

	if cfg.HeaderImage == "" {
		slog.Info("no header image set")
	}

	if cfg.PathPrefix != "" && !strings.HasPrefix(cfg.PathPrefix, "/") {
		cfg.PathPrefix = fmt.Sprint("/", cfg.PathPrefix)
	}
	cfg.PathPrefix = strings.TrimRight(cfg.PathPrefix, "/")

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

func (a *App) loadTemplates() error {
	pages := map[string]string{
		"version-select": filepath.Join("views", "version-select.html"),
		"doc":            filepath.Join("views", "doc.html"),
	}

	for name, page := range pages {
		t, err := template.ParseFiles(
			filepath.Join("views", "layouts", "main.html"),
			filepath.Join("views", "partials", "header.html"),
			page,
		)
		if err != nil {
			return err
		}

		a.templates[name] = t
	}

	return nil
}

func (a *App) route(path string) string {
	if a.cfg.PathPrefix == "" {
		return path
	}

	return a.cfg.PathPrefix + path
}

func (a *App) rootRoute() string {
	if a.cfg.PathPrefix == "" {
		return "/{$}"
	}

	return a.cfg.PathPrefix + "/{$}"
}

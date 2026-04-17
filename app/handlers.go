package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/theleeeo/docs-server/server"
)

func (a *App) getHeaderImageHandler(w http.ResponseWriter, r *http.Request) {
	if a.files.headerImage == nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", a.files.headerImage.contentType)
	_, _ = w.Write(a.files.headerImage.data)
}

func (a *App) getFaviconHandler(w http.ResponseWriter, r *http.Request) {
	if a.files.favicon == nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", a.files.favicon.contentType)
	_, _ = w.Write(a.files.favicon.data)
}

func (a *App) getScriptHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	_, _ = io.WriteString(w, a.files.script)
}

func (a *App) getStyleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	_, _ = io.WriteString(w, a.files.style)
}

func (a *App) getIndexHandler(w http.ResponseWriter, r *http.Request) {
	a.render(w, "version-select", map[string]any{
		"HeaderTitle": a.cfg.HeaderTitle,
		"Favicon":     a.cfg.Favicon,
		"PathPrefix":  a.cfg.PathPrefix,
		"HasTitle":    a.cfg.HeaderTitle != "",
		"HasImage":    a.cfg.HeaderImage != "",
	})
}

func (a *App) renderDocHandler(w http.ResponseWriter, r *http.Request) {
	version := r.PathValue("version")
	role := r.PathValue("role")

	var path string
	if a.serv.ProxyEnabled() {
		path = fmt.Sprint(a.cfg.PathPrefix, "/proxy/", version, "/", role)
	} else {
		path = a.serv.Path(version, role)
	}

	a.render(w, "doc", map[string]any{
		"HeaderTitle": a.cfg.HeaderTitle,
		"Favicon":     a.cfg.Favicon,
		"PathPrefix":  a.cfg.PathPrefix,
		"HasTitle":    a.cfg.HeaderTitle != "",
		"HasImage":    a.cfg.HeaderImage != "",
		"Path":        path,
	})
}

func (a *App) getVersionsHandler(w http.ResponseWriter, r *http.Request) {
	a.writeJSON(w, a.serv.GetVersions())
}

func (a *App) getRolesHandler(w http.ResponseWriter, r *http.Request) {
	version := r.PathValue("version")

	doc := a.serv.GetVersion(version)
	if doc == nil {
		http.Error(w, "404 Version Not Found", http.StatusNotFound)
		return
	}

	a.writeJSON(w, doc.Files)
}

func (a *App) proxyHandler(w http.ResponseWriter, r *http.Request) {
	if !a.serv.ProxyEnabled() {
		http.NotFound(w, r)
		return
	}

	version := r.PathValue("version")
	file := r.PathValue("file")

	data, err := a.serv.GetFile(r.Context(), version, file)
	if err != nil {
		if errors.Is(err, server.ErrNotFound) {
			http.NotFound(w, r)
			return
		}

		slog.Error("failed to get file from proxy", "error", err)
		http.Error(w, "An error occurred, please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", http.DetectContentType(data))
	_, _ = w.Write(data)
}

func (a *App) redirectToRootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, a.cfg.PathPrefix+"/", http.StatusMovedPermanently)
}

func (a *App) render(w http.ResponseWriter, name string, data map[string]any) {
	t, ok := a.templates[name]
	if !ok {
		slog.Error("template not found", "name", name)
		http.Error(w, "An error occurred, please try again later.", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout", data); err != nil {
		slog.Error("failed to render template", "name", name, "error", err)
		http.Error(w, "An error occurred, please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

func (a *App) writeJSON(w http.ResponseWriter, data any) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		slog.Error("failed to encode json", "error", err)
		http.Error(w, "An error occurred, please try again later.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(buf.Bytes())
}

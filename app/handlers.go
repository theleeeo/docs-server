package app

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/theleeeo/docs-server/server"
)

func (a *App) getHeaderImageHandler(c *fiber.Ctx) error {
	c.Set(fiber.HeaderContentType, a.files.headerImage.contentType)
	return c.Send(a.files.headerImage.data)
}

func (a *App) getFaviconHandler(c *fiber.Ctx) error {
	c.Set(fiber.HeaderContentType, a.files.favicon.contentType)
	return c.Send(a.files.favicon.data)
}

func (a *App) getScriptHandler(c *fiber.Ctx) error {
	c.Set(fiber.HeaderContentType, "application/javascript")
	return c.SendString(string(a.files.script))
}

func (a *App) getStyleHandler(c *fiber.Ctx) error {
	c.Set(fiber.HeaderContentType, "text/css")
	return c.SendString(string(a.files.style))
}

func (a *App) getIndexHandler(c *fiber.Ctx) error {
	return c.Render("version-select", fiber.Map{
		"HeaderTitle": a.cfg.HeaderTitle,
		"Favicon":     a.cfg.Favicon,
		"PathPrefix":  a.cfg.PathPrefix,
	}, "layouts/main")
}

func (a *App) renderDocHandler(c *fiber.Ctx) error {
	version := c.Params("version")
	role := c.Params("role")

	var path string
	if a.serv.ProxyEnabled() {
		path = fmt.Sprint(a.cfg.PathPrefix, "/proxy/", version, "/", role)
	} else {
		path = a.serv.Path(version, role)
	}

	return c.Render("doc", fiber.Map{
		"Path":        path,
		"HeaderTitle": a.cfg.HeaderTitle,
		"Favicon":     a.cfg.Favicon,
		"PathPrefix":  a.cfg.PathPrefix,
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

func (a *App) proxyHandler(c *fiber.Ctx) error {
	if !a.serv.ProxyEnabled() {
		return c.Status(fiber.StatusNotFound).SendString("404 Not Found")
	}

	version := c.Params("version")
	file := c.Params("file")

	data, err := a.serv.GetFile(c.Context(), version, file)
	if err != nil {
		if errors.Is(err, server.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).SendString("404 Not Found")
		}

		slog.Error("failed to get file from proxy", "error", err)
		return c.Status(fiber.StatusInternalServerError).SendString("An error occurred, please try again later.")
	}

	return c.Send(data)
}

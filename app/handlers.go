package app

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
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
	return c.SendString(string(a.files.script))
}

func (a *App) getStyleHandler(c *fiber.Ctx) error {
	return c.SendFile(fmt.Sprint(publicFilesPath, "/style.css"))
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

	return c.Render("doc", fiber.Map{
		"Path":        a.serv.Path(version, role),
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

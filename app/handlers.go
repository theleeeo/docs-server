package app

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func (a App) getScriptHandler(c *fiber.Ctx) error {
	return c.SendFile(fmt.Sprint(staticFilesPath, "/script.js"))
}

func (a App) getStyleHandler(c *fiber.Ctx) error {
	return c.SendFile(fmt.Sprint(staticFilesPath, "/style.css"))
}

func (a *App) getIndexHandler(c *fiber.Ctx) error {
	return c.Render("version-select", fiber.Map{
		"HeaderTitle": a.cfg.HeaderTitle,
		"HeaderLogo":  a.cfg.HeaderLogo,
		"Favicon":     a.cfg.Favicon,
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
		"Favicon":     a.cfg.Favicon,
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

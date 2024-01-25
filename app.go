package main

import (
	"fmt"
	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/theleeeo/docs-server/server"
)

func NewApp(s *server.Server) *fiber.App {
	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views:   engine,
		GETOnly: true,
	})

	// Serve static files
	app.Static("/docs", "./docs")
	app.Static("/", "./public")

	app.Get("/", createGetIndexHandler(s))
	app.Get("/:version/:role", createRenderDocHandler(s))

	app.Get("/versions", createGetVersionsHandler(s))
	app.Get("/version/:version/roles", createGetRolesHandler(s))

	return app
}

func createGetIndexHandler(s *server.Server) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return c.Render("version-select", fiber.Map{
			"CompanyName": s.CompanyName(),
		}, "layouts/main")
	}
}

func createRenderDocHandler(s *server.Server) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		version := c.Params("version")
		role := c.Params("role")

		doc := s.GetVersion(version)
		if doc == nil {
			return c.Status(fiber.StatusNotFound).SendString("404 Version Not Found")
		}

		ok := slices.Contains(doc.Files, role)
		if !ok {
			return c.Status(fiber.StatusNotFound).SendString("404 Role Not Found")
		}

		return c.Render("doc", fiber.Map{
			"Owner":       s.Owner(),
			"Repo":        s.Repo(),
			"Path":        fmt.Sprintf("%s%s%s", s.Path(), role, s.FileSuffix()),
			"Ref":         version,
			"CompanyName": s.CompanyName(),
		}, "layouts/main")
	}
}

func createGetVersionsHandler(s *server.Server) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return c.JSON(s.GetVersions())
	}
}

func createGetRolesHandler(s *server.Server) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		version := c.Params("version")

		doc := s.GetVersion(version)
		if doc == nil {
			return c.Status(fiber.StatusNotFound).SendString("404 Version Not Found")
		}

		return c.JSON(doc.Files)
	}
}

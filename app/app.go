package app

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/theleeeo/docs-server/server"
)

const (
	// staticPath is the path to the static files
	staticFilesPath = "./public"
)

func New(cfg *Config, s *server.Server) (*fiber.App, error) {
	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views:   engine,
		GETOnly: true,
	})

	// Load company logo
	logo, err := getCompanyLogo(cfg)
	if err != nil {
		return nil, err
	}

	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.Send(logo)
	})

	// Serve static files
	app.Static("/docs", "./docs")
	app.Static("/", staticFilesPath)

	app.Get("/", createGetIndexHandler(cfg, s))
	app.Get("/:version/:role", createRenderDocHandler(cfg, s))

	app.Get("/versions", createGetVersionsHandler(s))
	app.Get("/version/:version/roles", createGetRolesHandler(s))

	return app, nil
}

func getCompanyLogo(cfg *Config) ([]byte, error) {
	// Check if CompanyLogo is a file
	slog.Debug("Checking if CompanyLogo is a file")
	// Prepend "public/" to the path because that's where the static files are
	file, err := os.Open(fmt.Sprint(staticFilesPath, "/", cfg.CompanyLogo))
	if err == nil {
		slog.Info("Company logo loaded from file")
		defer file.Close()
		return io.ReadAll(file)
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	slog.Debug("CompanyLogo is not a file")

	// If CompanyLogo is not a file, assume it's a URL and make an HTTP request
	slog.Debug("Checking if CompanyLogo is a URL")
	resp, err := http.Get(cfg.CompanyLogo)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	slog.Info("Company logo loaded from URL")
	return io.ReadAll(resp.Body)
}

func createGetIndexHandler(cfg *Config, s *server.Server) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return c.Render("version-select", fiber.Map{
			"CompanyName": cfg.CompanyName,
			"CompanyLogo": cfg.CompanyLogo,
		}, "layouts/main")
	}
}

func createRenderDocHandler(cfg *Config, s *server.Server) func(c *fiber.Ctx) error {
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
			"Owner":       "LeoCorp", //s.Owner(),
			"Repo":        "LeoRepo", //s.Repo(),
			"Path":        fmt.Sprintf("%s%s%s", s.Path(), role, s.FileSuffix()),
			"Ref":         version,
			"CompanyName": cfg.CompanyName,
			"CompanyLogo": cfg.CompanyLogo,
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

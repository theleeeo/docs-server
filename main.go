package main

import (
	"os"
	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func main() {
	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// Serve static files
	app.Static("/docs", "./docs")
	app.Static("/", "./public")

	app.Get("/", getIndexfunc)
	app.Get("/:version/:role", getDoc)

	app.Get("/versions", getVersions)
	app.Get("/version/:version/roles", getRoles)

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "localhost:3000"
	}

	app.Listen(addr)
}

func getIndexfunc(c *fiber.Ctx) error {
	return c.Render("version-select", fiber.Map{}, "layouts/main")
}

func getDoc(c *fiber.Ctx) error {
	version := c.Params("version")
	role := c.Params("role")

	roles, ok := roles[version]
	if !ok {
		return c.Status(fiber.StatusNotFound).SendString("404 Version Not Found")
	}

	ok = slices.Contains(roles, role)
	if !ok {
		return c.Status(fiber.StatusNotFound).SendString("404 Role Not Found")
	}

	return c.Render("doc", fiber.Map{
		"Url": "/docs/" + version + "/" + role + ".swagger.json",
	}, "layouts/main")
}

var roles = map[string][]string{
	"1": {"SP", "INTERNAL", "OPERATOR"},
	"2": {"SP", "INTERNAL"},
}

func getRoles(c *fiber.Ctx) error {
	version := c.Params("version")
	roleList, ok := roles[version]
	if !ok {
		return c.Status(fiber.StatusNotFound).SendString("404 Version Not Found")
	}

	return c.JSON(roleList)
}

func getVersions(c *fiber.Ctx) error {
	// Fetch button data from backend
	versions := []string{"1", "2"}
	return c.JSON(versions)
}

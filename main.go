package main

import (
	"github.com/gofiber/fiber/v2"
)

func setupRoutes(app *fiber.App) {
	app.Get("/versions", getVersions)
	app.Get("/hello", hello)
	app.Get("/version/:version/roles", getRoles)
}

func main() {
	app := fiber.New()

	// Serve static files
	app.Static("/docs", "./docs")
	app.Static("/", "./public")

	setupRoutes(app)

	app.Listen(":3000")
}

func hello(c *fiber.Ctx) error {
	return c.SendString("Hello, World 👋!")
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

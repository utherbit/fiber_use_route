package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"main/fiber_use_route"
)

func main() {
	app := fiber.New()

	// you need to create a route manager
	furApp := fiber_use_route.NewManager()
	// Add Route Manager Middleware
	app.Use(furApp.GetUse())

	app.Use(func(ctx *fiber.Ctx) error {
		// Get information about the fiber router, inside the middleware
		route := fiber_use_route.GetRouteFromContext(ctx)

		if route == nil {
			ctx.Status(404)
			return ctx.SendString("route not found")
		}
		/*	console output when requested:

			[GET] /demo/path/v:vNum<int>
			 query: /demo/path/v2

			[GET] /demo/path/v1/*
			 query: /demo/path/v1

			[GET] /demo/file/:guid<uuid>
			 query: /demo/file/0c704004-d239-407b-95b8-09c029676bef

			[POST] /demo/user/:id<int>
			 query: /demo/user/42
		*/
		fmt.Printf("\n\n[%s] %s", route.Method, route.Path)
		fmt.Printf("\n query: %s", ctx.Path())
		return ctx.Next()
	})

	app.Get("/demo/path", demoHandler)
	app.Get("/demo/path/v1/*", demoHandler)
	app.Get("/demo/path/v:vNum<int>", demoHandler)
	app.Get("/demo/file/:guid<uuid>", demoHandler)
	app.Post("/demo/user/:id<int>", demoHandler)

	// Run InitFiberApp before fiber app.listen
	furApp.InitFiberApp(app)

	err := app.Listen(":3030")
	if err != nil {
		panic(err)
	}
}

func demoHandler(*fiber.Ctx) error { return nil }

# Fiber use route

## Gratitude

Hello. First of all, I want to thank the [fiber community](https://github.com/gofiber) for developing a great [framework fiber](https://github.com/gofiber/fiber).
For all the time I have been using the fiber framework, I have not been able to identify significant shortcomings, this is very pleasing, thanks for that.

## Issue

The only downside I've encountered is that it's not possible to get the path inside the middleware. I understand fiber routing and I understand why it is not possible in the concept of a fast fiber framework.

Also, this problem is described in one of the [fiber issues](https://github.com/gofiber/fiber/issues/2195)

## Solution

To solve this problem, I wrote this utility. It allows you to get *fiber.Route inside Middlewares

This solution, like a fiber, indexes all your routes at startup and searches them during middleware execution, regardless of whether you use a variable in the context. Be prepared for a slight increase in memory consumption and execution time

## Usage

In your project, install the fiber packages and utilities for using routes

```go get github.com/gofiber/fiber/v2```

```go get github.com/utherbit/fiber_use_route```

Then you can use the features of the utility according to the example.

```go
// To use it, you just need to call the function
furApp := fiber_use_route.NewManager()

// You should use the middleware generated by the utility to add the variable to the context
app.Use(furApp.GetUse())
// Or you can write your own middleware that uses the Find function for the route manager
app.Use(func(ctx *fiber.Ctx) error {
    findRoute, exist := furApp.Find(ctx.Method(), ctx.Path())
    if exist {
        ctx.Locals(fiber_use_route.LocalsKeyEndpointPath, &findRoute.Orig())
    }
    return ctx.Next()
})

// this middleware adds the found route to locals by the "endpointPath" key, which can be obtained from fiber_use_route.LocalsKeyEndpointPath
app.Use(func(ctx *fiber.Ctx) error {
    find := ctx.Locals(fiber_use_route.LocalsKeyEndpointPath)
	
    if find == nil {
        ctx.Status(404)
        return ctx.SendString("route not found")
    }
    route = find.(*fiber.Route)
	return nil
})

// The function in the utility does the same
app.Use(func(ctx *fiber.Ctx) error {
    // Get information about the fiber router, inside the middleware
    route := fiber_use_route.GetRouteFromContext(ctx)
    if route == nil {
        ctx.Status(404)
        return ctx.SendString("route not found")
    }
    return nil
}

 
// Next, make the manager index all of your methods before the server listens.
furApp.InitFiberApp(app)
```

## Example

```go
package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/utherbit/fiber_use_route"
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
```
## for the future

I do not consider the provided solution correct, it is a crutch. I suggest to use it only as a temporary solution that we have encountered. I hope that in version 3 of the fiber there will be a more correct way to get the path in the middleware

## License
Fiber_use_route is free and open source software licensed under the [MIT license](https://github.com/utherbit/fiber_use_route/blob/master/LICENSE).

the project uses a third party fiber library

[Fiber license](https://github.com/gofiber/fiber/blob/master/LICENSE)
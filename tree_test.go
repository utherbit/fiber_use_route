package fiber_use_route

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/valyala/fasthttp"
	"net/http/httptest"
	"os"
	"testing"
)

type testRoute struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}
type routeJSON struct {
	TestRoutes []testRoute `json:"test_routes"`
	GithubAPI  []testRoute `json:"github_api"`
}

var routesFixture routeJSON

func init() {
	dat, err := os.ReadFile("./testdata/testRoutes.json")
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(dat, &routesFixture); err != nil {
		panic(err)
	}
}

func registerDummyRoutes(app *fiber.App) {
	h := func(c *fiber.Ctx) error {
		return nil
	}
	for _, r := range routesFixture.GithubAPI {
		app.Add(r.Method, r.Path, h)
	}
}

func Test_Middleware_Endpoint_Path(t *testing.T) {

	app := fiber.New()
	ma := NewManager()

	app.Use(ma.GetUse())

	h := func(ctx *fiber.Ctx) error {
		findRoute := GetRouteFromContext(ctx)
		if findRoute == nil {
			t.Fatal("route not found in manager")
			return nil
		}
		expectedRoute := ctx.Route()

		utils.AssertEqual(t, findRoute.Path, expectedRoute.Path)
		utils.AssertEqual(t, findRoute.Method, expectedRoute.Method)
		return nil
	}

	for _, r := range routesFixture.GithubAPI {
		app.Add(r.Method, r.Path, h)
	}

	ma.InitFiberApp(app)

	appHandler := app.Handler()
	c := &fasthttp.RequestCtx{}

	for i := range routesFixture.TestRoutes {
		c.Request.Header.SetMethod(routesFixture.TestRoutes[i].Method)
		c.URI().SetPath(routesFixture.TestRoutes[i].Path)
		appHandler(c)
	}
}

func Test_Route_Match_UnescapedPath(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{UnescapePath: true})

	ma := NewManager()
	app.Use(ma.GetUse())

	app.Get("/test/:id<int>", func(c *fiber.Ctx) error {
		re := GetRouteFromContext(c)

		if re == nil {

			c.Status(fiber.StatusNotFound)
			return nil
		}
		c.Status(fiber.StatusOK)
		return nil
	})

	ma.InitFiberApp(app)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/test/123", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, fiber.StatusOK, resp.StatusCode, "Status code")

	// without special chars
	resp, err = app.Test(httptest.NewRequest(fiber.MethodGet, "/test/notFound", nil))
	utils.AssertEqual(t, nil, err, "app.Test(req)")
	utils.AssertEqual(t, fiber.StatusNotFound, resp.StatusCode, "Status code")

}

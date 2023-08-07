package fiber_use_route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"sort"
	"strings"
	"sync"
)

const LocalsKeyEndpointPath = "endpointPath"

func NewManager() *managerApp {
	return &managerApp{
		mutex: sync.Mutex{},
	}
}
func (app *managerApp) GetUse() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		findRoute, exist := app.Find(ctx.Method(), ctx.Path())
		if exist {
			ctx.Locals(LocalsKeyEndpointPath, &findRoute.orig)
		}
		return ctx.Next()
	}
}
func GetRouteFromContext(ctx *fiber.Ctx) *fiber.Route {
	find := ctx.Locals(LocalsKeyEndpointPath)
	if find != nil {
		return find.(*fiber.Route)
	}
	return nil
}

func (app *managerApp) InitFiberApp(fiberApp *fiber.App) *managerApp {
	app.app = fiberApp
	app.config = app.app.Config()
	app.stack = make([][]*route, len(app.config.RequestMethods))
	app.treeStack = make([]map[string][]*route, len(app.config.RequestMethods))

	for _, addRoute := range app.app.GetRoutes(true) {
		app.addRoute(addRoute)
	}
	app.buildTree()
	return app
}
func (app *managerApp) Find(method, path string) (*route, bool) {

	return app.find(app.methodInt(method), path)
	//methodInt := app.methodInt(method)
}
func (app *managerApp) find(methodInt int, path string) (*route, bool) {
	treePath := path[0:0]
	const maxDetectionPaths = 3
	if len(path) >= maxDetectionPaths {
		treePath = path[:maxDetectionPaths]
	}

	tree, ok := app.treeStack[methodInt][treePath]
	if !ok {
		tree = app.treeStack[methodInt][""]
	}
	lenTree := len(tree) - 1

	// Loop over the route stack starting from previous index
	indexRoute := -1
	for indexRoute < lenTree {
		// Increment findRoute index
		indexRoute++

		// Get *findRoute
		findRoute := tree[indexRoute]
		var match bool
		//var err error
		// skip for mounted apps (всегда false)
		//if findRoute.mount {
		//	continue
		//}

		// Check if it matches the request path
		params := [maxParams]string{}
		match = findRoute.match(path, path, &params)
		//match = findRoute.match(c.detectionPath, c.path, &c.values)
		if !match {
			// No match, next findRoute
			continue
		}
		// Pass findRoute reference and param values
		//c.findRoute = findRoute

		// Non use handler matched
		//if !c.matched && !findRoute.use {
		//	c.matched = true
		//}

		// Execute first handler of findRoute
		//indexHandler = 0
		//if len(findRoute.Handlers) > 0 {
		//	err = findRoute.Handlers[0](c)
		//}
		return findRoute, true // Stop scanning the stack
	}
	return nil, false
}

type managerApp struct {
	app       *fiber.App
	mutex     sync.Mutex
	stack     [][]*route
	treeStack []map[string][]*route
	config    fiber.Config

	routesRefreshed bool
	lastPos         uint
}

type route struct {
	orig fiber.Route

	pos       uint // Position in stack -> important for the sort of the matched routes
	star      bool // Path equals '*'
	root      bool // Path equals '/'
	methodInt int
	// Path data
	path        string      // Prettified path
	routeParser routeParser // Parameter parser
	Params      []string    `json:"params"` // Case sensitive param keys

	Method string `json:"method"` // HTTP method
	Path   string `json:"path"`   // Original registered route path
}

func (r *route) Orig() fiber.Route {
	return r.orig
}
func (app *managerApp) addRoute(orig fiber.Route) *route {
	method := utils.ToUpper(orig.Method)
	pathRaw := orig.Path
	// Check if the HTTP method is valid unless it's USE (метод уже проверен)
	//if method != "USE" && methodInt(app, method) == -1 {
	//	panic(fmt.Sprintf("add: invalid http method %s\n", method))
	//}

	// is mounted app (группы не важны)
	//isMount := group != nil && group.app != app
	// A newRoute requires atleast one ctx handler
	//if len(handlers) == 0 && !isMount { (handlers не нужны)
	//	panic(fmt.Sprintf("missing handler in newRoute: %s\n", pathRaw))
	//}
	app.routesRefreshed = true

	// Cannot have an empty path
	if pathRaw == "" {
		pathRaw = "/"
	}
	// Path always start with a '/'
	if pathRaw[0] != '/' {
		pathRaw = "/" + pathRaw
	}

	// Create a stripped path in-case sensitive / trailing slashes
	pathPretty := pathRaw
	// Case sensitive routing, all to lowercase
	if !app.config.CaseSensitive {
		pathPretty = utils.ToLower(pathPretty)
	}
	// Strict routing, remove trailing slashes
	if !app.config.StrictRouting && len(pathPretty) > 1 {
		pathPretty = utils.TrimRight(pathPretty, '/')
	}
	// Is layer a middleware?
	//isUse := method == methodUse (не может быть use)

	// Is path a direct wildcard?
	isStar := pathPretty == "/*"
	// Is path a root slash?
	isRoot := pathPretty == "/"
	// Parse path parameters
	parsedRaw := parseRoute(pathRaw)
	parsedPretty := parseRoute(pathPretty)

	newRoute := route{
		orig: orig,
		// Router booleans
		//use:   isUse,
		//mount: isMount,
		star:      isStar,
		root:      isRoot,
		methodInt: app.methodInt(method),

		// Path data
		path:        removeEscapeChar(pathPretty),
		routeParser: parsedPretty,
		Params:      parsedRaw.params,

		// Group data
		//group: group,

		// Public data
		Path:   pathRaw,
		Method: method,
		//Handlers: handlers,
	}

	newRoute.pos = app.lastPos
	app.lastPos++
	app.stack[newRoute.methodInt] = append(app.stack[newRoute.methodInt], &newRoute)

	return &newRoute
}

// HTTP methods and their unique INTs
func (app *managerApp) methodInt(s string) int {
	//configured := app.Config()
	// For better performance
	if len(app.config.RequestMethods) == 0 {
		// TODO: Use iota instead
		switch s {
		case fiber.MethodGet:
			return 0
		case fiber.MethodHead:
			return 1
		case fiber.MethodPost:
			return 2
		case fiber.MethodPut:
			return 3
		case fiber.MethodDelete:
			return 4
		case fiber.MethodConnect:
			return 5
		case fiber.MethodOptions:
			return 6
		case fiber.MethodTrace:
			return 7
		case fiber.MethodPatch:
			return 8
		default:
			return -1
		}
	}

	// For method customization
	for i, v := range app.config.RequestMethods {
		if s == v {
			return i
		}
	}

	return -1
}

// buildTree build the prefix tree from the previously registered routes
func (app *managerApp) buildTree() *managerApp {
	if !app.routesRefreshed {
		return app
	}

	// loop all the methods and stacks and create the prefix tree
	for m := range app.config.RequestMethods {
		tsMap := make(map[string][]*route)
		for _, route := range app.stack[m] {
			treePath := ""
			if len(route.routeParser.segs) > 0 && len(route.routeParser.segs[0].Const) >= 3 {
				treePath = route.routeParser.segs[0].Const[:3]
			}
			// create tree stack
			tsMap[treePath] = append(tsMap[treePath], route)
		}
		app.treeStack[m] = tsMap
	}

	// loop the methods and tree stacks and add global stack and sort everything
	for m := range app.config.RequestMethods {
		tsMap := app.treeStack[m]
		for treePart := range tsMap {
			if treePart != "" {
				// merge global tree routes in current tree stack
				tsMap[treePart] = uniqueRouteStack(append(tsMap[treePart], tsMap[""]...))
			}
			// sort tree slices with the positions
			slc := tsMap[treePart]
			sort.Slice(slc, func(i, j int) bool { return slc[i].pos < slc[j].pos })
		}
	}
	app.routesRefreshed = false

	return app
}

// uniqueRouteStack drop all not unique routes from the slice
func uniqueRouteStack(stack []*route) []*route {
	var unique []*route
	m := make(map[*route]int)
	for _, v := range stack {
		if _, ok := m[v]; !ok {
			// Unique key found. Record position and collect
			// in result.
			m[v] = len(unique)
			unique = append(unique, v)
		}
	}

	return unique
}

func (r *route) match(detectionPath, path string, params *[maxParams]string) bool {
	// root detectionPath check
	if r.root && detectionPath == "/" {
		return true
		// '*' wildcard matches any detectionPath
	} else if r.star {
		if len(path) > 1 {
			params[0] = path[1:]
		} else {
			params[0] = ""
		}
		return true
	}
	// Does this route have parameters
	if len(r.Params) > 0 {
		// Match params// todo убрать всегда false
		if match := r.routeParser.getMatch(detectionPath, path, params, false); match {
			// Get params from the path detectionPath
			return match
		}
	}
	// Is this route a Middleware?
	if false { // todo убрать всегда false
		// Single slash will match or detectionPath prefix
		if r.root || strings.HasPrefix(detectionPath, r.path) {
			return true
		}
		// Check for a simple detectionPath match
	} else if len(r.path) == len(detectionPath) && r.path == detectionPath {
		return true
	}
	// No match
	return false
}

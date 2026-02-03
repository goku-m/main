# Route Flow (Gateway → Agrifolio → Response)

This file walks the request lifecycle from the entry point (`cmd/apps/main.go`) to the final HTTP response. It also highlights why a "route not found" can happen in the current gateway setup.

## 1) Entry Point Bootstraps the Server

**File:** `cmd/apps/main.go`

1. Loads configuration (`config.LoadConfig`).
2. Initializes logger and DB migration (non-local).
3. Creates `server.Server` (`server.New`).
4. Builds the Agrifolio module via `agrifolio.Module(srv)`.
5. Creates the gateway router with `gateway.New(agrifolioModule)`.
6. Calls `srv.SetupHTTPServer(r)` and starts the HTTP server.

At this point, the HTTP server is listening, and all routing is delegated to the gateway router.

---

## 2) Gateway Router Mounts Each Module

**File:** `internal/gateway/module.go`

`gateway.New(...)` creates an `echo.Echo` instance and mounts each module:

- Each module has `Name`, `Prefix`, and `Router` (http.Handler).
- The gateway normalizes the prefix (e.g., `/agrifolio`).
- It **strips the prefix** before passing the request to the module router:
  - `http.StripPrefix(prefix, module.Router)`
- Two mount points are registered:
  - `gw.Any(prefix, ...)`
  - `gw.Any(prefix+"/*", ...)`

**Implication:**  
If the gateway prefix is `/agrifolio`, your app router receives paths **without** `/agrifolio`.

Example:  
Incoming request `/agrifolio/api/sites`  
→ Gateway strips `/agrifolio`  
→ App router sees `/api/sites`

---

## 3) Agrifolio Module Builds Handlers + Router

**File:** `apps/agrifolio/project.go`

`agrifolio.Module(srv)` does:

1. `repository.NewRepositories(srv)`
2. `service.NewServices(srv, repos)`
3. `api.NewHandlers(srv, services)`
4. `api.NewRouter(srv, handlers)`

It returns:

```
gateway.Module{
  Name:   "agrifolio",
  Prefix: "/agrifolio",
  Router: router,
}
```

---

## 4) Agrifolio Router Defines Routes and Middleware

**File:** `apps/agrifolio/api/router/router.go`

Main steps:

1. Builds middlewares via `middleware.NewMiddlewares(s)`.
2. Creates an Echo router.
3. Registers:
   - global middleware (CORS, auth context, tracing, rate limit, etc.)
   - system routes
   - page routes
   - API routes under `/api`

Key registrations:

```
registerSystemRoutes(router, h)
registerPagesRoutes(router, h, middlewares.Auth)

r := router.Group("/api")
registerSiteRoutes(r, h.Site, middlewares.Auth)
```

So the app-level routes are expected to be:

- `/api/...` for JSON APIs
- `/login`, `/`, `/create` for pages, etc.

---

## 5) Handler Processes the Request

**Example:** `apps/agrifolio/api/handler/site.go`

1. Echo matches the route.
2. Middleware may enforce auth (`auth.RequireAuthIP`).
3. Handler method executes (`CreateSite`, `GetSitesAPI`, etc.).
4. Calls into service layer:
   - `siteService.GetSites(...)` or `siteService.CreateSite(...)`
5. Service calls repository which hits DB.

---

## 6) Response Is Returned

Handlers return:

- JSON via `c.JSON(...)`
- HTML via `c.Render(...)`
- Errors via `echo.NewHTTPError(...)`

Echo writes the response back to the HTTP client.

---

## Why "Route Not Found" Happens Now

Given the gateway **prefix stripping** behavior:

### ✅ Correct
```
GET /agrifolio/api/sites
```

Gateway strips `/agrifolio` → app router sees `/api/sites`.

### ❌ Incorrect
```
GET /api/sites
```

Gateway expects `/agrifolio/...` and will not match `/api/sites`.

---

## Quick Checklist to Debug 404

1. Confirm you are hitting the correct prefix:
   - `/agrifolio/...`
2. Confirm app routes are registered under `/api`:
   - `/agrifolio/api/...`
3. Confirm the handler exists in `apps/agrifolio/api/router`.
4. Confirm `gateway.New(...)` is mounting the module.
5. Check whether auth middleware blocks access.

---

## Optional Improvements (if you want)

- Allow a default module mounted at `/` (no prefix)
- Add host-based routing (e.g., `agrifolio.localhost`)
- Add gateway logging to print route matches

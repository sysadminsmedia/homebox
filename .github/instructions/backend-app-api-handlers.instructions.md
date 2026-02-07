---
applyTo: '/backend/app/api/handlers/**/*'
---

# Backend API Handlers Instructions (`/backend/app/api/handlers/v1/`)

## Overview

API handlers are the HTTP layer that processes requests, calls services/repositories, and returns responses. Handlers use the adapter pattern from `internal/web/adapters` and require Swagger documentation for auto-generation.

## Architecture Flow

```
HTTP Request → chi Router → Middleware Chain → Handler (via adapters) → Service/Repo → Database
                                                      ↓
                                                HTTP Response (server.JSON)
```

## Directory Structure

```
backend/app/api/
├── routes.go                           # Route definitions with chi router
├── handlers/
│   └── v1/
│       ├── controller.go               # V1Controller struct and initialization
│       ├── partials.go                 # routeID, routeUUID helpers
│       ├── helpers.go                  # URL helpers
│       ├── query_params.go             # queryIntOrNegativeOne, queryBool, queryUUIDList
│       ├── v1_ctrl_items.go            # Item endpoints
│       ├── v1_ctrl_items_attachments.go# Attachment endpoints
│       ├── v1_ctrl_labels.go           # Label endpoints
│       ├── v1_ctrl_locations.go        # Location endpoints
│       ├── v1_ctrl_auth.go             # Authentication endpoints
│       ├── v1_ctrl_user.go             # User endpoints
│       ├── v1_ctrl_group.go            # Group endpoints
│       ├── v1_ctrl_maintenance.go      # Maintenance endpoints
│       └── ...
```

## Handler Structure

### V1Controller

All handlers are methods on `V1Controller`:

```go
type V1Controller struct {
    cookieSecure      bool
    repo              *repo.AllRepos       // Direct repo access
    svc               *services.AllServices // Service layer
    maxUploadSize     int64
    isDemo            bool
    allowRegistration bool
    bus               *eventbus.EventBus   // Event publishing
    url               string
    config            *config.Config
    oidcProvider      *providers.OIDCProvider
}

func NewControllerV1(svc *services.AllServices, repos *repo.AllRepos, bus *eventbus.EventBus, config *config.Config, options ...func(*V1Controller)) *V1Controller {
    ctrl := &V1Controller{
        repo:              repos,
        svc:               svc,
        allowRegistration: true,
        bus:               bus,
        config:            config,
    }

    for _, opt := range options {
        opt(ctrl)
    }

    return ctrl
}
```

### Swagger Documentation

**CRITICAL:** Every handler must have Swagger comments for API doc generation:

```go
// HandleLabelsGetAll godoc
//
//	@Summary	Get All Labels
//	@Tags		Labels
//	@Produce	json
//	@Success	200	{object}	[]repo.LabelOut
//	@Router		/v1/labels [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelsGetAll() errchain.HandlerFunc {
    // ...
}
```

**After modifying Swagger comments, ALWAYS run:**
```bash
task generate  # Regenerates Swagger docs and TypeScript types
```

## Adapter Pattern

Handlers use adapters from `internal/web/adapters/` that handle request decoding, UUID parsing, and response encoding. **This is the primary pattern used throughout the codebase.**

### Available Adapters

| Adapter              | Use Case                 | Signature                                                 |
|----------------------|--------------------------|-----------------------------------------------------------|
| `adapters.Command`   | No body, no query params | `func(r *http.Request) (T, error)`                        |
| `adapters.CommandID` | UUID from path, no body  | `func(r *http.Request, ID uuid.UUID) (T, error)`          |
| `adapters.Action`    | Body decoding            | `func(r *http.Request, body T) (Y, error)`                |
| `adapters.ActionID`  | UUID from path + body    | `func(r *http.Request, ID uuid.UUID, body T) (Y, error)`  |
| `adapters.Query`     | Query params decoding    | `func(r *http.Request, query T) (Y, error)`               |
| `adapters.QueryID`   | UUID from path + query   | `func(r *http.Request, ID uuid.UUID, query T) (Y, error)` |

### Services Context

All handlers use `services.NewContext(r.Context())` to create an authenticated context that extracts the user from middleware:

```go
auth := services.NewContext(r.Context())
// auth.UID  - Current user ID
// auth.GID  - Current user's group ID
// auth.User - Full user object
```

## Standard Handler Patterns

### GET - List All (using adapters.Command)

```go
// HandleLabelsGetAll godoc
//
//	@Summary	Get All Labels
//	@Tags		Labels
//	@Produce	json
//	@Success	200	{object}	[]repo.LabelOut
//	@Router		/v1/labels [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelsGetAll() errchain.HandlerFunc {
    fn := func(r *http.Request) ([]repo.LabelSummary, error) {
        auth := services.NewContext(r.Context())
        return ctrl.repo.Labels.GetAll(auth, auth.GID)
    }

    return adapters.Command(fn, http.StatusOK)
}
```

### GET - List with Query Params (using adapters.Query)

```go
// HandleLocationGetAll godoc
//
//	@Summary	Get All Locations
//	@Tags		Locations
//	@Produce	json
//	@Param		filterChildren	query		bool	false	"Filter locations with parents"
//	@Success	200				{object}	[]repo.LocationOutCount
//	@Router		/v1/locations [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLocationGetAll() errchain.HandlerFunc {
    fn := func(r *http.Request, q repo.LocationQuery) ([]repo.LocationOutCount, error) {
        auth := services.NewContext(r.Context())
        return ctrl.repo.Locations.GetAll(auth, auth.GID, q)
    }

    return adapters.Query(fn, http.StatusOK)
}
```

### GET - Single Item (using adapters.CommandID)

```go
// HandleItemGet godocs
//
//	@Summary	Get Item
//	@Tags		Items
//	@Produce	json
//	@Param		id	path		string	true	"Item ID"
//	@Success	200	{object}	repo.ItemOut
//	@Router		/v1/items/{id} [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemGet() errchain.HandlerFunc {
    fn := func(r *http.Request, ID uuid.UUID) (repo.ItemOut, error) {
        auth := services.NewContext(r.Context())
        return ctrl.repo.Items.GetOneByGroup(auth, auth.GID, ID)
    }

    return adapters.CommandID("id", fn, http.StatusOK)
}
```

### POST - Create (using adapters.Action)

```go
// HandleItemsCreate godoc
//
//	@Summary	Create Item
//	@Tags		Items
//	@Produce	json
//	@Param		payload	body		repo.ItemCreate	true	"Item Data"
//	@Success	201		{object}	repo.ItemSummary
//	@Router		/v1/items [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemsCreate() errchain.HandlerFunc {
    fn := func(r *http.Request, body repo.ItemCreate) (repo.ItemOut, error) {
        return ctrl.svc.Items.Create(services.NewContext(r.Context()), body)
    }

    return adapters.Action(fn, http.StatusCreated)
}
```

### PUT - Update (using adapters.ActionID)

```go
// HandleItemUpdate godocs
//
//	@Summary	Update Item
//	@Tags		Items
//	@Produce	json
//	@Param		id		path		string			true	"Item ID"
//	@Param		payload	body		repo.ItemUpdate	true	"Item Data"
//	@Success	200		{object}	repo.ItemOut
//	@Router		/v1/items/{id} [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemUpdate() errchain.HandlerFunc {
    fn := func(r *http.Request, ID uuid.UUID, body repo.ItemUpdate) (repo.ItemOut, error) {
        auth := services.NewContext(r.Context())
        body.ID = ID
        return ctrl.repo.Items.UpdateByGroup(auth, auth.GID, body)
    }

    return adapters.ActionID("id", fn, http.StatusOK)
}
```

### DELETE (using adapters.CommandID)

```go
// HandleLabelDelete godocs
//
//	@Summary	Delete Label
//	@Tags		Labels
//	@Produce	json
//	@Param		id	path	string	true	"Label ID"
//	@Success	204
//	@Router		/v1/labels/{id} [DELETE]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelDelete() errchain.HandlerFunc {
    fn := func(r *http.Request, ID uuid.UUID) (any, error) {
        auth := services.NewContext(r.Context())
        err := ctrl.repo.Labels.DeleteByGroup(auth, auth.GID, ID)
        return nil, err
    }

    return adapters.CommandID("id", fn, http.StatusNoContent)
}
```

### File Upload (manual handling)

For file uploads, adapters cannot be used - manual handling is required:

```go
// HandleItemAttachmentCreate godocs
//
//	@Summary	Create Item Attachment
//	@Tags		Items Attachments
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		id		path		string	true	"Item ID"
//	@Param		file	formData	file	true	"File attachment"
//	@Param		type	formData	string	false	"Type of file"
//	@Param		name	formData	string	true	"name of the file including extension"
//	@Success	200		{object}	repo.ItemOut
//	@Failure	422		{object}	validate.ErrorResponse
//	@Router		/v1/items/{id}/attachments [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemAttachmentCreate() errchain.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) error {
        err := r.ParseMultipartForm(ctrl.maxUploadSize << 20)
        if err != nil {
            return validate.NewRequestError(errors.New("failed to parse multipart form"), http.StatusBadRequest)
        }

        file, _, err := r.FormFile("file")
        if err != nil {
            // handle error...
        }

        attachmentName := r.FormValue("name")
        // ... process and save attachment
    }
}
```

### Complex Query Extraction (manual inside HandleFunc)

For complex queries that don't fit standard adapters:

```go
// HandleItemsGetAll godoc
//
//	@Summary	Query All Items
//	@Tags		Items
//	@Produce	json
//	@Param		q			query		string		false	"search string"
//	@Param		page		query		int			false	"page number"
//	@Param		pageSize	query		int			false	"items per page"
//	@Param		labels		query		[]string	false	"label Ids"		collectionFormat(multi)
//	@Param		locations	query		[]string	false	"location Ids"	collectionFormat(multi)
//	@Success	200			{object}	repo.PaginationResult[repo.ItemSummary]{}
//	@Router		/v1/items [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemsGetAll() errchain.HandlerFunc {
    extractQuery := func(r *http.Request) repo.ItemQuery {
        params := r.URL.Query()
        return repo.ItemQuery{
            Page:            queryIntOrNegativeOne(params.Get("page")),
            PageSize:        queryIntOrNegativeOne(params.Get("pageSize")),
            Search:          params.Get("q"),
            LocationIDs:     queryUUIDList(params, "locations"),
            LabelIDs:        queryUUIDList(params, "labels"),
            IncludeArchived: queryBool(params.Get("includeArchived")),
        }
    }

    return func(w http.ResponseWriter, r *http.Request) error {
        ctx := services.NewContext(r.Context())
        items, err := ctrl.repo.Items.QueryByGroup(ctx, ctx.GID, extractQuery(r))
        if err != nil {
            return validate.NewRequestError(err, http.StatusInternalServerError)
        }
        return server.JSON(w, http.StatusOK, items)
    }
}
```

## Routing

Routes are defined in `backend/app/api/routes.go` using chi router:

```go
func (a *app) mountRoutes(r *chi.Mux, chain *errchain.ErrChain, repos *repo.AllRepos) {
    v1Ctrl := v1.NewControllerV1(
        a.services,
        a.repos,
        a.bus,
        a.conf,
        v1.WithMaxUploadSize(a.conf.Web.MaxUploadSize),
        v1.WithRegistration(a.conf.Options.AllowRegistration),
        v1.WithDemoStatus(a.conf.Demo),
    )

    userMW := []errchain.Middleware{
        a.mwAuthToken,
        a.mwRoles(RoleModeOr, authroles.RoleUser.String()),
    }

    r.Route(prefix+"/v1", func(r chi.Router) {
        // Public endpoints
        r.Get("/status", chain.ToHandlerFunc(v1Ctrl.HandleBase(...)))
        r.Post("/users/login", chain.ToHandlerFunc(v1Ctrl.HandleAuthLogin(...)))

        // Protected endpoints (with userMW middleware)
        r.Get("/items", chain.ToHandlerFunc(v1Ctrl.HandleItemsGetAll(), userMW...))
        r.Post("/items", chain.ToHandlerFunc(v1Ctrl.HandleItemsCreate(), userMW...))
        r.Get("/items/{id}", chain.ToHandlerFunc(v1Ctrl.HandleItemGet(), userMW...))
        r.Put("/items/{id}", chain.ToHandlerFunc(v1Ctrl.HandleItemUpdate(), userMW...))
        r.Delete("/items/{id}", chain.ToHandlerFunc(v1Ctrl.HandleItemDelete(), userMW...))
    })
}
```

## Query Parameter Helpers

Located in `query_params.go`:

```go
func queryIntOrNegativeOne(s string) int {
    i, err := strconv.Atoi(s)
    if err != nil {
        return -1
    }
    return i
}

func queryBool(s string) bool {
    b, err := strconv.ParseBool(s)
    if err != nil {
        return false
    }
    return b
}

func queryUUIDList(params url.Values, key string) []uuid.UUID {
    var ids []uuid.UUID
    for _, id := range params[key] {
        uid, err := uuid.Parse(id)
        if err != nil {
            continue
        }
        ids = append(ids, uid)
    }
    return ids
}
```

## Response Helpers

```go
// From github.com/hay-kot/httpkit/server
server.JSON(w, http.StatusOK, data)           // JSON response with status code

// From internal/sys/validate
validate.NewRequestError(err, http.StatusBadRequest)  // Error response
validate.NewRouteKeyError(key)                        // Invalid route parameter error
```

## Controller Helpers

Located in `partials.go`:

```go
// Extract UUID from route parameter "id"
func (ctrl *V1Controller) routeID(r *http.Request) (uuid.UUID, error) {
    return ctrl.routeUUID(r, "id")
}

// Extract UUID from any route parameter
func (ctrl *V1Controller) routeUUID(r *http.Request, key string) (uuid.UUID, error) {
    ID, err := uuid.Parse(chi.URLParam(r, key))
    if err != nil {
        return uuid.Nil, validate.NewRouteKeyError(key)
    }
    return ID, nil
}
```

## Adding a New Endpoint

### 1. Create Handler

In `backend/app/api/handlers/v1/v1_ctrl_myentity.go`:

```go
// HandleMyEntityGet godoc
//
//	@Summary	Get MyEntity
//	@Tags		MyEntity
//	@Produce	json
//	@Param		id	path		string	true	"MyEntity ID"
//	@Success	200	{object}	repo.MyEntityOut
//	@Router		/v1/my-entity/{id} [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleMyEntityGet() errchain.HandlerFunc {
    fn := func(r *http.Request, ID uuid.UUID) (repo.MyEntityOut, error) {
        auth := services.NewContext(r.Context())
        return ctrl.repo.MyEntity.GetOneByGroup(auth, auth.GID, ID)
    }

    return adapters.CommandID("id", fn, http.StatusOK)
}
```

### 2. Add Route

In `backend/app/api/routes.go`:

```go
r.Get("/my-entity/{id}", chain.ToHandlerFunc(v1Ctrl.HandleMyEntityGet(), userMW...))
```

### 3. Generate Docs

```bash
task generate  # Generates Swagger docs and TypeScript types
```

### 4. Test

```bash
task go:build  # Verify builds
task go:test   # Run tests
```

## Critical Rules

1. **Use adapters pattern** - Use `adapters.Command`, `adapters.Action`, etc. instead of manual request handling
2. **ALWAYS add Swagger comments** - Required for API docs and TypeScript type generation
3. **Run `task generate` after handler changes** - Updates API documentation and frontend types
4. **Always use `services.NewContext`** - Extracts authenticated user from request context
5. **Use `auth.GID` for multi-tenancy** - Always scope queries to the user's group
6. **Handle errors properly** - Return errors directly (adapters handle conversion) or use `validate.NewRequestError()`
7. **Return correct status codes** - 200 OK, 201 Created, 204 No Content, 400 Bad Request, 404 Not Found

## Common Issues

- **"Missing Swagger docs"** → Add `@Summary`, `@Tags`, `@Router` comments, run `task generate`
- **TypeScript types outdated** → Run `task generate` to regenerate
- **Auth failures** → Ensure route has `userMW...` middleware and `@Security Bearer`
- **Wrong adapter** → Use `Command` for no body, `Action` for body, `Query` for query params
- **UUID not found** → Check the path parameter name matches `adapters.CommandID("id", ...)` or use `ctrl.routeUUID(r, "paramName")`

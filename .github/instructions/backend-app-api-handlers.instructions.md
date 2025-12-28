---
applyTo: '/backend/app/api/handlers/**/*'
---

# Backend API Handlers Instructions (`/backend/app/api/handlers/v1/`)

## Overview

API handlers are the HTTP layer that processes requests, calls services, and returns responses. All handlers use the V1 API pattern with Swagger documentation for auto-generation.

## Architecture Flow

```
HTTP Request → Router → Middleware → Handler → Service → Repository → Database
                                        ↓
                                  HTTP Response
```

## Directory Structure

```
backend/app/api/
├── routes.go                       # Route definitions and middleware
├── handlers/
│   └── v1/
│       ├── controller.go           # V1Controller struct and dependencies
│       ├── v1_ctrl_items.go        # Item endpoints
│       ├── v1_ctrl_users.go        # User endpoints
│       ├── v1_ctrl_locations.go    # Location endpoints
│       ├── v1_ctrl_auth.go         # Authentication endpoints
│       ├── helpers.go              # HTTP helper functions
│       ├── query_params.go         # Query parameter parsing
│       └── assets/                 # Asset handling
```

## Handler Structure

### V1Controller

All handlers are methods on `V1Controller`:

```go
type V1Controller struct {
    svc    *services.AllServices   // Service layer
    repo   *repo.AllRepos          // Direct repo access (rare)
    bus    *eventbus.EventBus      // Event publishing
}

func (ctrl *V1Controller) HandleItemCreate() errchain.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) error {
        // Handler logic
    }
}
```

### Swagger Documentation

**CRITICAL:** Every handler must have Swagger comments for API doc generation:

```go
// HandleItemsGetAll godoc
//
//	@Summary	Query All Items
//	@Tags		Items
//	@Produce	json
//	@Param		q			query		string		false	"search string"
//	@Param		page		query		int			false	"page number"
//	@Param		pageSize	query		int			false	"items per page"
//	@Success	200			{object}	repo.PaginationResult[repo.ItemSummary]{}
//	@Router		/v1/items [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemsGetAll() errchain.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) error {
        // ...
    }
}
```

**After modifying Swagger comments, ALWAYS run:**
```bash
task generate  # Regenerates Swagger docs and TypeScript types
```

## Standard Handler Pattern

### 1. Decode Request

```go
func (ctrl *V1Controller) HandleItemCreate() errchain.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) error {
        var itemData repo.ItemCreate
        if err := server.Decode(r, &itemData); err != nil {
            return validate.NewRequestError(err, http.StatusBadRequest)
        }
        
        // ... rest of handler
    }
}
```

### 2. Extract Context

```go
// Get current user from request (added by auth middleware)
user := ctrl.CurrentUser(r)

// Create service context with group/user IDs
ctx := services.NewContext(r.Context(), user)
```

### 3. Call Service

```go
result, err := ctrl.svc.Items.Create(ctx, itemData)
if err != nil {
    return validate.NewRequestError(err, http.StatusInternalServerError)
}
```

### 4. Return Response

```go
return server.JSON(w, result, http.StatusCreated)
```

## Common Handler Patterns

### GET - Single Item

```go
// HandleItemGet godoc
//
//	@Summary	Get Item
//	@Tags		Items
//	@Produce	json
//	@Param		id	path		string	true	"Item ID"
//	@Success	200	{object}	repo.ItemOut
//	@Router		/v1/items/{id} [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemGet() errchain.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) error {
        id, err := ctrl.RouteUUID(r, "id")
        if err != nil {
            return err
        }
        
        ctx := services.NewContext(r.Context(), ctrl.CurrentUser(r))
        item, err := ctrl.svc.Items.Get(ctx, id)
        if err != nil {
            return validate.NewRequestError(err, http.StatusNotFound)
        }
        
        return server.JSON(w, item, http.StatusOK)
    }
}
```

### GET - List with Pagination

```go
func (ctrl *V1Controller) HandleItemsGetAll() errchain.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) error {
        // Parse query parameters
        query := extractItemQuery(r)
        
        ctx := services.NewContext(r.Context(), ctrl.CurrentUser(r))
        items, err := ctrl.svc.Items.GetAll(ctx, query)
        if err != nil {
            return err
        }
        
        return server.JSON(w, items, http.StatusOK)
    }
}

// Helper to extract query params
func extractItemQuery(r *http.Request) repo.ItemQuery {
    params := r.URL.Query()
    return repo.ItemQuery{
        Page:       queryIntOrNegativeOne(params.Get("page")),
        PageSize:   queryIntOrNegativeOne(params.Get("pageSize")),
        Search:     params.Get("q"),
        LocationIDs: queryUUIDList(params, "locations"),
    }
}
```

### POST - Create

```go
// HandleItemCreate godoc
//
//	@Summary	Create Item
//	@Tags		Items
//	@Accept		json
//	@Produce	json
//	@Param		payload	body		repo.ItemCreate	true	"Item Data"
//	@Success	201		{object}	repo.ItemOut
//	@Router		/v1/items [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemCreate() errchain.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) error {
        var data repo.ItemCreate
        if err := server.Decode(r, &data); err != nil {
            return validate.NewRequestError(err, http.StatusBadRequest)
        }
        
        ctx := services.NewContext(r.Context(), ctrl.CurrentUser(r))
        item, err := ctrl.svc.Items.Create(ctx, data)
        if err != nil {
            return err
        }
        
        return server.JSON(w, item, http.StatusCreated)
    }
}
```

### PUT - Update

```go
// HandleItemUpdate godoc
//
//	@Summary	Update Item
//	@Tags		Items
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string			true	"Item ID"
//	@Param		payload	body		repo.ItemUpdate	true	"Item Data"
//	@Success	200		{object}	repo.ItemOut
//	@Router		/v1/items/{id} [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemUpdate() errchain.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) error {
        id, err := ctrl.RouteUUID(r, "id")
        if err != nil {
            return err
        }
        
        var data repo.ItemUpdate
        if err := server.Decode(r, &data); err != nil {
            return validate.NewRequestError(err, http.StatusBadRequest)
        }
        
        ctx := services.NewContext(r.Context(), ctrl.CurrentUser(r))
        item, err := ctrl.svc.Items.Update(ctx, id, data)
        if err != nil {
            return err
        }
        
        return server.JSON(w, item, http.StatusOK)
    }
}
```

### DELETE

```go
// HandleItemDelete godoc
//
//	@Summary	Delete Item
//	@Tags		Items
//	@Param		id	path	string	true	"Item ID"
//	@Success	204
//	@Router		/v1/items/{id} [DELETE]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemDelete() errchain.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) error {
        id, err := ctrl.RouteUUID(r, "id")
        if err != nil {
            return err
        }
        
        ctx := services.NewContext(r.Context(), ctrl.CurrentUser(r))
        err = ctrl.svc.Items.Delete(ctx, id)
        if err != nil {
            return err
        }
        
        return server.JSON(w, nil, http.StatusNoContent)
    }
}
```

### File Upload

```go
func (ctrl *V1Controller) HandleItemAttachmentCreate() errchain.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) error {
        id, err := ctrl.RouteUUID(r, "id")
        if err != nil {
            return err
        }
        
        // Parse multipart form
        err = r.ParseMultipartForm(32 << 20) // 32MB max
        if err != nil {
            return err
        }
        
        file, header, err := r.FormFile("file")
        if err != nil {
            return err
        }
        defer file.Close()
        
        ctx := services.NewContext(r.Context(), ctrl.CurrentUser(r))
        attachment, err := ctrl.svc.Items.CreateAttachment(ctx, id, file, header.Filename)
        if err != nil {
            return err
        }
        
        return server.JSON(w, attachment, http.StatusCreated)
    }
}
```

## Routing

Routes are defined in `backend/app/api/routes.go`:

```go
func (a *app) mountRoutes(repos *repo.AllRepos, svc *services.AllServices) {
    v1 := v1.NewControllerV1(svc, repos)
    
    a.server.Get("/api/v1/items", v1.HandleItemsGetAll())
    a.server.Post("/api/v1/items", v1.HandleItemCreate())
    a.server.Get("/api/v1/items/{id}", v1.HandleItemGet())
    a.server.Put("/api/v1/items/{id}", v1.HandleItemUpdate())
    a.server.Delete("/api/v1/items/{id}", v1.HandleItemDelete())
}
```

## Helper Functions

### Query Parameter Parsing

Located in `query_params.go`:

```go
func queryIntOrNegativeOne(s string) int
func queryBool(s string) bool
func queryUUIDList(params url.Values, key string) []uuid.UUID
```

### Response Helpers

```go
// From httpkit/server
server.JSON(w, data, statusCode)           // JSON response
server.Respond(w, statusCode)              // Empty response
validate.NewRequestError(err, statusCode)  // Error response
```

### Authentication

```go
user := ctrl.CurrentUser(r)  // Get authenticated user (from middleware)
```

## Adding a New Endpoint

### 1. Create Handler

In `backend/app/api/handlers/v1/v1_ctrl_myentity.go`:

```go
// HandleMyEntityCreate godoc
//
//	@Summary	Create MyEntity
//	@Tags		MyEntity
//	@Accept		json
//	@Produce	json
//	@Param		payload	body		repo.MyEntityCreate	true	"Data"
//	@Success	201		{object}	repo.MyEntityOut
//	@Router		/v1/my-entity [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleMyEntityCreate() errchain.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) error {
        var data repo.MyEntityCreate
        if err := server.Decode(r, &data); err != nil {
            return validate.NewRequestError(err, http.StatusBadRequest)
        }
        
        ctx := services.NewContext(r.Context(), ctrl.CurrentUser(r))
        result, err := ctrl.svc.MyEntity.Create(ctx, data)
        if err != nil {
            return err
        }
        
        return server.JSON(w, result, http.StatusCreated)
    }
}
```

### 2. Add Route

In `backend/app/api/routes.go`:

```go
a.server.Post("/api/v1/my-entity", v1.HandleMyEntityCreate())
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

1. **ALWAYS add Swagger comments** - required for API docs and TypeScript type generation
2. **Run `task generate` after handler changes** - updates API documentation
3. **Use services, not repos directly** - handlers call services, services call repos
4. **Always use `services.Context`** - includes auth and multi-tenancy
5. **Handle errors properly** - use `validate.NewRequestError()` with appropriate status codes
6. **Validate input** - decode and validate request bodies
7. **Return correct status codes** - 200 OK, 201 Created, 204 No Content, 400 Bad Request, 404 Not Found

## Common Issues

- **"Missing Swagger docs"** → Add `@Summary`, `@Tags`, `@Router` comments, run `task generate`
- **TypeScript types outdated** → Run `task generate` to regenerate
- **Auth failures** → Ensure route has auth middleware and `@Security Bearer`
- **CORS errors** → Check middleware configuration in `routes.go`

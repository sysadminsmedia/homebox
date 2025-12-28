# Backend Services Layer Instructions (`/backend/internal/core/services/`)

## Overview

The services layer contains business logic that orchestrates between repositories and API handlers. Services handle complex operations, validation, and cross-cutting concerns.

## Architecture Pattern

```
Handler (API) → Service (Business Logic) → Repository (Data Access) → Database
```

**Separation of concerns:**
- **Handlers** (`backend/app/api/handlers/v1/`) - HTTP request/response, routing, auth
- **Services** (`backend/internal/core/services/`) - Business logic, orchestration
- **Repositories** (`backend/internal/data/repo/`) - Database operations, queries

## Directory Structure

```
backend/internal/core/services/
├── all.go                          # Service aggregation
├── service_items.go                # Item business logic
├── service_items_attachments.go    # Item attachments logic
├── service_user.go                 # User management logic
├── service_group.go                # Group management logic
├── service_background.go           # Background tasks
├── contexts.go                     # Service context types
├── reporting/                      # Reporting subsystem
│   ├── eventbus/                   # Event bus for notifications
│   └── *.go                        # Report generation logic
└── *_test.go                       # Service tests
```

## Service Structure

### Standard Pattern

```go
type ItemService struct {
    repo     *repo.AllRepos          // Access to all repositories
    filepath string                   // File storage path
    autoIncrementAssetID bool        // Feature flags
}

func (svc *ItemService) Create(ctx Context, item repo.ItemCreate) (repo.ItemOut, error) {
    // 1. Validation
    if item.Name == "" {
        return repo.ItemOut{}, errors.New("name required")
    }
    
    // 2. Business logic
    if svc.autoIncrementAssetID {
        highest, err := svc.repo.Items.GetHighestAssetID(ctx, ctx.GID)
        if err != nil {
            return repo.ItemOut{}, err
        }
        item.AssetID = highest + 1
    }
    
    // 3. Repository call
    return svc.repo.Items.Create(ctx, ctx.GID, item)
}
```

### Service Context

Services use a custom `Context` type that extends `context.Context`:

```go
type Context struct {
    context.Context
    GID uuid.UUID  // Group ID for multi-tenancy
    UID uuid.UUID  // User ID for audit
}
```

**Always use `Context` from services package, not raw `context.Context`.**

## Common Service Patterns

### 1. CRUD with Business Logic

```go
func (svc *ItemService) Update(ctx Context, id uuid.UUID, data repo.ItemUpdate) (repo.ItemOut, error) {
    // Fetch existing
    existing, err := svc.repo.Items.Get(ctx, id)
    if err != nil {
        return repo.ItemOut{}, err
    }
    
    // Business rules
    if existing.Archived && data.Quantity != nil {
        return repo.ItemOut{}, errors.New("cannot modify archived items")
    }
    
    // Update
    return svc.repo.Items.Update(ctx, id, data)
}
```

### 2. Orchestrating Multiple Repositories

```go
func (svc *ItemService) CreateWithAttachment(ctx Context, item repo.ItemCreate, file io.Reader) (repo.ItemOut, error) {
    // Create item
    created, err := svc.repo.Items.Create(ctx, ctx.GID, item)
    if err != nil {
        return repo.ItemOut{}, err
    }
    
    // Upload attachment
    attachment, err := svc.repo.Attachments.Create(ctx, created.ID, file)
    if err != nil {
        // Rollback - delete item
        _ = svc.repo.Items.Delete(ctx, created.ID)
        return repo.ItemOut{}, err
    }
    
    created.Attachments = []repo.AttachmentOut{attachment}
    return created, nil
}
```

### 3. Background Tasks

```go
func (svc *ItemService) EnsureAssetID(ctx context.Context, gid uuid.UUID) (int, error) {
    // Get items without asset IDs
    items, err := svc.repo.Items.GetAllZeroAssetID(ctx, gid)
    if err != nil {
        return 0, err
    }
    
    // Batch assign
    highest := svc.repo.Items.GetHighestAssetID(ctx, gid)
    for _, item := range items {
        highest++
        _ = svc.repo.Items.Update(ctx, item.ID, repo.ItemUpdate{
            AssetID: &highest,
        })
    }
    
    return len(items), nil
}
```

### 4. Event Publishing

Services can publish events to the event bus:

```go
func (svc *ItemService) Delete(ctx Context, id uuid.UUID) error {
    err := svc.repo.Items.Delete(ctx, id)
    if err != nil {
        return err
    }
    
    // Publish event for notifications
    svc.repo.Bus.Publish(eventbus.Event{
        Type: "item.deleted",
        Data: map[string]interface{}{"id": id},
    })
    
    return nil
}
```

## Service Aggregation

All services are bundled in `all.go`:

```go
type AllServices struct {
    User  *UserService
    Group *GroupService
    Items *ItemService
    // ... other services
}

func New(repos *repo.AllRepos, filepath string) *AllServices {
    return &AllServices{
        User:  &UserService{repo: repos},
        Items: &ItemService{repo: repos, filepath: filepath},
        // ...
    }
}
```

**Accessed in handlers via:**
```go
ctrl.svc.Items.Create(ctx, itemData)
```

## Working with Services from Handlers

Handlers call services, not repositories directly:

```go
// In backend/app/api/handlers/v1/v1_ctrl_items.go
func (ctrl *V1Controller) HandleItemCreate() errchain.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) error {
        var itemData repo.ItemCreate
        if err := server.Decode(r, &itemData); err != nil {
            return err
        }
        
        // Get context with group/user IDs
        ctx := services.NewContext(r.Context(), ctrl.CurrentUser(r))
        
        // Call service (not repository)
        item, err := ctrl.svc.Items.Create(ctx, itemData)
        if err != nil {
            return err
        }
        
        return server.JSON(w, item, http.StatusCreated)
    }
}
```

## Testing Services

Service tests mock repositories using interfaces:

```go
func TestItemService_Create(t *testing.T) {
    mockRepo := &mockItemRepo{
        CreateFunc: func(ctx context.Context, gid uuid.UUID, data repo.ItemCreate) (repo.ItemOut, error) {
            return repo.ItemOut{ID: uuid.New(), Name: data.Name}, nil
        },
    }
    
    svc := &ItemService{repo: &repo.AllRepos{Items: mockRepo}}
    
    ctx := services.Context{GID: uuid.New(), UID: uuid.New()}
    result, err := svc.Create(ctx, repo.ItemCreate{Name: "Test"})
    
    assert.NoError(t, err)
    assert.Equal(t, "Test", result.Name)
}
```

**Run service tests:**
```bash
cd backend && go test ./internal/core/services -v
```

## Adding a New Service

### 1. Create Service File

Create `backend/internal/core/services/service_myentity.go`:

```go
package services

type MyEntityService struct {
    repo *repo.AllRepos
}

func (svc *MyEntityService) Create(ctx Context, data repo.MyEntityCreate) (repo.MyEntityOut, error) {
    // Business logic here
    return svc.repo.MyEntity.Create(ctx, ctx.GID, data)
}
```

### 2. Add to AllServices

Edit `backend/internal/core/services/all.go`:

```go
type AllServices struct {
    // ... existing services
    MyEntity *MyEntityService
}

func New(repos *repo.AllRepos, filepath string) *AllServices {
    return &AllServices{
        // ... existing services
        MyEntity: &MyEntityService{repo: repos},
    }
}
```

### 3. Use in Handler

In `backend/app/api/handlers/v1/`:

```go
func (ctrl *V1Controller) HandleMyEntityCreate() errchain.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) error {
        ctx := services.NewContext(r.Context(), ctrl.CurrentUser(r))
        result, err := ctrl.svc.MyEntity.Create(ctx, data)
        // ...
    }
}
```

### 4. Run Tests

```bash
task generate    # If you modified schemas
task go:test     # Run all tests
```

## Common Service Responsibilities

**Services should:**
- ✅ Contain business logic and validation
- ✅ Orchestrate multiple repository calls
- ✅ Handle transactions (when needed)
- ✅ Publish events for side effects
- ✅ Enforce access control and multi-tenancy
- ✅ Transform data between API and repository formats

**Services should NOT:**
- ❌ Handle HTTP requests/responses (that's handlers)
- ❌ Construct SQL queries (that's repositories)
- ❌ Import handler packages (creates circular deps)
- ❌ Directly access database (use repositories)

## Critical Rules

1. **Always use `services.Context`** - includes group/user IDs for multi-tenancy
2. **Services call repos, handlers call services** - maintains layer separation
3. **No direct database access** - always through repositories
4. **Business logic goes here** - not in handlers or repositories
5. **Test services independently** - mock repository dependencies

## Common Patterns to Follow

- **Validation:** Check business rules before calling repository
- **Error wrapping:** Add context to repository errors
- **Logging:** Use `log.Ctx(ctx)` for contextual logging
- **Transactions:** Use `repo.WithTx()` for multi-step operations
- **Events:** Publish to event bus for notifications/side effects

# Backend Data Layer Instructions (`/backend/internal/data/`)

## Overview

This directory contains the data access layer using **Ent ORM** (entity framework). It follows a clear separation between schema definitions, generated code, and repository implementations.

## Directory Structure

```
backend/internal/data/
├── ent/                    # Ent ORM generated code (DO NOT EDIT)
│   ├── schema/             # Schema definitions (EDIT THESE)
│   │   ├── item.go         # Item entity schema
│   │   ├── user.go         # User entity schema
│   │   ├── location.go     # Location entity schema
│   │   ├── label.go        # Label entity schema
│   │   └── mixins/         # Reusable schema mixins
│   ├── *.go                # Generated entity code
│   └── migrate/            # Generated migrations
├── repo/                   # Repository pattern implementations
│   ├── repos_all.go        # Aggregates all repositories
│   ├── repo_items.go       # Item repository
│   ├── repo_users.go       # User repository
│   ├── repo_locations.go   # Location repository
│   └── *_test.go           # Repository tests
├── migrations/             # Manual SQL migrations
│   ├── sqlite3/            # SQLite-specific migrations
│   └── postgres/           # PostgreSQL-specific migrations
└── types/                  # Custom data types
```

## Ent ORM Workflow

### 1. Defining Schemas (`ent/schema/`)

**ALWAYS edit schema files here** - these define your database entities:

```go
// Example: backend/internal/data/ent/schema/item.go
type Item struct {
    ent.Schema
}

func (Item) Fields() []ent.Field {
    return []ent.Field{
        field.String("name").NotEmpty(),
        field.Int("quantity").Default(1),
        field.Bool("archived").Default(false),
    }
}

func (Item) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("location", Location.Type).Ref("items").Unique(),
        edge.From("labels", Label.Type).Ref("items"),
    }
}

func (Item) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("name"),
        index.Fields("archived"),
    }
}
```

**Common schema patterns:**
- Use `mixins.BaseMixin{}` for `id`, `created_at`, `updated_at` fields
- Use `mixins.DetailsMixin{}` for `name` and `description` fields
- Use `GroupMixin{ref: "items"}` to link entities to groups
- Add indexes for frequently queried fields

### 2. Generating Code

**After modifying any schema file, ALWAYS run:**

```bash
task generate
```

This:
1. Runs `go generate ./...` in `backend/internal/` (generates Ent code)
2. Generates Swagger docs from API handlers
3. Generates TypeScript types for frontend

**Generated files you'll see:**
- `ent/*.go` - Entity types, builders, queries
- `ent/migrate/migrate.go` - Auto migrations
- `ent/predicate/predicate.go` - Query predicates

**NEVER edit generated files directly** - changes will be overwritten.

### 3. Using Generated Code in Repositories

Repositories in `repo/` use the generated Ent client:

```go
// Example: backend/internal/data/repo/repo_items.go
type ItemsRepository struct {
    db  *ent.Client
    bus *eventbus.EventBus
}

func (r *ItemsRepository) Create(ctx context.Context, gid uuid.UUID, data ItemCreate) (ItemOut, error) {
    entity, err := r.db.Item.Create().
        SetName(data.Name).
        SetQuantity(data.Quantity).
        SetGroupID(gid).
        Save(ctx)
    
    return mapToItemOut(entity), err
}
```

## Repository Pattern

### Structure

Each entity typically has:
- **Repository struct** (`ItemsRepository`) - holds DB client and dependencies
- **Input types** (`ItemCreate`, `ItemUpdate`) - API input DTOs
- **Output types** (`ItemOut`, `ItemSummary`) - API response DTOs
- **Query types** (`ItemQuery`) - search/filter parameters
- **Mapper functions** (`mapToItemOut`) - converts Ent entities to output DTOs

### Key Methods

Repositories typically implement:
- `Create(ctx, gid, input)` - Create new entity
- `Get(ctx, id)` - Get single entity by ID
- `GetAll(ctx, gid, query)` - Query with pagination/filters
- `Update(ctx, id, input)` - Update entity
- `Delete(ctx, id)` - Delete entity

### Working with Ent Queries

**Loading relationships (edges):**
```go
items, err := r.db.Item.Query().
    WithLocation().        // Load location edge
    WithLabels().          // Load labels edge
    WithChildren().        // Load child items
    Where(item.GroupIDEQ(gid)).
    All(ctx)
```

**Filtering:**
```go
query := r.db.Item.Query().
    Where(
        item.GroupIDEQ(gid),
        item.ArchivedEQ(false),
        item.NameContainsFold(search),
    )
```

**Ordering and pagination:**
```go
items, err := query.
    Order(ent.Desc(item.FieldCreatedAt)).
    Limit(pageSize).
    Offset((page - 1) * pageSize).
    All(ctx)
```

## Common Workflows

### Adding a New Entity

1. **Create schema:** `backend/internal/data/ent/schema/myentity.go`
2. **Run:** `task generate` (generates Ent code)
3. **Create repository:** `backend/internal/data/repo/repo_myentity.go`
4. **Add to AllRepos:** Edit `repo/repos_all.go` to include new repo
5. **Run tests:** `task go:test`

### Adding Fields to Existing Entity

1. **Edit schema:** `backend/internal/data/ent/schema/item.go`
   ```go
   field.String("new_field").Optional()
   ```
2. **Run:** `task generate`
3. **Update repository:** Add field to input/output types in `repo/repo_items.go`
4. **Update mappers:** Ensure mapper functions handle new field
5. **Run tests:** `task go:test`

### Adding Relationships (Edges)

1. **Edit both schemas:**
   ```go
   // In item.go
   edge.From("location", Location.Type).Ref("items").Unique()
   
   // In location.go
   edge.To("items", Item.Type)
   ```
2. **Run:** `task generate`
3. **Use in queries:** `.WithLocation()` to load the edge
4. **Run tests:** `task go:test`

## Testing

Repository tests use `enttest` for in-memory SQLite:

```go
func TestItemRepo(t *testing.T) {
    client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
    defer client.Close()
    
    repo := &ItemsRepository{db: client}
    // Test methods...
}
```

**Run repository tests:**
```bash
cd backend && go test ./internal/data/repo -v
```

## Critical Rules

1. **ALWAYS run `task generate` after schema changes** - builds will fail otherwise
2. **NEVER edit files in `ent/` except `ent/schema/`** - they're generated
3. **Use repositories, not raw Ent queries in services/handlers** - maintains separation
4. **Include `group_id` in all queries** - ensures multi-tenancy
5. **Use `.WithX()` to load edges** - avoids N+1 queries
6. **Test with both SQLite and PostgreSQL** - CI tests both

## Common Errors

- **"undefined: ent.ItemX"** → Run `task generate` after schema changes
- **Migration conflicts** → Check `migrations/` for manual migration files
- **Foreign key violations** → Ensure edges are properly defined in both schemas
- **Slow queries** → Add indexes in schema `Indexes()` method

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Task Runner

This project uses [Task](https://taskfile.dev) (`task`) as its build system. Run `task --list-all` to see all available commands.

## Common Commands

### Backend (Go)
```bash
task go:run          # Start API server (localhost:7745)
task go:run:postgresql  # Start with PostgreSQL instead of SQLite
task go:test         # Run all Go tests
task go:lint         # Run golangci-lint
task go:build        # Build backend binary to ./build/backend
task go:coverage     # Run tests with race flag and coverage report
```

### Frontend (Vue/Nuxt)
```bash
task ui:dev          # Start dev server (localhost:3000)
task ui:watch        # Run Vitest in watch mode
task ui:fix          # Run prettier + eslint --fix
task ui:check        # Run TypeScript typecheck
```

### Code Generation (run after schema/API changes)
```bash
task generate        # Run all code generation (Ent ORM, Swagger, TypeScript types)
task db:generate     # Entgo ORM code generation only
task swag            # Generate Swagger docs from Go comments only
task typescript-types # Generate TypeScript types from Swagger only
```

### PR / CI
```bash
task pr              # Full PR check: generate, lint, test, build
task test:ci         # E2E tests against live server (SQLite)
task test:e2e        # Full E2E with Playwright
```

### Running single tests
```bash
# Backend - specific package or test function
cd backend
go test -v ./internal/data/repo/... -run TestRepo_Items
go test -v ./internal/core/services/...

# Frontend - specific Playwright E2E spec
cd frontend
pnpm exec playwright test test/e2e/login.browser.spec.ts
```

## Architecture

HomeBox is a multi-tenant inventory management system. The backend is an embedded binary that serves both the REST API and the pre-built frontend.

### Backend Layers (`/backend`)

```
HTTP Request → API Handler (app/api/handlers/v1/) 
            → Service Layer (internal/core/services/)
            → Repository Layer (internal/data/repo/)
            → Entgo ORM (internal/data/ent/)
            → SQLite / PostgreSQL
```

- **`app/api/`** — Chi router, middleware (auth, rate limiting, CORS), route registration in `routes.go`
- **`app/api/handlers/v1/`** — One file per domain (items, tags, locations, assets, auth, etc.). `controller.go` holds the base struct injected into all handlers.
- **`internal/core/services/`** — All business logic. Services receive a repository and operate on domain types. Event bus lives here for WebSocket notifications.
- **`internal/data/repo/`** — Thin query layer on top of Ent. One file per entity.
- **`internal/data/ent/schema/`** — Entgo schema definitions. **Edit these to change the data model**, then run `task db:generate`.
- **`internal/data/migrations/`** — Goose SQL migration files (auto-applied at startup).
- **`internal/sys/config/`** — All app configuration loaded from environment variables.
- **`pkgs/`** — Standalone packages: hasher, faker (test data), labelmaker, cgofreesqlite.

### Frontend (`/frontend`)

Nuxt 4 / Vue 3 app with file-based routing, Pinia stores, and Tailwind + shadcn-nuxt components.

- **`lib/api/`** — Auto-generated TypeScript API client. **Do not edit by hand**; regenerate with `task typescript-types`.
- **`pages/`** — File-based routes. Items, locations, labels, tags each have their own pages.
- **`stores/`** — Pinia stores (locations, tags). Item state is mostly local to pages.
- **`composables/`** — Reusable Vue composables.
- **`components/ui/`** — shadcn-nuxt component library (do not modify unless replacing a component).
- **`test/e2e/`** — Playwright browser tests. These require the backend to be running.

### Code Generation Pipeline

Modifying the data model requires a chain of regenerations:
1. Edit `internal/data/ent/schema/*.go`
2. `task db:generate` → regenerates all Entgo ORM code in `internal/data/ent/`
3. Add Swagger annotations to API handlers
4. `task swag` → updates `app/api/static/docs/`
5. `task typescript-types` → updates `frontend/lib/api/types/data-contracts.ts`

### Authentication

- Local username/password with JWT tokens
- OIDC/OAuth2 via `app/api/providers/`
- Auth middleware in `app/api/middleware.go` gates all `/api/v1/` routes

### Storage

- Files/attachments: GoCloud blob (local filesystem by default; configurable for S3, Azure, GCS)
- Database: SQLite (default dev) or PostgreSQL (set via `HBOX_DATABASE_DRIVER`)

### Testing Notes

- Backend unit tests use a real in-memory SQLite database (not mocks). Test helpers are in `pkgs/faker/` for generating test data.
- Frontend E2E tests (Playwright) require the API server to be running. First run may fail due to a race condition — re-run if that happens.
- Unit tests (Vitest) also require the API server running.

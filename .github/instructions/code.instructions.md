# Homebox Repository Instructions for Coding Agents

## Repository Overview

**Type**: Full-stack web application for home inventory management  
**Size**: ~265 Go files, ~371 TypeScript/Vue files  
**Architecture**: Monorepo with separate backend (Go) and frontend (Nuxt/Vue) projects  
**Database**: SQLite (default) or PostgreSQL  
**Primary Build Tool**: [Task](https://taskfile.dev) (Taskfile.yml in repo root)

### Technology Stack
- **Backend**: Go 1.24+ (located in `/backend`)
  - Framework: Chi router, Ent ORM
  - API: RESTful with Swagger/OpenAPI documentation
  - Port: 7745 (default)
- **Frontend**: Nuxt 4, Vue 3, TypeScript (located in `/frontend`)
  - Styling: Tailwind CSS, Shadcn-vue components
  - Testing: Vitest (unit), Playwright (e2e)
  - Package Manager: pnpm 9.1.4+
  - Dev server proxies `/api` to `http://localhost:7745`

## Critical Build & Validation Commands

**ALWAYS use `task` commands - they handle dependencies and proper sequencing.**

### Initial Setup (Run Once)
```bash
task setup
```
This installs:
- `swag` (Swagger code generator)
- `goose` (database migrations)
- Go dependencies
- pnpm dependencies for frontend

### Code Generation (Required Before Most Tasks)
```bash
task generate
```
**ALWAYS run this after**:
- Changing Go API code, handlers, or data models
- Modifying Ent schema files in `backend/internal/data/ent/schema/`
- Before running backend server or tests

This task:
1. Runs `task db:generate` - Generates Ent ORM code from schemas
2. Runs `task swag` - Generates Swagger/OpenAPI docs from Go comments
3. Runs `task typescript-types` - Generates TypeScript types from Swagger
4. Copies API docs to `docs/en/api/`

**Note**: `task generate` produces many "TypeSpecDef is nil" warnings - these are normal and do not indicate failure.

### Backend Commands
```bash
task go:build          # Build backend binary to build/backend
task go:test           # Run Go unit tests (fast)
task go:coverage       # Run tests with coverage report
task go:lint           # Run golangci-lint (required before PR)
task go:tidy           # Run go mod tidy
task go:all            # Run tidy + lint + test (comprehensive check)
task go:run            # Start backend server with SQLite (needs task generate first)
task go:run:postgresql # Start backend server with PostgreSQL
```

**Backend test runtime**: ~5-10 seconds  
**Backend build time**: ~60-90 seconds  
**Linting timeout**: Uses 6m timeout in CI

### Frontend Commands
```bash
cd frontend && pnpm install     # Install dependencies
task ui:dev                     # Start dev server on port 3000
task ui:check                   # Run TypeScript type checking
task ui:fix                     # Run eslint --fix and prettier
task ui:watch                   # Run Vitest in watch mode
```

**Frontend lint**: `pnpm run lint:ci` allows max 1 warning  
**Type checking**: `pnpm nuxi typecheck --noEmit`

### Integration & E2E Testing
```bash
task test:ci              # Run integration tests with SQLite (CI mode)
task test:ci:postgresql   # Run integration tests with PostgreSQL
task test:e2e             # Run Playwright E2E tests (builds frontend + backend)
```

**IMPORTANT**: Integration tests (`test:ci`) require:
1. Backend to be built and running
2. Frontend dependencies installed
3. Tests run in `frontend/` directory with Vitest
4. Runtime: ~15-30 seconds + startup time

**E2E tests** (`test:e2e`):
- Builds frontend into static files
- Copies to `backend/app/api/static/`
- Starts backend server
- Waits 30 seconds for startup
- Runs Playwright tests in 4 shards
- Runtime: ~60+ seconds per shard
- May require: `pnpm exec playwright install-deps && pnpm exec playwright install`

### PR Validation (Run Before Submitting)
```bash
task pr
```
This runs the full CI validation sequence:
1. `task generate` - Code generation
2. `task go:all` - Backend lint + test
3. `task ui:check` - Frontend type checking
4. `task ui:fix` - Frontend linting
5. `task test:ci` - Integration tests

**Total runtime**: ~3-5 minutes

## Project Structure

### Root Level Files
- `Taskfile.yml` - All build/test/run commands (always use `task`)
- `package.json` - Root workspace config (for docs only)
- `.gitlab-ci.yml` - GitLab CI pipeline (reference only)
- `docker-compose.yml` - Quick start Docker setup
- `Dockerfile*` - Three variants: standard, rootless, hardened

### Backend Structure (`/backend`)
```
backend/
├── app/
│   ├── api/              # Main API application
│   │   ├── main.go       # Entry point
│   │   ├── routes.go     # Route definitions
│   │   ├── handlers/     # HTTP handlers (v1 API)
│   │   ├── static/       # Swagger docs, embedded frontend
│   │   └── providers/    # Service providers
│   └── tools/
│       └── typegen/      # TypeScript type generation tool
├── internal/
│   ├── core/
│   │   └── services/     # Business logic layer
│   ├── data/
│   │   ├── ent/          # Ent ORM generated code + schemas
│   │   │   └── schema/   # Schema definitions (edit these)
│   │   └── repo/         # Repository pattern implementations
│   ├── sys/              # System utilities (config, validation)
│   └── web/              # Web middleware
├── pkgs/                 # Reusable packages
├── go.mod, go.sum        # Go dependencies
└── .golangci.yml         # Linter configuration
```

**Key patterns**:
- Schema changes: Edit `backend/internal/data/ent/schema/*.go`, then `task generate`
- API changes: Edit handlers in `backend/app/api/handlers/v1/`, add Swagger comments, then `task generate`
- Generated code in `backend/internal/data/ent/` - DO NOT edit directly

### Frontend Structure (`/frontend`)
```
frontend/
├── app.vue               # Root component
├── nuxt.config.ts        # Nuxt configuration
├── package.json          # Frontend dependencies
├── components/           # Vue components (auto-imported)
├── pages/                # File-based routing
├── layouts/              # Layout components
├── composables/          # Vue composables (auto-imported)
├── stores/               # Pinia state stores
├── lib/
│   └── api/
│       └── types/        # Generated TypeScript API types
├── locales/              # i18n translations
├── test/                 # Vitest + Playwright tests
├── eslint.config.mjs     # ESLint configuration
└── tailwind.config.js    # Tailwind configuration
```

**Key patterns**:
- Components in `components/` are auto-imported
- Composables in `composables/` are auto-imported
- API types regenerated via `task generate` - DO NOT edit manually
- Tests in `test/` use Vitest config from `test/vitest.config.ts`

## CI/CD Workflows

### Pull Request Checks (`.github/workflows/pull-requests.yaml`)
Triggers on PRs to `main` or `vnext` branches when files in `backend/`, `frontend/`, or workflows change.

Runs 3 parallel jobs:
1. **Backend Tests** (`partial-backend.yaml`):
   - Go 1.24, golangci-lint (latest)
   - Runs: `task go:build` and `task go:coverage`
   - Timeout: 6m for linting
   
2. **Frontend Tests** (`partial-frontend.yaml`):
   - Linting: `pnpm run lint:ci` (max 1 warning)
   - Type checking: `pnpm run typecheck`
   - Integration tests with SQLite: `task test:ci`
   - Integration tests with PostgreSQL (matrix: v15, v16, v17): `task test:ci:postgresql`
   
3. **E2E Tests** (`e2e-partial.yaml`):
   - 4 sharded Playwright test runs
   - Timeout: 60 minutes
   - Runs: `task test:e2e --shard=N/4`
   - Uploads artifacts: test reports

**All CI checks must pass before merge.**

## Common Pitfalls & Workarounds

### 1. "command not found" errors
- **Missing `task`**: Install via `brew install go-task/tap/go-task` (Mac) or see [taskfile.dev](https://taskfile.dev)
- **Missing `swag`/`goose`**: Run `task setup` first
- **Missing `pnpm`**: Install via `npm install -g pnpm` or `brew install pnpm`

### 2. Code generation errors
- **"TypeSpecDef is nil" warnings**: Ignore - these are normal from swagger generation
- **Stale generated files**: Always run `task generate` after schema/API changes
- **Build failures after schema changes**: Must run `task generate` before `task go:build`

### 3. Test failures
- **Integration test race condition**: First run may fail - run again
- **Frontend tests fail**: Ensure backend is built: `cd backend && go build ./app/api`
- **E2E timeouts**: Increase wait time in `test:e2e` task (default 30s)
- **Port already in use**: Backend uses port 7745 - kill existing process

### 4. Linting issues
- **Go lint timeout**: CI uses `--timeout=6m` flag for golangci-lint
- **Frontend lint warnings**: Max 1 warning allowed in CI (`lint:ci`)
- **Auto-fix**: Use `task ui:fix` for frontend, golangci-lint auto-fixes Go

### 5. Database issues
- **SQLite locked**: Check for running processes, delete `.data/homebox.db-*` files
- **PostgreSQL tests**: Requires PostgreSQL service running on port 5432
- **Connection string**: Uses WAL mode and busy timeout for SQLite

### 6. Build artifacts
- **Cached builds**: Clean with `rm -rf build/ backend/app/api/static/public/ frontend/.nuxt`
- **.gitignore**: Build artifacts excluded: `build/`, `backend/api`, `.nuxt`, `.output`, etc.

## Environment Variables

### Backend Development
```bash
HBOX_LOG_LEVEL=debug                    # Log verbosity
HBOX_DATABASE_DRIVER=sqlite3            # Or postgres
HBOX_DATABASE_SQLITE_PATH=.data/homebox.db?_pragma=busy_timeout=1000&_pragma=journal_mode=WAL&_fk=1&_time_format=sqlite
HBOX_OPTIONS_ALLOW_REGISTRATION=true    # Allow user registration
UNSAFE_DISABLE_PASSWORD_PROJECTION=yes_i_am_sure  # For testing
HBOX_DEMO=true                          # Demo mode (set by task commands)
```

### PostgreSQL Configuration
```bash
HBOX_DATABASE_DRIVER=postgres
HBOX_DATABASE_USERNAME=homebox
HBOX_DATABASE_PASSWORD=homebox
HBOX_DATABASE_DATABASE=homebox
HBOX_DATABASE_HOST=localhost
HBOX_DATABASE_PORT=5432
HBOX_DATABASE_SSL_MODE=disable
```

**Note**: `Taskfile.yml` sets these automatically for task commands.

## Validation Checklist

Before submitting a PR, ensure:
- [ ] Run `task generate` after any schema/API changes
- [ ] Run `task pr` - all checks pass
- [ ] Run `task go:lint` - no linting errors
- [ ] Run `task ui:check` - no type errors
- [ ] Run `task go:test` - all Go tests pass
- [ ] Frontend linting: `cd frontend && pnpm run lint:ci` - max 1 warning
- [ ] No untracked build artifacts committed (check `.gitignore`)
- [ ] Changes match existing code style and patterns

## Quick Reference

**Start development environment**:
```bash
# Terminal 1 - Backend
task go:run

# Terminal 2 - Frontend  
task ui:dev
```

**Make API changes**:
1. Edit Go code in `backend/app/api/handlers/v1/`
2. Add/update Swagger comments
3. Run `task generate`
4. Run `task go:build` to verify
5. Run `task go:test`

**Make schema changes**:
1. Edit `backend/internal/data/ent/schema/*.go`
2. Run `task generate` (generates Ent code + types)
3. Update repo methods in `backend/internal/data/repo/`
4. Run `task go:test`

**Run specific tests**:
```bash
# Backend
cd backend && go test ./internal/data/repo -v

# Frontend
cd frontend && pnpm run test:watch
```

## Trust These Instructions

These instructions are validated and up-to-date. Only perform additional exploration if:
- Information is incomplete for your specific task
- Instructions are found to be incorrect
- You encounter an error not documented here

When in doubt, use `task --list-all` to see all available commands.

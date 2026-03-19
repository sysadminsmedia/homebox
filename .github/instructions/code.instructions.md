# Homebox Repository Instructions for Coding Agents

## Repository Overview

**Type**: Full-stack home inventory management web app (monorepo)  
**Size**: ~265 Go files, ~371 TypeScript/Vue files  
**Build Tool**: Task (Taskfile.yml) - **ALWAYS use `task` commands**  
**Database**: SQLite (default) or PostgreSQL

### Stack
- **Backend** (`/backend`): Go 1.26+, Chi router, Ent ORM, port 7745
- **Frontend** (`/frontend`): Nuxt 4, Vue 3, TypeScript, Tailwind CSS, pnpm 9.1.4+, dev proxies to backend

## Critical Build & Validation Commands

### Initial Setup (Run Once)
```bash
task setup  # Installs swag, goose, Go deps, pnpm deps
```

### Code Generation (Required Before Backend Work)
```bash
task generate  # Generates Ent ORM, Swagger docs, TypeScript types
```
**ALWAYS run after**: schema changes, API handler changes, before backend server/tests  
**Note**: "TypeSpecDef is nil" warnings are normal - ignore them

### Backend Commands
```bash
task go:build    # Build binary (60-90s)
task go:test     # Unit tests (5-10s)
task go:lint     # golangci-lint (6m timeout in CI)
task go:all      # Tidy + lint + test
task go:run      # Start server (SQLite)
task pr          # Full PR validation (3-5 min)
```

### Frontend Commands
```bash
task ui:dev    # Dev server port 3000
task ui:check  # Type checking
task ui:fix    # eslint --fix + prettier
task ui:watch  # Vitest watch mode
```
**Lint**: Max 1 warning in CI (`pnpm run lint:ci`)

### Testing
```bash
task test:ci            # Integration tests (15-30s + startup)
task test:e2e           # Playwright E2E (60s+ per shard, needs playwright install)
task pr                 # Full PR validation: generate + go:all + ui:check + ui:fix + test:ci (3-5 min)
```

## Project Structure

### Key Root Files
- `Taskfile.yml` - All commands (always use `task`)
- `docker-compose.yml`, `Dockerfile*` - Docker configs
- `CONTRIBUTING.md` - Contribution guidelines

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

**Patterns**: Schema/API changes → edit source → `task generate`. Never edit generated code in `ent/`.

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

**Patterns**: Auto-imports for `components/` and `composables/`. API types auto-generated - never edit manually.

## CI/CD Workflows

PR checks (`.github/workflows/pull-requests.yaml`) on `main`/`vnext`:
1. **Backend**: Go 1.24, golangci-lint, `task go:build`, `task go:coverage`
2. **Frontend**: Lint (max 1 warning), typecheck, `task test:ci` (SQLite + PostgreSQL v15-17)
3. **E2E**: 4 sharded Playwright runs (60min timeout)

All must pass before merge.

## Common Pitfalls

1. **Missing tools**: Run `task setup` first (installs swag, goose, deps)
2. **Stale generated code**: Always `task generate` after schema/API changes
3. **Test failures**: Integration tests may fail first run (race condition) - retry
4. **Port in use**: Backend uses 7745 - kill existing process
5. **SQLite locked**: Delete `.data/homebox.db-*` files
6. **Clean build**: `rm -rf build/ backend/app/api/static/public/ frontend/.nuxt`

## Environment Variables

Backend defaults in `Taskfile.yml`:
- `HBOX_LOG_LEVEL=debug`
- `HBOX_DATABASE_DRIVER=sqlite3` (or `postgres`)
- `HBOX_DATABASE_SQLITE_PATH=.data/homebox.db?_pragma=busy_timeout=1000&_pragma=journal_mode=WAL&_fk=1`
- PostgreSQL: `HBOX_DATABASE_*` vars for username/password/host/port/database

## Validation Checklist

Before PR:
- [ ] `task generate` after schema/API changes
- [ ] `task pr` passes (includes lint, test, typecheck)
- [ ] No build artifacts committed (check `.gitignore`)
- [ ] Code matches existing patterns

## Quick Reference

**Dev environment**: `task go:run` (terminal 1) + `task ui:dev` (terminal 2)

**API changes**: Edit handlers → add Swagger comments → `task generate` → `task go:build` → `task go:test`

**Schema changes**: Edit `ent/schema/*.go` → `task generate` → update repo methods → `task go:test`

**Specific tests**: `cd backend && go test ./path -v` or `cd frontend && pnpm run test:watch`

## Trust These Instructions

Instructions are validated and current. Only explore further if info is incomplete, incorrect, or you encounter undocumented errors. Use `task --list-all` for all commands.

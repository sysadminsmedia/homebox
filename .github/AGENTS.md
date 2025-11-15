This is a Go based repository with a VueJS client for the frontend built with Vite and Nuxt, with ShadCN. 

To make life easier, the use of a Taskfile is included for the majority of development commands.

Please follow these guidelines when contributing:

## Required Before Each Commit
- Generate Swagger Files: `task swag --force`
- Generate JS API Client: `task typescript-types --force`
- Lint Golang: `task go:lint`
- Lint frontend: `task ui:fix`

## Repository Structure
### Backend
- `backend/`: Contains the backend folders
- `backend/app`: Contains main app code including API endpoints
- `backend/internal/core`: Contains basic services such as currencies
- `backend/data`: Contains all information related to data, including `ent` schemas, repos, migrations, etc.
- `backend/data/migrations`: Contains migration data, the `sqlite3` sub-folder contains sqlite migrations, `postgres` sub-folder the postgres migrations, BOTH are REQUIRED.
- `backend/data/ent/schema`: Contains the actual `ent` data models.
- `backend/data/repo`: Contains the data repositories
- `backend/pkgs`: Contains general helper functions and services

### Frontend
- `frontend/`: Contains initial frontend files
- `frontend/components`: Contains the ShadCN components
- `frontend/locales`: Contains the i18n JSON for languages
- `frontend/pages`: Contains VueJS pages
- `frontend/test`: Contains Playwright setup
- `frontend/test/e2e`: Contains actual Playwright test files

### Docs
- `docs/`: Contains VitePress based documentation

## Key Guidelines
1. Follow best practices for the various programming languages
2. Maintain existing code structure and organization when possible
3. Use dependency injection when reasonable
4. Write tests for new functionality and after fixing bugs to validate they're fixed
5. Document changes to the `docs/` folder when appropriate
# Contributing

## We Develop with GitHub

We use GitHub to host code, to track issues and feature requests, as well as accept pull requests.

## Branch Flow

We use the `main` branch as the development branch. All PRs should be made to the `main` branch from a feature branch. To create a pull request, you can use the following steps:

1. Fork the repository and create a new branch from `main`.
2. If you've added code that should be tested, add tests.
3. If you've changed APIs, update the documentation.
4. Ensure that the test suite and linters pass
5. Issue your pull request

## How To Get Started

### Prerequisites

There is a devcontainer available for this project. If you are using VSCode, you can use the devcontainer to get started. If you are not using VSCode, you need to ensure that you have the following tools installed:

- [Go 1.19+](https://golang.org/doc/install)
- [Swaggo](https://github.com/swaggo/swag)
- [Node.js 16+](https://nodejs.org/en/download/)
- [pnpm](https://pnpm.io/installation)
- [Taskfile](https://taskfile.dev/#/installation) (Optional but recommended)
- For code generation, you'll need to have `python3` available on your path. In most cases, this is already installed and available.

If you're using `taskfile` you can run `task --list-all` for a list of all commands and their descriptions.

### Setup

If you're using the taskfile, you can use the `task setup` command to run the required setup commands. Otherwise, you can review the commands required in the `Taskfile.yml` file.

### API Development Notes

start command `task go:run`

1. API Server does not auto reload. You'll need to restart the server after making changes.
2. Unit tests should be written in Go, however, end-to-end or user story tests should be written in TypeScript using the client library in the frontend directory.

### Frontend Development Notes

start command `task ui:dev`

1. The frontend is a Vue 3 app with Nuxt.js that uses Tailwind and Shadcn-vue for styling.
2. We're using Vitest for our automated testing. You can run these with `task ui:watch`.
3. Tests require the API server to be running, and in some cases the first run will fail due to a race condition. If this happens, just run the tests again and they should pass.

## Publishing Release

Create a new tag in GitHub with the version number vX.X.X. This will trigger a new release to be created.

Test -> Goreleaser -> Publish Release -> Trigger Docker Builds -> Deploy Docs + Fly.io Demo

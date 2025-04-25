# Getting Started With Contributing

## Get Started

### Prerequisites

There is a devcontainer available for this project. If you are using VSCode, you can use the devcontainer to get started. If you are not using VSCode, you need to ensure that you have the following tools installed:

- [Go 1.23+](https://golang.org/doc/install)
- [Swaggo](https://github.com/swaggo/swag)
- [Node.js 16+](https://nodejs.org/en/download/)
- [pnpm](https://pnpm.io/installation)
- [Taskfile](https://taskfile.dev/#/installation) (Optional but recommended)
- For code generation, you'll need to have `python3` available on your path. In most cases, this is already installed and available.

If you're using `taskfile` you can run `task --list-all` for a list of all commands and their descriptions.

### Setup

If you're using the taskfile, you can use the `task setup` command to run the required setup commands. Otherwise, you can review the commands required in the `Taskfile.yml` file.

Note that when installing dependencies with pnpm, you must use the `--shamefully-hoist` flag. If you don't use this flag, you will get an error when running the frontend server.

### API Development Notes
start command `task go:run`

1. API Server does not auto reload. You'll need to restart the server after making changes.
2. Unit tests should be written in Go, however, end-to-end or user story tests should be written in TypeScript using the client library in the frontend directory.

test command `task go:test`

lint command `task go:lint`

swagger update command `task swag`

### Frontend Development Notes

start command `task ui:dev`

1. The frontend is a Vue 3 app with Nuxt.js that uses Tailwind and DaisyUI for styling.
2. We're using Vitest for our automated testing. You can run these with `task ui:watch`.
3. Tests require the API server to be running, and in some cases the first run will fail due to a race condition. If this happens, just run the tests again and they should pass.

fix/lint code `task ui:fix`

type checking `task ui:check`

## Documentation
We use [Vitepress](https://vitepress.dev/) for the web documentation of homebox. Anyone is welcome to contribute the documentation if they wish.
For documentation contributions, you only need Node.js and PNPM.

::: info Notes
- Languages are separated by folder (e.g `/en`, `/fr`, etc.)
- The Sidebar must be updated on a per language basis
- Each language's files can be named independently (slugs can match the language)
- The `public/_redirects` file is used to redirect the default to english
- Redirects can also be configured per language by adding `Language=` after the redirect code
:::

## Translations
We use our own [Weblate instance](https://translate.sysadminsmedia.com/projects/homebox/) for translations. If you would like to help translate Homebox, please visit the 
Weblate instance and help us translate the project. We accept translations for any language.

If you add a new language, please go to the English translation, press the `Add new translation string` button and then
use `languages.<language_code>` as the key. For example, if you are adding a French translation, the key would be `languages.fr`.
And then the string should be the name of the language in English. This is used to display the language in the language switcher.

[![Translation status](http://translate.sysadminsmedia.com/widget/homebox/multi-auto.svg)](http://translate.sysadminsmedia.com/engage/homebox/)

## Branch Flow
We use the `main` branch as the development branch. All PRs should be made to the `main` branch from a feature branch.
To create a pull request you can use the following steps:

1. Fork the repo and create a new branch from `main`
2. If you added code that should be tested, add tests
3. If you've changed APIs, update the documentation
4. Ensure that the test suite and linters pass
5. Create your PR


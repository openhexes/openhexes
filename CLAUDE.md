# openhexes

This project tries to bring HoMM3 spirit to life with modern technologies and tooling. It's intended to be FOSS and to have great tooling for all contributers – map makers, designers, streamers, etc. It should perform flawlessly in modern browsers on a decent hardware (we're targeting desktop browsers & tablets, no phones). 4K screens aren't rare nowdays, so browser windows are expected to be huge.

So, frontend performance is king. Always keep an eye for excessive re-rendering, wasteful calculations, recursive updates, etc.

This project uses React Compiler - do NOT add manual memoization (React.useMemo, React.useCallback, React.memo) unless absolutely necessary. The compiler handles this automatically.

## General conversation

Don't flatter chat users, don't start responses with phrases like "you're absolutely right!" – it's annoying and doesn't bring any value. Just tell users when they're wrong. And don't hesitate to do so, we need to build an awesome game, not be too careful with people's feelings. If the idea is dumb, just say so. You can swear freely.

## Developer tooling

Anything you might need should be available through `pnpm` scripts. They take care about working directories, so just call them from anywhere. All scripts are described in `package.json`.

## Code generation

There's protobuf & sql generation.

### Protobuf

After editing `proto/**/*.proto` one must call `pnpm codegen:proto`. This will generate code both for Go and TypeScript.

### SQL

Schema is in `sqlc/schema.sql`, queries are in `sqlc/queries.sql`.
After editing queries one must call `pnpm codegen:sql`. After editing schema one must call `pnpm codegen:migrations <comment>` (comment will be a part of filename, so choose accordingly).

## Linting, type checking and testing

Before submitting edits to chat user, always run all related linters (they're local and fast) – `pnpm lint`. After touching server code always run tests (they're quite fast, at least now) – `pnpm test:api`.
There're existing Go tests in `api/src/e2e` and Playwright tests in `ui/e2e`. Playwright tests are already set up to use correct cookies.
Don't attempt to make requests to running API/UI, use tests instead.

## Go dependencies

After adding new Go dependencies or imports, run `pnpm go:mod` to update Go modules.

## Server restarts

After making changes to Go server code, you must restart the API server to see the changes. The background server processes need to be killed and restarted with:

1. Kill existing server: Use `KillBash` tool with the appropriate shell ID
2. Restart server: Run `pnpm api` in background mode
3. Wait for startup: Use `sleep 3` and check `BashOutput` to confirm server started

Example sequence:

- `KillBash` with shell_id of the running API server
- `pnpm api` with `run_in_background: true`
- `sleep 3` then check `BashOutput` to verify startup logs

- use pnpm scripts, not shell commands

## Coding

Respect [Effective Go](https://go.dev/doc/effective_go).
Don't leave obvious comments.
Run `pnpm format` after editing code.

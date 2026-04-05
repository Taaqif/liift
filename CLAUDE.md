# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Liift is a workout tracking application with a Go backend and Vue 3 frontend served as a SPA from the same process.

## Commands

### Development
```bash
make dev          # Run both backend (air) and frontend (vite) in dev mode
```
This starts `air` for Go hot-reloading and Vite dev server concurrently. The frontend dev server proxies API calls to the Go backend.

### Production build
```bash
make build        # Builds frontend (npm run build) then compiles Go binary to ./bin/liift
```

### Backend only
```bash
go build -o ./tmp/go-liift .   # Build Go binary
go run .                        # Run without building
```

### Frontend only
```bash
cd web && npm run dev      # Vite dev server
cd web && npm run build    # TypeScript check + Vite build
```

## Architecture

### Backend (Go)

The backend is a single Go module (`liift`) using Echo v4 as the HTTP framework and GORM as the ORM.

**Entry point:** `main.go` — connects to DB, runs migrations, seeds reference data, registers handlers, starts Echo server.

**Layers:**
- `api/` — HTTP handlers and route registration. Each feature has a `*_handler.go` file and a `Register*Routes` function. All routes under `/api/` are registered in `api/api.go`.
- `api/middleware/` — JWT auth middleware (`RequireAuth`). Validates Bearer tokens, attaches user to context.
- `api/types/` — Shared response types (e.g. `ErrorResponse`).
- `internal/models/` — GORM models. All embed `BaseModel` (ID, CreatedAt, UpdatedAt, DeletedAt).
- `internal/repository/` — Data access layer. Each feature repo embeds `BaseRepository` which wraps `*gorm.DB`.
- `internal/database/` — DB connection (`postgres` or `sqlite` via `DB_DRIVER` env), auto-migration, and seeding of reference/enum tables.
- `internal/utils/` — Env helpers (`GetEnv`, `GetEnvAsInt`).

**Auth:** JWT tokens signed with `JWT_SECRET`. The middleware validates tokens and the user ID/username are embedded in JWT claims.

**Database:** Supports both PostgreSQL (default) and SQLite (set `DB_DRIVER=sqlite`). SQLite defaults to `./data/liift.db`. Migrations run automatically at startup via GORM `AutoMigrate`.

### Frontend (Vue 3)

Located in `web/`. Built with Vite, TypeScript, Vue 3 Composition API, Tailwind CSS v4, and Reka UI (headless component primitives).

**Key libraries:**
- `@tanstack/vue-query` — server state management; all API calls go through query/mutation hooks
- `vee-validate` + `zod` — form validation
- `vue-router` — client-side routing with auth guards
- `vue-i18n` — i18n (check `web/src/i18n/`)
- `vue-sonner` — toast notifications

**Structure:**
- `web/src/lib/api.ts` — `ApiClient` class; reads/writes JWT token from `localStorage`, attaches `Authorization: Bearer` header to all requests, handles 401 by clearing the token.
- `web/src/lib/queryKeys.ts` — centralized TanStack Query key factory; always use these keys when writing queries/mutations.
- `web/src/lib/auth/` — auth composables and types.
- `web/src/features/` — feature-scoped components, composables, and types (exercises, workouts, workout-plans, workout-session, reference).
- `web/src/views/` — top-level page components tied to routes.
- `web/src/components/ui/` — shared UI primitives.
- `web/src/router/index.ts` — routes with `requiresAuth` / `requiresGuest` meta guards.

**In dev mode:** Vite runs on a separate port; in production the Go binary serves the built `web/dist/` assets and the SPA fallback via `web/` package (`web.RegisterHandlers`).

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `3000` | HTTP server port |
| `JWT_SECRET` | `""` | JWT signing secret |
| `DB_DRIVER` | `postgres` | `postgres` or `sqlite` |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | PostgreSQL user |
| `DB_PASSWORD` | `""` | PostgreSQL password |
| `DB_NAME` | `liift` | PostgreSQL database name |
| `DB_SQLITE_PATH` | `./data/liift.db` | SQLite file path |
| `IMAGE_STORAGE_PATH` | `./storage/images` | Filesystem path for uploaded images |

A `.env` file in the project root is auto-loaded at startup.

## Adding a New Feature

The standard pattern for a new backend feature:
1. Add model(s) to `internal/models/`
2. Register model in `database.Migrate()` in `internal/database/migrate.go`
3. Add repository to `internal/repository/`
4. Add handler to `api/handlers/`
5. Register routes in `api/api.go`

For the frontend:
1. Add feature folder under `web/src/features/<feature>/` with composables, components, and types
2. Add query keys to `web/src/lib/queryKeys.ts`
3. Add route to `web/src/router/index.ts` and view to `web/src/views/`

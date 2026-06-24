# File Management System — engineering roadmap

Personal pet project: **clean architecture**, **performance**, **Google Drive–like UX** (once a real explorer exists).  
**Stack:** Go (Gin) + PostgreSQL (sqlx) + local disk storage. **No frontend in-repo yet.**

---

## Agent Brief

**What:** Go backend (Clean Architecture) — JWT auth, local file storage, PostgreSQL metadata, folder tree, async image→PDF via worker pool. **No frontend in repo.**

**Entry:** `server/cmd/main.go` → `go run ./server/cmd` from repo root (or `cd server && go run ./cmd` with `.env` in CWD).

**Stack:** Go 1.25 · Gin · sqlx/PostgreSQL · Viper · JWT (access only) · bcrypt · local disk · custom worker pool (5 goroutines).

**Layout:**
- `server/cmd/main.go` — DI, HTTP server, graceful shutdown (5s), worker pool start
- `server/internal/delivery/` — Gin router, handlers, JWT middleware
- `server/internal/service/` — business logic
- `server/internal/repository/postgres/` — SQL repos + `migrations/*.sql` (manual apply!)
- `server/internal/domain/` — entities + interfaces
- `server/internal/worker/` — Task/Pool
- `server/config/` — Viper `.env` loader

**API base:** `/api` — all routes below need `Authorization: Bearer <token>` except auth + `/api/test/ping`.

| Method | Path | Notes |
|--------|------|-------|
| POST | `/api/auth/register` | `{username, password, email}` |
| POST | `/api/auth/login` | `{username, password}` → `{token}` |
| POST | `/api/files/upload` | multipart `file`, optional `folder_id` |
| GET | `/api/files/` | list; query `folder_id` optional |
| GET | `/api/files/:id` | download attachment |
| DELETE | `/api/files/:id` | delete blob + row |
| POST | `/api/folders/` | `{name, parent_id?}` |
| DELETE | `/api/folders/:id` | |
| PATCH | `/api/folders/:id/rename` | `{name}` |
| GET | `/api/tree?parentId=` | folders + files at level |
| POST | `/api/convert/img-pdf/:id` | async conversion |
| GET | `/api/test/ping` | public health |

**Env** (`server/.env.example`): `PORT`, `JWT_SECRET`, `STORAGE_PATH`, `DB_DSN` (Postgres on `:5436` via docker-compose).

**Migrations:** SQL in `server/internal/repository/postgres/migrations/` — **no runner wired**; apply manually before first run.

**Milestone status:** M0 ✅ · M1 ✅ · M2 ✅ · M3 ❌ (no UI) · M4 🔄 (Docker API, OpenAPI, tests, S3 absent).

**Known gaps:** no refresh tokens; internal errors leak to client in some handlers; worker `Stop()` not called on shutdown; no job status for convert; no `_test.go`; docker-compose = Postgres only; filename typo `fodler_handler.go`.

**Do not trust old docs:** routes live under `/api`; entrypoint is `server/cmd/main.go`; no GORM, no `/auth/me`, no `/worker/status`.

---

## Status Quo

### Backend (Go API)

- ✅ Layering: `cmd` → [`server/internal/delivery`](server/internal/delivery) → [`server/internal/service`](server/internal/service) → [`server/internal/repository/postgres`](server/internal/repository/postgres) / [`memory`](server/internal/repository/memory); domain in [`server/internal/domain`](server/internal/domain).
- ✅ Config via Viper + `.env`: [`server/config/config.go`](server/config/config.go), sample [`server/.env.example`](server/.env.example).
- ✅ Postgres connection helper with pool limits: [`server/internal/repository/postgres/connection/connection.go`](server/internal/repository/postgres/connection/connection.go).
- ✅ Schema migration SQL: [`000001_init.up.sql`](server/internal/repository/postgres/migrations/000001_init.up.sql) (`users`, `file_metadata`), [`000002_folders.up.sql`](server/internal/repository/postgres/migrations/000002_folders.up.sql) (`folders`, `folder_id` on files).
- ✅ User auth: register/login with bcrypt + JWT — [`server/internal/service/user_service.go`](server/internal/service/user_service.go). Email validation via `mail.ParseAddress`; login returns generic `401 invalid credentials`. No refresh/revocation yet.
- ✅ File upload pipeline: stream to `STORAGE_PATH`, uuid stored filename — [`server/internal/service/file_service.go`](server/internal/service/file_service.go). Sets `Path`, `MimeType` (`application/octet-stream`), `CreatedAt`; optional `folder_id`; orphan blob cleanup on DB save failure. `Checksum` still empty.
- ✅ Repos: [`user_repo.go`](server/internal/repository/postgres/user_repo.go), [`file_repo.go`](server/internal/repository/postgres/file_repo.go), [`folder_repo.go`](server/internal/repository/postgres/folder_repo.go). Memory repos available but commented out in `main.go`.
- ✅ Worker pool: [`server/internal/worker/pool.go`](server/internal/worker/pool.go); convert task: [`server/internal/service/tasks.go`](server/internal/service/tasks.go).
- ✅ **Build green** — router wired in [`server/internal/delivery/handler.go`](server/internal/delivery/handler.go); `go build ./...` passes.
- 🔄 **Runtime hardening gaps** — see Technical Debt (error handling, worker lifecycle, job status).

### Frontend

- ❌ **No SPA or static UI in this repository.** Explorer / Drive-like UX = greenfield (add under e.g. `web/`).

---

## Infrastructure & DX

- ✅ **Single entrypoint:** [`server/cmd/main.go`](server/cmd/main.go) only.
- 🔄 **Docker Compose:** Postgres only in [`docker-compose.yml`](docker-compose.yml); API service not containerized yet.
- ❌ **Migrations runner:** wire golang-migrate / goose / embed SQL — today only `.sql` files, no automated apply from [`server/cmd`](server/cmd).
- 🔄 **Config paths:** [`server/config/config.go`](server/config/config.go) checks `.` and `../` for `.env` — still CWD-dependent.
- ❌ **API docs:** OpenAPI 3 + Swagger UI not present.
- 🔄 **Logging & tracing:** Gin default logger/recovery; no structured logs, request IDs, or unified error JSON — extend [`server/internal/delivery/middleware.go`](server/internal/delivery/middleware.go).
- 🔄 **Graceful shutdown:** `signal.Notify` + `http.Server.Shutdown` in `main.go`; worker `Stop()` not called on shutdown.
- ✅ **README accuracy:** [`README.md`](README.md) aligned with real paths, routes, and migration story.

---

## Core Engine (Priority 1) — files, storage, tree API

- ✅ **Fix build:** router and handlers wired; project builds cleanly.
- ✅ **Metadata vs disk:** `Path`, `MimeType`, `CreatedAt` set before save; orphan cleanup on DB failure. `Checksum` still unset.
- ✅ **Register vs DB:** email flows through domain, service, handler, repo; duplicate checks in place.
- 🔄 **Download contract:** [`file_handler.go`](server/internal/delivery/file_handler.go) uses `c.Param("id")`, ownership checks work, `c.FileAttachment` for download. Still maps non-access errors to `500`; no MIME from metadata headers beyond attachment default.
- ✅ **JWT middleware:** uses `h.jwtSecret` injected from config — [`middleware.go`](server/internal/delivery/middleware.go).
- 🔄 **CRUD completeness:** list + delete implemented; rename/move for files still absent.
- ✅ **Folder model:** `folders` table + repo + service + handlers; per-user isolation enforced.
- ✅ **Tree API:** `GET /api/tree?parentId=` returns folders + files; Postman-tested.
- ✅ **Authorization:** file/folder ops check `user_id` matches resource owner.

---

## Explorer UI (Priority 2) — Drive-like UX

*Depends on choosing a frontend stack and a `web/` app.*

- ❌ Auth client (login/register, token storage, Authorization header)
- ❌ Navigation (sidebar tree + main content)
- ❌ Breadcrumbs from folder path API
- ❌ Grid vs list views
- ❌ File icons / thumbnails
- ❌ Upload UX (progress, retry)

---

## Advanced Features (Priority 3)

- ❌ Search (full-text / trigram on filename)
- ❌ Preview (images/PDF in-browser)
- ❌ Bulk actions (multi-select delete/move)
- ❌ Drag & drop between folders
- ❌ Sharing (links, permissions, `shares` table)
- 🔄 **Async convert:** HTTP endpoint + worker submission wired; no persisted job status or lifecycle tracking.

---

## Technical Debt

- ✅ **Compile blockers:** resolved.
- ✅ **Duplicate `main`:** only `server/cmd/main.go` exists.
- ✅ **Handler bugs:** missing returns after errors fixed; login uses dedicated DTO.
- 🔄 **Security:** JWT secret wiring fixed; raw internal errors still returned in several handlers; no refresh/revocation.
- 🔄 **Storage:** orphan cleanup on DB failure done; no virus scan, quotas, or streaming checksum.
- 🔄 **DB / domain:** email mismatch resolved; `Checksum` still empty on upload.
- ❌ **Tests:** no `_test.go` files; add repo integration tests (testcontainers) + handler tests.
- 🔄 **Polish:** filename typo `fodler_handler.go`; debug `fmt.Println` in rename handler; README now accurate.
- 🔄 **Worker pool:** panic-on-submit after `Stop` fixed via mutex + `closed` flag; `Stop()` not invoked on server shutdown; no job drain coordination.

---

## Personal Milestone

| Milestone | Target outcome | Status |
|-----------|----------------|--------|
| M0 — **Green build** | `go build` clean | ✅ |
| M1 — **Trustworthy core** | Register/login + upload/download + list/delete E2E | ✅ |
| M2 — **Folders + tree** | Drive-like hierarchy in API | ✅ |
| M3 — **Explorer v1** | Usable UI on top of API | ❌ |
| M4 — **Polish** | Docker all-in-one, OpenAPI, logging, graceful worker shutdown, S3-ready storage | 🔄 |

---

## Last updated (2026-06-21)

- README and tasks synced: API routes under `/api`; entrypoint `server/cmd/main.go`; manual SQL migrations documented.
- File metadata: `Path`, `MimeType`, `CreatedAt` populated on upload; orphan blob cleanup on DB save failure.
- Folders + tree API complete (M2); endpoints tested via Postman.
- Build green; no `_test.go` yet.
- Next priorities: migration runner, API in Docker Compose, OpenAPI, tests, refresh tokens, worker shutdown on SIGTERM, convert job status, frontend (M3).

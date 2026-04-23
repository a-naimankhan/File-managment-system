# File Management System — engineering roadmap

Personal pet project: **clean architecture**, **performance**, **Google Drive–like UX** (once a real explorer exists).  
**Stack:** Go (Gin) + PostgreSQL (sqlx) + local disk storage. **No frontend in-repo yet.**

---

## Status Quo

### Backend (Go API)

- [x] Layering present: `cmd` → [`server/internal/delivery`](server/internal/delivery) → [`server/internal/service`](server/internal/service) → [`server/internal/repository/postgres`](server/internal/repository/postgres) / [`memory`](server/internal/repository/memory); domain types in [`server/internal/domain`](server/internal/domain).
- [x] Config via Viper + `.env`: [`server/config/config.go`](server/config/config.go), sample [`server/.env.example`](server/.env.example).
- [x] Postgres connection helper with pool limits: [`server/internal/repository/postgres/connection/conection.go`](server/internal/repository/postgres/connection/conection.go) (note typo: `conection`).
- [x] Schema migration SQL: [`server/internal/repository/postgres/migrations/000001_init.up.sql`](server/internal/repository/postgres/migrations/000001_init.up.sql) (`users`, `file_metadata`).
- [x] User auth flow (service): register/login with bcrypt + JWT signing using `cfg.JWTSECRET` — [`server/internal/service/user_service.go`](server/internal/service/user_service.go).
- [x] File upload pipeline (service): stream to `STORAGE_PATH`, uuid stored filename — [`server/internal/service/file_service.go`](server/internal/service/file_service.go).
- [x] Repos: [`server/internal/repository/postgres/user_repo.go`](server/internal/repository/postgres/user_repo.go), [`server/internal/repository/postgres/file_repo.go`](server/internal/repository/postgres/file_repo.go).
- [x] Worker pool skeleton: [`server/internal/worker/pool.go`](server/internal/worker/pool.go); convert task wrapper: [`server/internal/service/tasks.go`](server/internal/service/tasks.go).
- [ ] **API does not compile** — broken router block + placeholders in [`server/internal/delivery/handler.go`](server/internal/delivery/handler.go).
- [ ] **Runtime correctness gaps** — see Technical Debt (JWT verify vs sign, `path` / `email` / download route, error handling in handlers).

### Frontend

- [ ] **No SPA or static UI in this repository.** Explorer / Drive-like UX = greenfield (Angular, React, or other — pick one stack and add under e.g. `web/`).

---

## Infrastructure & DX

- [ ] **Single entrypoint:** remove or relocate Hello World [`server/cmd/main.go`](server/cmd/main.go); keep one `main` (likely [`server/cmd/app.go`](server/cmd/app.go)) or split `cmd/api` vs `cmd/migrate`.
- [ ] **Docker Compose:** add API service + shared network; fix Postgres volume path typo (`postgresgoql` → `postgresql`) in [`docker-compose.yml`](docker-compose.yml); document env vars for DB + app.
- [ ] **Migrations runner:** wire golang-migrate / goose / embed SQL — today only `.sql` file exists, no automated apply from [`server/cmd`](server/cmd).
- [ ] **Config paths:** [`server/config/config.go`](server/config/config.go) uses `../` for `.env` — make CWD-independent (e.g. `server/.env` or env-only in containers).
- [ ] **API docs:** add OpenAPI 3 + Swagger UI (e.g. `swaggo/swag` or manual `openapi.yaml` served under `/api/docs`).
- [ ] **Logging & tracing:** structured logs (slog/zap), request ID middleware, Gin recovery + consistent error JSON — extend [`server/internal/delivery/middleware.go`](server/internal/delivery/middleware.go).
- [ ] **Graceful shutdown:** `signal.Notify` + `http.Server.Shutdown` + worker `context` cancel (README already mentions this).
- [ ] **README accuracy:** align [`README.md`](README.md) with real paths (`go run` target), migration tool, and actual routes (`/api/...` prefix).

---

## Core Engine (Priority 1) — files, storage, tree API

- [ ] **Fix build:** repair [`server/internal/delivery/handler.go`](server/internal/delivery/handler.go) (`/convert` group, real handlers or remove until ready).
- [ ] **Metadata vs disk:** [`server/internal/service/file_service.go`](server/internal/service/file_service.go) must set `Path` (and ideally `MimeType`, `Checksum`, `CreatedAt`) before [`fileRepo.Save`](server/internal/repository/postgres/file_repo.go) — `path` is `NOT NULL` in SQL.
- [ ] **Register vs DB:** [`server/internal/service/user_service.go`](server/internal/service/user_service.go) omits `Email`; [`users`](server/internal/repository/postgres/migrations/000001_init.up.sql) requires unique `email` — registration will fail until API + domain align.
- [ ] **Download contract:** [`server/internal/delivery/file_handler.go`](server/internal/delivery/file_handler.go) — `GET /files/:id` should read `c.Param("id")`, not `PostForm`; return `404` must `return` after `400`; consider `Content-Disposition` / MIME from metadata.
- [ ] **JWT middleware:** [`server/internal/delivery/middleware.go`](server/internal/delivery/middleware.go) hardcodes `[]byte("jwt-secret")` — must use same secret as [`user_service`](server/internal/service/user_service.go) (inject from config).
- [ ] **CRUD completeness:** implement `GET /api/files` (list by `user_id`), `DELETE /api/files/:id` (delete row + blob), optional rename/move — extend [`domain.FileRepository`](server/internal/domain/file.go) + [`file_repo.go`](server/internal/repository/postgres/file_repo.go) + handlers.
- [ ] **Folder model:** add `folders` table + `parent_id` / `path` / `name`; nest files under folders; enforce per-user isolation — domain + migration + repo + service.
- [ ] **Tree API:** `GET /api/tree?parentId=` or materialized path listing; pagination; stable sort (name, modified).
- [ ] **Authorization:** ensure every file/folder op checks `user_id` matches resource owner (not only JWT presence).

---

## Explorer UI (Priority 2) — Drive-like UX

*Depends on choosing a frontend stack (e.g. Angular) and a `web/` app.*

- [ ] **Auth client:** login/register, store access token, attach `Authorization` header to API calls.
- [ ] **Navigation:** sidebar tree + main content; keyboard-friendly focus.
- [ ] **Breadcrumbs** from folder path API.
- [ ] **Grid vs list** views with persisted user preference (localStorage).
- [ ] **File icons** from mime/extension mapping; thumbnails later (Priority 3).
- [ ] **Upload UX:** multipart upload, progress, error retry; align with [`Upload`](server/internal/delivery/file_handler.go) contract.

---

## Advanced Features (Priority 3)

- [ ] **Search:** full-text or trigram on `filename` + metadata; optional Elasticsearch later.
- [ ] **Preview:** images/PDF in-browser; office docs = out of scope or external service.
- [ ] **Bulk actions:** multi-select delete/move; optimistic UI + batch API.
- [ ] **Drag & drop:** move between folders (client sends `parentId` updates); debounce server calls.
- [ ] **Sharing:** share links, permissions (view/edit), optional public tokens — new tables `shares`, `share_members`.
- [ ] **Async convert:** wire [`worker.Pool`](server/internal/worker/pool.go) from HTTP job endpoint; job status persisted; fix [`domain.FileService`](server/internal/domain/file.go) vs `*FileService` method signature drift (`ConvertImageToPDF` + `context`).

---

## Technical Debt

- [ done ] **Does not compile:** [`handler.go`](server/internal/delivery/handler.go) syntax + `toFill` placeholders.
- [ ] **Duplicate `main`:** [`server/cmd/main.go`](server/cmd/main.go) vs [`server/cmd/app.go`](server/cmd/app.go).
- [ ] **Handler bugs:** [`Register`/`Login`](server/internal/delivery/user_handler.go) missing `return` after error responses (double `JSON` write risk); [`Download`](server/internal/delivery/file_handler.go) wrong input binding + missing return on parse error.
- [ ] **Security:** JWT verify secret mismatch; leaking internal errors to clients in some paths; no refresh token / token revocation.
- [ ] **Storage:** no virus scan, no size quotas, no streaming checksum; failed DB insert after write leaves orphan files.
- [ ] **DB / domain mismatch:** `email` column vs registration payload; [`FileMetadata`](server/internal/domain/file.go) insert missing required fields.
- [ ] **Tests:** no `_test.go` files; add repo integration tests (testcontainers) + handler tests with mocked services.
- [ ] **Typos / polish:** package dir `conection`, README bilingual noise vs actionable docs; remove `fmt.Println` debug in [`Upload`](server/internal/delivery/file_handler.go).
- [ ] **Worker pool:** [`Stop`](server/internal/worker/pool.go) closes channel while workers may still send — risk of panic; coordinate shutdown properly.

---

## Personal Milestone

| Milestone | Target outcome | Key deliverables |
|-----------|----------------|------------------|
| M0 — **Green build** | `go test ./...` and `go build` clean | Fix router, JWT secret injection, download handler, single `main` |
| M1 — **Trustworthy core** | Register/login + upload/download + list/delete work end-to-end | Path/metadata fixes, user email story, authz checks |
| M2 — **Folders + tree** | Drive-like hierarchy in API | `folders` model, tree endpoint, migration |
| M3 — **Explorer v1** | Usable UI on top of API | Frontend app: nav, breadcrumbs, grid/list, upload |
| M4 — **Polish** | Production-ish DX | Docker all-in-one, OpenAPI, logging, graceful shutdown, S3-ready storage abstraction |

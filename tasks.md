# File Management System — engineering roadmap

Personal pet project: **clean architecture**, **performance**, **Google Drive–like UX** (once a real explorer exists).  
**Stack:** Go (Gin) + PostgreSQL (sqlx) + local disk storage. **No frontend in-repo yet.**

---

## Status Quo

### Backend (Go API)

- ✅ Layering present: `cmd` → [`server/internal/delivery`](server/internal/delivery) → [`server/internal/service`](server/internal/service) → [`server/internal/repository/postgres`](server/internal/repository/postgres) / [`memory`](server/internal/repository/memory); domain types in [`server/internal/domain`](server/internal/domain). Notes: architectural layering is still clear across `cmd`, delivery, service, repository, and domain packages.
- ✅ Config via Viper + `.env`: [`server/config/config.go`](server/config/config.go), sample [`server/.env.example`](server/.env.example). Notes: config loading is present and wired through startup.
- ✅ Postgres connection helper with pool limits: [`server/internal/repository/postgres/connection/conection.go`](server/internal/repository/postgres/connection/conection.go) (note typo: `conection`). Notes: helper opens, pings, and configures the pool; directory typo remains.
- ✅ Schema migration SQL: [`server/internal/repository/postgres/migrations/000001_init.up.sql`](server/internal/repository/postgres/migrations/000001_init.up.sql) (`users`, `file_metadata`). Notes: both core tables are defined in SQL.
- 🔄 User auth flow (service): register/login with bcrypt + JWT signing using `cfg.JWTSECRET` — [`server/internal/service/user_service.go`](server/internal/service/user_service.go). Notes: email support and `mail.ParseAddress` validation are in place, but `Login` still binds `registerRequest` in the handler and requires `email` on login requests.
- 🔄 File upload pipeline (service): stream to `STORAGE_PATH`, uuid stored filename — [`server/internal/service/file_service.go`](server/internal/service/file_service.go). Notes: upload now stores `Path` before `Save`, but `MimeType`, `Checksum`, and `CreatedAt` are still not populated.
- 🔄 Repos: [`server/internal/repository/postgres/user_repo.go`](server/internal/repository/postgres/user_repo.go), [`server/internal/repository/postgres/file_repo.go`](server/internal/repository/postgres/file_repo.go). Notes: Postgres repos were extended for email and list support, but memory repos no longer satisfy the updated interfaces.
- ✅ Worker pool skeleton: [`server/internal/worker/pool.go`](server/internal/worker/pool.go); convert task wrapper: [`server/internal/service/tasks.go`](server/internal/service/tasks.go). Notes: worker pool and conversion task wrapper are both present.
- 🔄 **API does not compile** — broken router block + placeholders in [`server/internal/delivery/handler.go`](server/internal/delivery/handler.go). Notes: router/handler placeholders are fixed, but the repo still fails `go build ./...` because memory repositories are missing `GetByEmail` and `ListByUserId`.
- 🔄 **Runtime correctness gaps** — see Technical Debt (JWT verify vs sign, `path` / `email` / download route, error handling in handlers). Notes: JWT secret wiring, email registration, and download ID parsing were fixed, but several handler, storage, and ownership gaps remain.

### Frontend

- ❌ **No SPA or static UI in this repository.** Explorer / Drive-like UX = greenfield (Angular, React, or other — pick one stack and add under e.g. `web/`). Notes: backend-only repo; no frontend app has been added.

---

## Infrastructure & DX

- ✅ **Single entrypoint:** remove or relocate Hello World [`server/cmd/main.go`](server/cmd/main.go); keep one `main` (likely [`server/cmd/app.go`](server/cmd/app.go)) or split `cmd/api` vs `cmd/migrate`. Notes: there is only one entrypoint now in `server/cmd/main.go`; `server/cmd/app.go` does not exist.
- 🔄 **Docker Compose:** add API service + shared network; fix Postgres volume path typo (`postgresgoql` → `postgresql`) in [`docker-compose.yml`](docker-compose.yml); document env vars for DB + app. Notes: Compose file exists for Postgres only; API service and fuller app docs are still missing.
- ❌ **Migrations runner:** wire golang-migrate / goose / embed SQL — today only `.sql` file exists, no automated apply from [`server/cmd`](server/cmd). Notes: migration SQL exists, but no runner command is wired.
- ❌ **Config paths:** [`server/config/config.go`](server/config/config.go) uses `../` for `.env` — make CWD-independent (e.g. `server/.env` or env-only in containers). Notes: config loading still depends on the current working directory.
- ❌ **API docs:** add OpenAPI 3 + Swagger UI (e.g. `swaggo/swag` or manual `openapi.yaml` served under `/api/docs`). Notes: no OpenAPI spec or docs route exists.
- 🔄 **Logging & tracing:** structured logs (slog/zap), request ID middleware, Gin recovery + consistent error JSON — extend [`server/internal/delivery/middleware.go`](server/internal/delivery/middleware.go). Notes: Gin default logger/recovery and basic `log` usage exist, but not structured logging, request IDs, or unified error responses.
- 🔄 **Graceful shutdown:** `signal.Notify` + `http.Server.Shutdown` + worker `context` cancel (README already mentions this). Notes: `main.go` now uses `http.Server`, `signal.Notify`, and `context.WithCancel`, but worker shutdown coordination is still basic.
- ❌ **README accuracy:** align [`README.md`](README.md) with real paths (`go run` target), migration tool, and actual routes (`/api/...` prefix). Notes: README still points to `cmd/app/main.go` and documents routes without the real `/api` prefix.

---

## Core Engine (Priority 1) — files, storage, tree API

- 🔄 **Fix build:** repair [`server/internal/delivery/handler.go`](server/internal/delivery/handler.go) (`/convert` group, real handlers or remove until ready). Notes: router and convert endpoints are wired, but the current build is blocked by memory repository interface drift rather than delivery code.
- 🔄 **Metadata vs disk:** [`server/internal/service/file_service.go`](server/internal/service/file_service.go) must set `Path` (and ideally `MimeType`, `Checksum`, `CreatedAt`) before [`fileRepo.Save`](server/internal/repository/postgres/file_repo.go) — `path` is `NOT NULL` in SQL. Notes: `Path` is now populated before save; the remaining metadata fields are still unset.
- ✅ **Register vs DB:** [`server/internal/service/user_service.go`](server/internal/service/user_service.go) omits `Email`; [`users`](server/internal/repository/postgres/migrations/000001_init.up.sql) requires unique `email` — registration will fail until API + domain align. Notes: email now flows through domain, service, handler, and repository, and duplicate email checks are implemented.
- 🔄 **Download contract:** [`server/internal/delivery/file_handler.go`](server/internal/delivery/file_handler.go) — `GET /files/:id` should read `c.Param("id")`, not `PostForm`; return `404` must `return` after `400`; consider `Content-Disposition` / MIME from metadata. Notes: ID parsing and early returns are fixed; response headers and MIME handling are still not implemented.
- ✅ **JWT middleware:** [`server/internal/delivery/middleware.go`](server/internal/delivery/middleware.go) hardcodes `[]byte("jwt-secret")` — must use same secret as [`user_service`](server/internal/service/user_service.go) (inject from config). Notes: middleware now uses `h.jwtSecret`, which is injected from config.
- 🔄 **CRUD completeness:** implement `GET /api/files` (list by `user_id`), `DELETE /api/files/:id` (delete row + blob), optional rename/move — extend [`domain.FileRepository`](server/internal/domain/file.go) + [`file_repo.go`](server/internal/repository/postgres/file_repo.go) + handlers. Notes: list and delete endpoints plus repo/service support are now present, and delete removes the blob before deleting metadata; rename/move are still absent and memory repo support is incomplete.
- ❌ **Folder model:** add `folders` table + `parent_id` / `path` / `name`; nest files under folders; enforce per-user isolation — domain + migration + repo + service. Notes: no folder model or migration exists yet.
- ❌ **Tree API:** `GET /api/tree?parentId=` or materialized path listing; pagination; stable sort (name, modified). Notes: no tree endpoint or tree service logic exists yet.
- 🔄 **Authorization:** ensure every file/folder op checks `user_id` matches resource owner (not only JWT presence). Notes: delete now enforces file ownership, but download/convert still do not validate ownership against the resource.

---

## Explorer UI (Priority 2) — Drive-like UX

*Depends on choosing a frontend stack (e.g. Angular) and a `web/` app.*

- ❌ **Auth client:** login/register, store access token, attach `Authorization` header to API calls. Notes: no frontend client exists in-repo.
- ❌ **Navigation:** sidebar tree + main content; keyboard-friendly focus. Notes: no frontend client exists in-repo.
- ❌ **Breadcrumbs** from folder path API. Notes: no frontend client or folder path API exists.
- ❌ **Grid vs list** views with persisted user preference (localStorage). Notes: no frontend client exists in-repo.
- ❌ **File icons** from mime/extension mapping; thumbnails later (Priority 3). Notes: no frontend client exists in-repo.
- ❌ **Upload UX:** multipart upload, progress, error retry; align with [`Upload`](server/internal/delivery/file_handler.go) contract. Notes: upload endpoint exists, but there is no UI/client implementation.

---

## Advanced Features (Priority 3)

- ❌ **Search:** full-text or trigram on `filename` + metadata; optional Elasticsearch later. Notes: no search endpoint or indexing exists yet.
- ❌ **Preview:** images/PDF in-browser; office docs = out of scope or external service. Notes: no preview-specific behavior exists yet.
- ❌ **Bulk actions:** multi-select delete/move; optimistic UI + batch API. Notes: no batch APIs or UI exist yet.
- ❌ **Drag & drop:** move between folders (client sends `parentId` updates); debounce server calls. Notes: no folder model or frontend drag-and-drop exists yet.
- ❌ **Sharing:** share links, permissions (view/edit), optional public tokens — new tables `shares`, `share_members`. Notes: no sharing schema or API exists yet.
- 🔄 **Async convert:** wire [`worker.Pool`](server/internal/worker/pool.go) from HTTP job endpoint; job status persisted; fix [`domain.FileService`](server/internal/domain/file.go) vs `*FileService` method signature drift (`ConvertImageToPDF` + `context`). Notes: HTTP endpoint and worker submission are wired, and the handler now returns success; persisted job status and broader lifecycle tracking are still missing.

---

## Technical Debt

- 🔄 **Does not compile:** [`handler.go`](server/internal/delivery/handler.go) syntax + `toFill` placeholders. Notes: the original delivery compile issue is fixed, but the project currently fails to build because the memory repositories were not updated to the new interfaces.
- ✅ **Duplicate `main`:** [`server/cmd/main.go`](server/cmd/main.go) vs [`server/cmd/app.go`](server/cmd/app.go). Notes: only `server/cmd/main.go` exists now.
- 🔄 **Handler bugs:** [`Register`/`Login`](server/internal/delivery/user_handler.go) missing `return` after error responses (double `JSON` write risk); [`Download`](server/internal/delivery/file_handler.go) wrong input binding + missing return on parse error. Notes: the listed `return` and download binding issues are fixed, but `Login` still reuses `registerRequest` and requires `email`.
- 🔄 **Security:** JWT verify secret mismatch; leaking internal errors to clients in some paths; no refresh token / token revocation. Notes: JWT secret wiring is fixed, but raw internal errors are still returned in several handlers and there is still no refresh/revocation flow.
- ❌ **Storage:** no virus scan, no size quotas, no streaming checksum; failed DB insert after write leaves orphan files. Notes: none of the listed storage hardening items are implemented yet.
- 🔄 **DB / domain mismatch:** `email` column vs registration payload; [`FileMetadata`](server/internal/domain/file.go) insert missing required fields. Notes: email mismatch is resolved, and `Path` is now set, but file metadata is still incomplete.
- ❌ **Tests:** no `_test.go` files; add repo integration tests (testcontainers) + handler tests with mocked services. Notes: there are still no `_test.go` files in the repo.
- 🔄 **Typos / polish:** package dir `conection`, README bilingual noise vs actionable docs; remove `fmt.Println` debug in [`Upload`](server/internal/delivery/file_handler.go). Notes: `fmt.Println` was removed from upload, but the `conection` directory typo and README cleanup remain.
- ❌ **Worker pool:** [`Stop`](server/internal/worker/pool.go) closes channel while workers may still send — risk of panic; coordinate shutdown properly. Notes: worker pool shutdown logic has not been reworked yet.

---

## Personal Milestone

| Milestone | Target outcome | Key deliverables |
|-----------|----------------|------------------|
| M0 — **Green build** | `go test ./...` and `go build` clean | Fix router, JWT secret injection, download handler, single `main` |
| M1 — **Trustworthy core** | Register/login + upload/download + list/delete work end-to-end | Path/metadata fixes, user email story, authz checks |
| M2 — **Folders + tree** | Drive-like hierarchy in API | `folders` model, tree endpoint, migration |
| M3 — **Explorer v1** | Usable UI on top of API | Frontend app: nav, breadcrumbs, grid/list, upload |
| M4 — **Polish** | Production-ish DX | Docker all-in-one, OpenAPI, logging, graceful shutdown, S3-ready storage abstraction |

---

## Last updated

- Reflected the new email registration flow across domain, service, handler, and Postgres repository, including `mail.ParseAddress` validation and `GetByEmail`.
- Updated file lifecycle status to match the current code: `Path` is now persisted on upload, list/delete endpoints exist, delete removes the local blob first, and upload debug printing is gone.
- Recorded the new graceful shutdown work in `main.go`, including `http.Server`, `signal.Notify`, and worker-context cancellation wiring.
- Marked remaining gaps that still block a clean build or full completion, especially the out-of-date memory repositories, incomplete file metadata population, missing ownership checks on some file operations, and absent tests/docs/frontend work.

# File Management System — engineering roadmap

Personal pet project: **clean architecture**, **performance**, **Google Drive–like UX** (once a real explorer exists).  
**Stack:** Go (Gin) + PostgreSQL (sqlx) + local disk storage. **No frontend in-repo yet.**

---

## Status Quo

### Backend (Go API)

- ✅ Layering present: `cmd` → [`server/internal/delivery`](server/internal/delivery) → [`server/internal/service`](server/internal/service) → [`server/internal/repository/postgres`](server/internal/repository/postgres) / [`memory`](server/internal/repository/memory); domain types in [`server/internal/domain`](server/internal/domain). Notes: architectural layering is still clear across `cmd`, delivery, service, repository, and domain packages.
- ✅ Config via Viper + `.env`: [`server/config/config.go`](server/config/config.go), sample [`server/.env.example`](server/.env.example). Notes: config loading is present and wired through startup.
- ✅ Postgres connection helper with pool limits: [`server/internal/repository/postgres/connection/connection.go`](server/internal/repository/postgres/connection/conection.go) (note typo: `conection`). Notes: helper opens, pings, and configures the pool; directory typo remains.
- ✅ Schema migration SQL: [`server/internal/repository/postgres/migrations/000001_init.up.sql`](server/internal/repository/postgres/migrations/000001_init.up.sql) (`users`, `file_metadata`). Notes: both core tables are defined in SQL.
- 🔄 User auth flow (service): register/login with bcrypt + JWT signing using `cfg.JWTSECRET` — [`server/internal/service/user_service.go`](server/internal/service/user_service.go). Notes: email support and `mail.ParseAddress` validation are in place, and login now uses a dedicated DTO with generic `401 invalid credentials`; registration/login still lack broader auth features like refresh/revocation.
- ✅ File upload pipeline (service): stream to `STORAGE_PATH`, uuid stored filename — [`server/internal/service/file_service.go`](server/internal/service/file_service.go). Notes: upload now stores `Path` before `Save`. UploadFile accepts optional `folder_id`; `MimeType`, `Checksum`, and `CreatedAt` still not populated.
- 🔄 Repos: [`server/internal/repository/postgres/user_repo.go`](server/internal/repository/postgres/user_repo.go), [`server/internal/repository/postgres/file_repo.go`](server/internal/repository/postgres/file_repo.go). Notes: Postgres and memory repos now satisfy the updated interfaces, but the in-memory file repo still has rough edges like returning an error for an empty file list.
- ✅ Worker pool skeleton: [`server/internal/worker/pool.go`](server/internal/worker/pool.go); convert task wrapper: [`server/internal/service/tasks.go`](server/internal/service/tasks.go). Notes: worker pool and conversion task wrapper are both present.
- ✅ **API does not compile** — broken router block + placeholders in [`server/internal/delivery/handler.go`](server/internal/delivery/handler.go). Notes: router/handler placeholders are fixed, memory repo interface mismatch is resolved, and `go build ./...` completes cleanly.
- 🔄 **Runtime correctness gaps** — see Technical Debt (remaining error handling, storage consistency, and worker lifecycle). Notes: JWT secret wiring, email registration, and download route parsing are fixed; ownership checks now exist for download/convert, but multiple production-hardening gaps remain.

### Frontend

- ❌ **No SPA or static UI in this repository.** Explorer / Drive-like UX = greenfield (Angular, React, or other — pick one stack and add under e.g. `web/`). Notes: backend-only repo; no frontend app has been added.

---

## Infrastructure & DX

- ✅ **Single entrypoint:** remove or relocate Hello World [`server/cmd/main.go`](server/cmd/main.go); keep one `main` (likely [`server/cmd/app.go`](server/cmd/app.go)) or split `cmd/api` vs `cmd/migrate`. Notes: there is only one entrypoint now in `server/cmd/main.go`; `server/cmd/app.go` does not exist.
- 🔄 **Docker Compose:** add API service + shared network; fix Postgres volume path typo (`postgresgoql` → `postgresql`) in [`docker-compose.yml`](docker-compose.yml); document env vars for DB + app. Notes: Compose file exists for Postgres only; API service and fuller app docs are still missing.
- ❌ **Migrations runner:** wire golang-migrate / goose / embed SQL — today only `.sql` file exists, no automated apply from [`server/cmd`](server/cmd). Notes: migration SQL exists, but no runner command is wired.
- 🔄 **Config paths:** [`server/config/config.go`](server/config/config.go) uses `../` for `.env` — make CWD-independent (e.g. `server/.env` or env-only in containers). Notes: loader now checks both `.` and `../`, but startup is still CWD-dependent.
- ❌ **API docs:** add OpenAPI 3 + Swagger UI (e.g. `swaggo/swag` or manual `openapi.yaml` served under `/api/docs`). Notes: no OpenAPI spec or docs route exists.
- 🔄 **Logging & tracing:** structured logs (slog/zap), request ID middleware, Gin recovery + consistent error JSON — extend [`server/internal/delivery/middleware.go`](server/internal/delivery/middleware.go). Notes: Gin default logger/recovery and basic `log` usage exist, but not structured logging, request IDs, or unified error responses.
- 🔄 **Graceful shutdown:** `signal.Notify` + `http.Server.Shutdown` + worker `context` cancel (README already mentions this). Notes: `main.go` now uses `http.Server`, `signal.Notify`, and `context.WithCancel`, but worker shutdown coordination is still basic.
- ❌ **README accuracy:** align [`README.md`](README.md) with real paths (`go run` target), migration tool, and actual routes (`/api/...` prefix). Notes: README still points to `cmd/app/main.go` and documents routes without the real `/api` prefix.

---

## Core Engine (Priority 1) — files, storage, tree API

- ✅ **Fix build:** repair [`server/internal/delivery/handler.go`](server/internal/delivery/handler.go) (`/convert` group, real handlers or remove until ready). Notes: router and convert endpoints are wired, the memory repo interface drift is fixed, and the project builds cleanly.
- 🔄 **Metadata vs disk:** [`server/internal/service/file_service.go`](server/internal/service/file_service.go) must set `Path` (and ideally `MimeType`, `Checksum`, `CreatedAt`) before [`fileRepo.Save`](server/internal/repository/postgres/file_repo.go) — `path` is `NOT NULL` in SQL. Notes: `Path` is populated before save and orphan file cleanup now removes the blob if `fileRepo.Save` fails; `MimeType`, `Checksum`, and `CreatedAt` are still unset.
- ✅ **Register vs DB:** [`server/internal/service/user_service.go`](server/internal/service/user_service.go) omits `Email`; [`users`](server/internal/repository/postgres/migrations/000001_init.up.sql) requires unique `email` — registration will fail until API + domain align. Notes: email now flows through domain, service, handler, and repository, and duplicate email checks are implemented.
- 🔄 **Download contract:** [`server/internal/delivery/file_handler.go`](server/internal/delivery/file_handler.go) — `GET /files/:id` should read route param and return correct status codes; consider `Content-Disposition` / MIME from metadata. Notes: ownership checks are in place and the service now returns `access denied` consistently, but the handler still reads `c.Param("file")` while routes declare `:id`, still maps non-access errors to `500`, and still lacks download headers.
- ✅ **JWT middleware:** [`server/internal/delivery/middleware.go`](server/internal/delivery/middleware.go) hardcodes `[]byte("jwt-secret")` — must use same secret as [`user_service`](server/internal/service/user_service.go) (inject from config). Notes: middleware now uses `h.jwtSecret`, which is injected from config.
- 🔄 **CRUD completeness:** implement `GET /api/files` (list by `user_id`), `DELETE /api/files/:id` (delete row + blob), optional rename/move — extend [`domain.FileRepository`](server/internal/domain/file.go) + [`file_repo.go`](server/internal/repository/postgres/file_repo.go) + handlers. Notes: list and delete endpoints plus repo/service support are now present, and delete removes the blob before deleting metadata; rename/move are still absent.
- ✅ **Folder model:** add `folders` table + `parent_id` / `path` / `name`; nest files under folders; enforce per-user isolation — domain + migration + repo + service. Notes: folder domain, repo, service and migration implemented; endpoints tested via Postman.
- ✅ **Tree API:** `GET /api/tree?parentId=` or materialized path listing; pagination; stable sort (name, modified). Notes: tree endpoint implemented and tested.
- ✅ **Authorization:** ensure every file/folder op checks `user_id` matches resource owner (not only JWT presence). Notes: ownership checks are implemented in service paths for download, delete, and convert operations.


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

- ✅ **Does not compile:** [`handler.go`](server/internal/delivery/handler.go) syntax + `toFill` placeholders. Notes: the original delivery compile issue is fixed, the memory repo interface mismatch is resolved, and `go build ./...` is green.
- ✅ **Duplicate `main`:** [`server/cmd/main.go`](server/cmd/main.go) vs [`server/cmd/app.go`](server/cmd/app.go). Notes: only `server/cmd/main.go` exists now.
- ✅ **Handler bugs:** [`Register`/`Login`](server/internal/delivery/user_handler.go) missing `return` after error responses (double `JSON` write risk); [`Download`](server/internal/delivery/file_handler.go) wrong input binding + missing return on parse error. Notes: handler returns are fixed, login has a separate request DTO, and invalid credentials now map to HTTP 401 with a generic message.
- 🔄 **Security:** JWT verify secret mismatch; leaking internal errors to clients in some paths; no refresh token / token revocation. Notes: JWT secret wiring is fixed, but raw internal errors are still returned in several handlers and there is still no refresh/revocation flow.
- 🔄 **Storage:** no virus scan, no size quotas, no streaming checksum; failed DB insert after write leaves orphan files. Notes: orphan file cleanup on DB save failure is now implemented with `os.Remove(finalPath)`, but the rest of the storage hardening work is still missing.
- 🔄 **DB / domain mismatch:** `email` column vs registration payload; [`FileMetadata`](server/internal/domain/file.go) insert missing required fields. Notes: email mismatch is resolved, and `Path` is now set, but file metadata is still incomplete.
- ❌ **Tests:** no `_test.go` files; add repo integration tests (testcontainers) + handler tests with mocked services. Notes: there are still no `_test.go` files in the repo.
- 🔄 **Typos / polish:** package dir `conection`, README bilingual noise vs actionable docs; remove `fmt.Println` debug in [`Upload`](server/internal/delivery/file_handler.go). Notes: `fmt.Println` was removed from upload, but the `conection` directory typo and README cleanup remain.
- 🔄 **Worker pool:** [`Stop`](server/internal/worker/pool.go) closes channel while workers may still send — risk of panic; coordinate shutdown properly. Notes: panic-on-submit after `Stop` is fixed via `sync.Mutex`, `closed` flag, and setting `closed=true` before `close(p.tasks)`, but broader worker drain/shutdown coordination is still basic.

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

- Updated roadmap notes after the memory repository fixes: `ListByUserID` now matches the interface, `GetByEmail` returns `nil, nil` when a user is absent, and build-related compile blockers are cleared.
- Reflected the new storage consistency safeguard: `UploadFile` now removes the just-written blob if `fileRepo.Save` fails, so orphan files are no longer left behind on DB write failure.
- Updated worker pool notes to capture the `sync.Mutex` + `closed` flag protection in `Submit` and the `closed=true` before `close(p.tasks)` change in `Stop`, which removes the previous panic-on-submit risk.
- Reflected the auth/download polish from this session: login now returns generic `401 invalid credentials`, and `DownloadFile` now uses the same `access denied` message shape expected by the handler.
- Folder subsystem completed: M2 Folders complete (domain, repo, service, handlers, routes, migration). UploadFile accepts optional `folder_id`. RenameFolder param bug fixed. folder_repo Save query typo fixed (`%5` -> `$5`). All folder endpoints tested via Postman.

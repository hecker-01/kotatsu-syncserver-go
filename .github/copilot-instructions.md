# Copilot instructions for kotatsu-syncserver-go

## Build, test, and verification commands

- Build all packages: `go build ./...`
- Run all tests: `go test ./...`
- Run tests for a single package: `go test ./services` (replace with target package path)
- Run a single test by name: `go test ./services -run TestName`
- Additional static check used in this repo context: `go vet ./...`
- Build Docker image used by this project: `docker build -t kotatsu-syncserver-go .`

## High-level architecture

This is a layered Go HTTP API server using `chi` and MySQL:

1. `main.go` bootstraps configuration (`.env` loading + required env validation), initializes `logger` and `db`, then configures the top-level router.
2. Global middlewares are mounted on the root router: `middleware.StructuredLogger` and `chi` recoverer.
3. `/api` is mounted as a subrouter with a general IP rate limiter (`middleware.NewRateLimiter(100, 5*time.Minute)`).
4. `routes.RegisterAPIRoutes` composes domain route groups (`/auth`, `/users`, `/games`, `/history`, `/player`).
5. Route handlers are methods on controllers in `controllers/`.
6. Controllers delegate business logic to services in `services/`.
7. Services use shared infrastructure (`db.DB`, `utils` helpers) and return typed errors that controllers map to HTTP responses.

Auth flow is split across middleware and utils:

- `services.AuthService` handles register/login and JWT creation via `utils.GenerateJWT`.
- `middleware.RequireAuth` validates bearer tokens via `utils.ParseAndValidateJWT` and injects `user_id` into context.
- Authenticated services/controllers read user identity from context with `utils.UserIDFromContext`.

## Key repository conventions

- When making code changes, document what was done in the touched files: add/update concise Go doc comments, inline comments for non-obvious logic, and/or adjacent documentation as appropriate. Do not leave behavior changes undocumented.
- Keep request/response formatting centralized through `utils.WriteJSON` and `utils.WriteError`; controllers use these consistently instead of `http.Error`.
- Preserve the controller-service split: controllers translate HTTP concerns, services implement domain logic and return sentinel errors (for example `services.ErrInvalidInput`, `services.ErrInvalidCredentials`) that controllers map to status codes.
- Add protected endpoints by applying `middleware.RequireAuth` at route-group level (see `routes/user.go`, `routes/player.go`, `routes/history.go`).
- Rate limiting is intentionally layered:
  - global API limiter in `main.go`
  - stricter auth limiter in `routes/auth.go` (`5` requests per `5m`)
- Use `utils` for JWT and auth context plumbing rather than re-implementing token parsing/signing in handlers.
- `handlers/` contains older direct HTTP handlers; active routing uses `controllers/` + `services/`. Prefer the controller/service path for new changes.
- Environment configuration is required at startup (`DB_HOST`, `DB_NAME`, `DB_USER`, `DB_PASS`, `JWT_SECRET`) and validated before server start.
- Database access currently uses package-global `db.DB` initialized by `db.Init()`. New DB calls in services should use that shared connection.

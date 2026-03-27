# Copilot instructions for kotatsu-syncserver-go

## Build, test, and verification commands

- **Build all packages**: `go build ./...`
- **Run all tests**: `go test ./...`
- **Run tests with verbose output**: `go test -v ./...`
- **Run short tests only** (skips integration tests): `go test -short ./...`
- **Run tests for a single package**: `go test ./controllers` (replace with target package path)
- **Run a single test by name**: `go test ./controllers -run TestAuthEndpoint`
- **Run tests with coverage**: `go test -cover ./...`
- **Generate coverage report**: `go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out`
- **Static analysis**: `go vet ./...`
- **Build Docker image**: `docker build -t kotatsu-syncserver-go .`
- **Pre-commit checks** (run this before committing): `go build ./... && go vet ./... && go test -short ./...`

## High-level architecture

This is a layered Go HTTP API server using `chi` and MySQL for the Kotatsu manga reader sync:

1. `main.go` bootstraps configuration via `utils.LoadConfig()`, initializes `logger` and `db`, then configures routes.
2. Global middlewares: `middleware.StructuredLogger` and `chi.Recoverer`.
3. Routes are mounted at root level (not /api) to match the Kotlin Kotatsu server.
4. `routes.RegisterRoutes()` composes all domain routes (`/auth`, `/me`, `/manga`, `/resource/*`, `/deeplink/*`).
5. Route handlers are methods on controllers in `controllers/`.
6. Controllers delegate business logic to services in `services/`.
7. Services return typed errors that controllers map to HTTP responses.

### API Routes

```
GET  /                          - Health check ("Alive")
POST /auth                      - Login/Register (combined endpoint)
POST /forgot-password           - Request password reset email
POST /reset-password            - Reset password with token
GET  /deeplink/reset-password   - HTML page with kotatsu:// deep link
GET  /me                        - User profile (auth required)
GET  /manga?offset=&limit=      - Paginated manga list
GET  /manga/{id}                - Single manga
GET  /resource/favourites       - Get favourites (auth required)
POST /resource/favourites       - Sync favourites (auth required)
GET  /resource/history          - Get history (auth required)
POST /resource/history          - Sync history (auth required)
```

### Auth flow

- `services.AuthService.Authenticate()` handles combined login/register and JWT creation via `utils.GenerateJWT`.
- `middleware.RequireAuth` validates bearer tokens via `utils.ParseAndValidateJWT` and injects `user_id` into context.
- JWT tokens expire after 30 days with issuer/audience claims.
- Password hashing uses Argon2id (`utils.HashPassword`, `utils.VerifyPassword`).

### Sync flow

- Favourites and history use timestamp-based merging.
- Client sends `timestamp` with sync data.
- Server compares with `favourites_sync_timestamp` or `history_sync_timestamp` on user record.
- Returns 204 (No Content) if no conflicts, 200 with merged data if conflicts exist.

## Key repository conventions

- Keep request/response formatting centralized through `utils.WriteJSON` and `utils.WriteError`.
- Preserve the controller-service split: controllers translate HTTP concerns, services implement domain logic.
- Services return sentinel errors (e.g., `services.ErrWrongPassword`, `services.ErrInvalidToken`) that controllers map to status codes.
- Protected endpoints use `middleware.RequireAuth` at route level.
- Rate limiting is applied per endpoint group:
  - `AuthLimiter` (5 req/5min) for `/auth`
  - `ForgotPasswordLimiter` (3 req/15min) for `/forgot-password`
  - `ResetPasswordLimiter` (5 req/5min) for `/reset-password`
  - `GlobalAPILimiter` (100 req/5min) for other endpoints
- Use `utils.UserIDFromContext()` to get authenticated user ID in controllers/services.
- Environment configuration via `utils.LoadConfig()` - see `.env.example` for all variables.
- Database access uses package-global `db.DB` initialized by `db.Init()`.

## Testing conventions

- Tests are in `*_test.go` files alongside the code they test.
- Use `testutil` package for common test helpers.
- Integration tests that need a database should check for `TEST_DATABASE_*` env vars.
- Run `go test ./...` before committing to ensure nothing is broken.
- Use table-driven tests for multiple test cases.

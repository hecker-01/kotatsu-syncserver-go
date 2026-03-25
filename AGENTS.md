# AGENTS.md - Coding Guidelines for kotatsu-syncserver-go

## Build, Test, and Verification Commands

### Build

- Build all packages: `go build ./...`
- Build main binary: `go build -o kotatsu-server .`
- Build Docker image: `docker build -t kotatsu-syncserver-go .`

### Run Tests

- Run all tests: `go test ./...`
- Run tests for a single package: `go test ./services`
- Run a single test by name: `go test ./services -run TestName`
- Run tests with verbose output: `go test -v ./...`
- Run tests with race detection: `go test -race ./...`

### Static Analysis

- Run vet: `go vet ./...`

### Clean

- Clean module cache: `go clean -modcache`
- Delete binaries: `rm -f *.exe kotatsu-server` (platform-specific)

## Code Style Guidelines

### Imports

- Place standard library imports first
- Follow with third-party imports
- Follow with local imports
- Use `// import "path"` comments for indirect dependencies in imports
- Group imports logically when using multiple packages from same vendor
- No blank lines between import groups

### Formatting

- Run `go fmt` on all changes before commit
- Use consistent spacing around operators
- Single statement per line
- Maximum line length: 100 characters (trim long lines)
- Indent with spaces (Go default)

### Types

- Prefer `struct` over `new-type` for clarity
- Use exported types for public API, unexported for internal
- Follow Go naming conventions for types
- Add struct tags for JSON mapping: `json:"field_name"`
- Use pointers for structs that might be nil

### Naming Conventions

- Functions/methods: snake_case for Go, proper nouns with lowercase
- Types: PascalCase for exported, lowerCamelCase for unexported
- Variables: lowercase, descriptive
- Constants: ALL_CAPS with underscore separators
- Files: lowercase_with_underscores for non-Go files
- Prefixes: `New` for constructor functions

### Error Handling

- Return typed errors from services (e.g., `ErrInvalidInput`)
- Do not panic (except unrecoverable runtime errors)
- Use `errors.Is()` for error checking
- Provide meaningful error messages for API context
- Convert service errors to HTTP responses at controller layer
- Use standard HTTP status codes: 400 for bad request, 401 for auth failed, 404 for not found, 500 for server error

### Comments

- Document non-obvious logic
- Add function signatures with parameter descriptions
- Document exported types and functions
- Keep inline comments brief and current
- Do not comment on obvious code

## Architecture Patterns

### Layered Architecture

```txt
Routes       # Route registration
  ↓
Middleware   # Cross-cutting concerns (auth, logging, rate limiting)
  ↓
Controllers   # HTTP request/response translation
  ↓
Services      # Business logic
  ↓
DB Utils      # Database access, utilities
```

### Authentication Flow

1. User registers/logs in via public auth endpoints
2. `services.AuthService` validates and creates JWT via `utils.GenerateJWT`
3. Client includes JWT in `Authorization: Bearer <token>` header
4. `middleware.RequireAuth` validates token and injects `user_id` into context
5. Services/controllers retrieve user identity via `utils.UserIDFromContext`

### Rate Limiting Layering

- Global API rate limiter: 100 requests/5 minutes (in `main.go`)
- Auth endpoints: 5 requests/5 minutes (separate configuration)
- All API routes use middleware layers

### Controller-Service Split

- Controllers handle HTTP concerns (request parsing, response formatting)
- Services implement domain logic and return typed errors
- Use `utils.WriteJSON` and `utils.WriteError` for consistent responses
- Prefer controllers over deprecated handlers/

## Utilities and Helpers

### Error Types

Use sentinel errors from services:

- `services.ErrInvalidInput`
- `services.ErrInvalidCredentials`
- `services.ErrEmailExists`

From utils:

- `utils.ErrInvalidToken`

### Context Access

- `utils.UserIDFromContext` - Extract user_id from request context
- Always use `context.Context` for cancellable operations

### JSON Responses

- `utils.WriteJSON(w, statusCode, data)` - Success responses
- `utils.WriteError(w, statusCode, error)` - Error responses

## Database Access

- Use package-global `db.DB` initialized by `db.Init()`
- QueryRow for single row expectations
- Exec for INSERT/UPDATE/DELETE statements
- Scan results into prepared variables
- Wrap database operations in error checks

## Environment Variables

Required at startup:

- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASS`
- `JWT_SECRET`

Optional:

- `PORT` (defaults to 9292)
- `LOG_LEVEL`, `LOG_FORMAT`, `LOG_DIRECTORY`, etc.

## Testing Guidelines

- Test business logic in services layer
- Mock external dependencies when appropriate
- Test error paths, not just happy paths
- Use table-driven tests where possible
- Test authentication flows end-to-end
- Verify rate limiting behavior

## Security Considerations

- Never log sensitive data (passwords, tokens, credentials)
- Validate all user inputs
- Use parameterized queries to prevent SQL injection
- Hash passwords with bcrypt
- Use strong JWT secrets (long, random strings)
- Implement rate limiting to prevent brute force attacks

## Docker Deployment

Build image: `docker build -t kotatsu-syncserver-go .`

Run with environment variables for DB connection and JWT secret.

## Additional Configuration

Check `.github/copilot-instructions.md` for Copilot-specific guidance.

This file guides agentic coding operations. Follow these conventions when modifying, testing, or extending this codebase.

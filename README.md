# Kotatsu Sync Server

A high-performance HTTP API server written in Go for managing user authentication and synchronization data. Built with `chi` router and MySQL, featuring JWT authentication, structured logging, and rate limiting.

## Features

- **JWT Authentication** - Secure user registration and login with bcrypt password hashing
- **Layered Architecture** - Clean separation of concerns with controllers, services, and handlers
- **Structured Logging** - Comprehensive logging with configurable levels, formats (JSON/text), and file rotation
- **Rate Limiting** - IP-based rate limiting with separate limits for auth endpoints
- **MySQL Database** - Connection pooling and efficient database access
- **Docker Support** - Multi-stage Docker builds for production deployment

## Prerequisites

- **Go 1.25+**
- **MySQL 5.7+** or **MariaDB 10.3+**
- **Docker** (optional, for containerized deployment)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/hecker-01/kotatsu-syncserver-go.git
cd kotatsu-syncserver-go
```

2. Install dependencies:
```bash
go mod download
```

3. Set up the database:
```bash
mysql -u root -p < setup.sql
```

4. Configure environment variables:
```bash
cp .env.example .env
# Edit .env with your database credentials and JWT secret
```

## Configuration

Create a `.env` file in the project root with the following variables:

### Required Variables

```env
# Database connection
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=kotatsugo
DB_USER=kotatsugo
DB_PASS=your_secure_password

# JWT authentication (use a long random string)
JWT_SECRET=your_jwt_secret_key_here
```

### Optional Variables

```env
# Server configuration
PORT=9292

# Logging configuration
LOG_LEVEL=info              # Options: trace, debug, info, warn, error, fatal
LOG_FORMAT=text             # Options: text, json
LOG_DIRECTORY=logs
LOG_MAX_FILE_SIZE=20m       # Max size per log file
LOG_MAX_FILES=14            # Number of backup files to keep

# Enable/disable logging outputs
ENABLE_FILE_LOGGING=true
ENABLE_CONSOLE_LOGGING=true
ENABLE_ACCESS_FILE_LOGGING=true
```

## Running the Server

### Development

```bash
go run main.go
```

The server will start on `http://localhost:9292` (or the port specified in your `.env`).

### Production

```bash
go build -o kotatsu-server .
./kotatsu-server
```

### Docker

```bash
# Build the image
docker build -t kotatsu-syncserver .

# Run the container
docker run -p 9292:9292 \
  -e DB_HOST=your_db_host \
  -e DB_NAME=kotatsugo \
  -e DB_USER=kotatsugo \
  -e DB_PASS=your_password \
  -e JWT_SECRET=your_secret \
  kotatsu-syncserver
```

## API Endpoints

### Health Check

```http
GET /
GET /health
```

Returns server status.

### Authentication

All authentication endpoints are rate-limited to 5 requests per 5 minutes per IP.

#### Register

```http
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**Response:** `201 Created`
```json
{
  "message": "User created"
}
```

#### Login

```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**Response:** `200 OK`
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Protected Endpoints

All endpoints below require authentication via Bearer token:

```http
Authorization: Bearer <your_jwt_token>
```

#### Get Current User

```http
GET /api/users/me
```

**Response:** `200 OK`
```json
{
  "user_id": 1
}
```

#### List Games

```http
GET /api/games
```

**Response:** `200 OK`
```json
{
  "data": []
}
```

#### List Player Data

```http
GET /api/player
```

**Response:** `200 OK`
```json
{
  "data": []
}
```

#### List History

```http
GET /api/history
```

**Response:** `200 OK`
```json
{
  "data": []
}
```

## Development

### Project Structure

```
.
├── main.go                 # Application entry point
├── controllers/            # HTTP handlers (request/response translation)
├── services/              # Business logic layer
├── routes/                # Route registration
├── middleware/            # Auth, logging, rate limiting
├── db/                    # Database initialization
├── logger/                # Structured logging implementation
├── utils/                 # JWT, context, and response utilities
├── handlers/              # Legacy handlers (deprecated)
├── setup.sql              # Database schema
└── .env.example           # Environment variable template
```

### Building

```bash
# Build all packages
go build ./...

# Build the main binary
go build -o kotatsu-server .
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./services

# Run a single test by name
go test ./services -run TestName
```

### Linting

```bash
go vet ./...
```

## Architecture

This server follows a layered architecture pattern:

1. **Routes** (`routes/`) - Define HTTP routes and mount controllers
2. **Middleware** (`middleware/`) - Handle cross-cutting concerns (auth, logging, rate limiting)
3. **Controllers** (`controllers/`) - Parse HTTP requests, call services, format responses
4. **Services** (`services/`) - Implement business logic, return typed errors
5. **Database** (`db/`) - Manage database connection pool
6. **Utils** (`utils/`) - Provide shared utilities (JWT, response formatting, context)

### Authentication Flow

1. User registers or logs in via `/api/auth/register` or `/api/auth/login`
2. `services.AuthService` validates credentials and generates a JWT via `utils.GenerateJWT`
3. Client includes JWT in `Authorization: Bearer <token>` header for protected routes
4. `middleware.RequireAuth` validates the token and injects `user_id` into request context
5. Controllers/services retrieve user identity via `utils.UserIDFromContext`

## Rate Limiting

- **Global API limit:** 100 requests per 5 minutes per IP
- **Auth endpoints limit:** 5 requests per 5 minutes per IP (more restrictive)

Rate limits return `429 Too Many Requests` with a `Retry-After` header when exceeded.

## Logging

The server uses structured logging with the following features:

- **Multiple outputs:** Console (with colors) and/or file logging
- **Log rotation:** Automatic rotation based on size and retention policy
- **Access logs:** Separate HTTP access logs with request details
- **Contextual logging:** Structured key-value pairs for easy filtering

Log files are written to the `logs/` directory:
- `combined.log` - All log levels
- `error.log` - Error and fatal logs only
- `access.log` - HTTP access logs

## License

[Specify your license here]

## Contributing

[Add contribution guidelines here]

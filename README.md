# Kotatsu Sync Server (Go)

A Go implementation of the Kotatsu manga reader sync server. This is a port of the original [Kotatsu Sync Server](https://github.com/KotatsuApp/kotatsu-sync-server) (Kotlin/Ktor) to Go, providing synchronization capabilities for the [Kotatsu](https://github.com/KotatsuApp/Kotatsu) manga reader application.

## Features

- **User Authentication** - JWT-based authentication with 30-day token expiration
- **Password Reset** - Email-based password reset with configurable SMTP or console provider
- **Manga Library Sync** - Synchronize favourites with category support across devices
- **Reading History Sync** - Track and sync reading progress across devices
- **Rate Limiting** - IP-based rate limiting with stricter limits on auth endpoints
- **Argon2 Password Hashing** - Secure password storage using Argon2id
- **Structured Logging** - Comprehensive logging with configurable levels, formats, and file rotation
- **Docker Support** - Multi-stage Docker builds for production deployment

## Requirements

- **Go 1.21+**
- **MySQL 5.7+** or **MariaDB 10.3+**
- **Docker** (optional, for containerized deployment)

## Quick Start

```bash
# Clone the repository
git clone https://github.com/KotatsuApp/kotatsu-syncserver-go.git
cd kotatsu-syncserver-go

# Configure environment
cp .env.example .env
# Edit .env with your database credentials and JWT secret

# Setup database (choose one method):

# Method 1: Manual setup (traditional)
mysql -u root -p < setup.sql

# Method 2: Automatic setup (Docker Compose friendly)
# Set DATABASE_ROOT_PASSWORD in .env, then just run the server
# The database will be created automatically on first run

# Run the server
go run main.go
```

The server will start on `http://localhost:9292` (or the port specified in your `.env`).

## Environment Variables

Create a `.env` file in the project root:
- For **local development**: Copy `.env.example` 
- For **Docker Compose**: Copy `.env.docker.example` (has better defaults for containers)

Both files have the same variables, just different default values and comments.

### Required Variables

| Variable            | Description                                               |
| ------------------- | --------------------------------------------------------- |
| `DATABASE_HOST`     | Database server hostname (default: `localhost`)           |
| `DATABASE_PORT`     | Database server port (default: `3306`)                    |
| `DATABASE_NAME`     | Database name (default: `kotatsu_db`)                     |
| `DATABASE_USER`     | Database username                                         |
| `DATABASE_PASSWORD` | Database password                                         |
| `JWT_SECRET`        | HMAC256 secret for JWT signing (use a long random string) |

### Optional Variables

| Variable                 | Default                 | Description                                                                                                            |
| ------------------------ | ----------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `DATABASE_ROOT_PASSWORD` | -                       | MySQL root password. If set, server auto-creates database, user (from DATABASE_USER/PASSWORD), and tables on first run |
| `PORT`                   | `9292`                  | HTTP port the server listens on                                                                                        |
| `ALLOW_NEW_REGISTER`     | `true`                  | Enable/disable new user registration                                                                                   |
| `BASE_URL`               | `http://localhost:9292` | Base URL for email links                                                                                               |
| `MAIL_PROVIDER`          | `console`               | Mail provider: `console` (logs) or `smtp`                                                                              |
| `SMTP_HOST`              | -                       | SMTP server hostname                                                                                                   |
| `SMTP_PORT`              | `587`                   | SMTP server port                                                                                                       |
| `SMTP_USERNAME`          | -                       | SMTP authentication username                                                                                           |
| `SMTP_PASSWORD`          | -                       | SMTP authentication password                                                                                           |
| `SMTP_FROM`              | -                       | From email address for outgoing mail                                                                                   |

### Logging Variables

| Variable                     | Default | Description                                                   |
| ---------------------------- | ------- | ------------------------------------------------------------- |
| `LOG_LEVEL`                  | `info`  | Log level: `trace`, `debug`, `info`, `warn`, `error`, `fatal` |
| `LOG_FORMAT`                 | `text`  | Log format: `text` (with colors) or `json`                    |
| `LOG_DIRECTORY`              | `logs`  | Directory for log files                                       |
| `LOG_MAX_FILE_SIZE`          | `20m`   | Maximum size per log file                                     |
| `LOG_MAX_FILES`              | `14`    | Number of backup log files to keep                            |
| `ENABLE_FILE_LOGGING`        | `true`  | Enable file logging                                           |
| `ENABLE_CONSOLE_LOGGING`     | `true`  | Enable console logging                                        |
| `ENABLE_ACCESS_FILE_LOGGING` | `true`  | Enable HTTP access file logging                               |

## Running the Server

### Development

```bash
go run main.go
```

### Production Build

```bash
go build -o kotatsu-server .
./kotatsu-server
```

### Docker

```bash
# Build the image
docker build -t kotatsu-syncserver-go .

# Run the container
docker run -p 9292:9292 --env-file .env kotatsu-syncserver-go

# Or with individual environment variables
docker run -p 9292:9292 \
  -e DATABASE_HOST=your_db_host \
  -e DATABASE_NAME=kotatsu_db \
  -e DATABASE_USER=kotatsu \
  -e DATABASE_PASSWORD=your_password \
  -e JWT_SECRET=your_secret \
  kotatsu-syncserver-go
```

### Docker Compose (Recommended)

The easiest way to run the server with MySQL:

```bash
# 1. Copy the Docker Compose environment file
cp .env.docker.example .env

# 2. Edit .env and set secure passwords and JWT secret
#    IMPORTANT: Change these values in production!
#    - DATABASE_PASSWORD
#    - DATABASE_ROOT_PASSWORD
#    - JWT_SECRET (generate with: openssl rand -base64 32)

# 3. Start both MySQL and the server
docker-compose up -d

# The database will be automatically created on first run
# No manual setup.sql execution needed!

# View logs
docker-compose logs -f kotatsu-server

# Stop services
docker-compose down

# Stop and remove data
docker-compose down -v
```

**Security Note**: The `docker-compose.yml` reads sensitive information from your `.env` file. Never commit your `.env` file to version control - it's already in `.gitignore`.

## API Endpoints

### Health Check

| Method | Endpoint | Description     |
| ------ | -------- | --------------- |
| `GET`  | `/`      | Returns "Alive" |

### Authentication

| Method | Endpoint           | Description               | Rate Limit |
| ------ | ------------------ | ------------------------- | ---------- |
| `POST` | `/auth`            | Login or register         | 5 req/5min |
| `POST` | `/forgot-password` | Request password reset    | 3 req/5min |
| `POST` | `/reset-password`  | Reset password with token | 5 req/5min |

### Deeplink

| Method | Endpoint                             | Description                      |
| ------ | ------------------------------------ | -------------------------------- |
| `GET`  | `/deeplink/reset-password?token=...` | HTML page with Kotatsu deep link |

### User (Protected)

All endpoints below require `Authorization: Bearer <token>` header.

| Method | Endpoint | Description              |
| ------ | -------- | ------------------------ |
| `GET`  | `/me`    | Get current user profile |

### Manga (Public)

| Method | Endpoint                   | Description                |
| ------ | -------------------------- | -------------------------- |
| `GET`  | `/manga?offset=0&limit=20` | List manga with pagination |
| `GET`  | `/manga/{id}`              | Get single manga by ID     |

### Sync Resources (Protected)

| Method | Endpoint               | Description                          |
| ------ | ---------------------- | ------------------------------------ |
| `GET`  | `/resource/favourites` | Get favourites and categories        |
| `POST` | `/resource/favourites` | Sync favourites (returns 204 or 200) |
| `GET`  | `/resource/history`    | Get reading history                  |
| `POST` | `/resource/history`    | Sync history (returns 204 or 200)    |

For detailed API documentation, see [API_DOCUMENTATION.md](API_DOCUMENTATION.md).

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
├── models/                # Data models and DTOs
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

1. User registers or logs in via `POST /auth`
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

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

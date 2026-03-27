# Kotatsu Sync Server API Documentation

## Overview

The Kotatsu Sync Server is a Kotlin/Ktor-based REST API that provides synchronization capabilities for the Kotatsu manga reader application. It manages user accounts, manga collections, reading history, favorites, and categories across multiple devices.

**Current Stack:**

- Language: Kotlin
- Framework: Ktor (Server)
- Database: MariaDB/MySQL
- ORM: Ktorm
- Authentication: JWT (Auth0)
- Migration Tool: Flyway
- Password Hashing: Argon2

---

## Database Schema

### 1. Users Table (`users`)

Stores user account information and synchronization timestamps.

```sql
CREATE TABLE users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(320) NOT NULL,
    password_hash VARCHAR(128) NOT NULL,
    nickname VARCHAR(100),
    favourites_sync_timestamp BIGINT,
    history_sync_timestamp BIGINT,
    password_reset_token_hash CHAR(64),
    password_reset_token_expires_at BIGINT,
    CONSTRAINT uq_users_email UNIQUE (email),
    UNIQUE INDEX uq_users_password_reset_token_hash (password_reset_token_hash)
);
```

**Key Columns:**

- `id`: Auto-incremented user ID
- `email`: Unique email address (max 320 chars)
- `password_hash`: Argon2 hashed password (128 chars)
- `nickname`: Optional user display name
- `favourites_sync_timestamp`: Last sync timestamp for favorites
- `history_sync_timestamp`: Last sync timestamp for reading history
- `password_reset_token_hash`: SHA256 hash of reset token
- `password_reset_token_expires_at`: Unix timestamp (seconds) for token expiration

---

### 2. Manga Table (`manga`)

Stores manga metadata from various sources.

```sql
CREATE TABLE manga (
    id BIGINT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    alt_title VARCHAR(255),
    url VARCHAR(255) NOT NULL,
    public_url VARCHAR(255) NOT NULL,
    rating FLOAT NOT NULL,
    content_rating ENUM('SAFE', 'SUGGESTIVE', 'ADULT'),
    cover_url VARCHAR(255) NOT NULL,
    large_cover_url VARCHAR(255),
    state ENUM('ONGOING', 'FINISHED', 'ABANDONED', 'PAUSED', 'UPCOMING', 'RESTRICTED'),
    author VARCHAR(64),
    source VARCHAR(32) NOT NULL
);
```

**Key Columns:**

- `id`: Manga ID from source (primary key, not auto-increment)
- `source`: Manga source identifier (e.g., "mangadex", "mangahub")
- `state`: Current publication state enum
- `content_rating`: Content rating category
- `public_url`: URL for accessing the manga

---

### 3. Tags Table (`tags`)

Stores genre/category tags associated with manga.

```sql
CREATE TABLE tags (
    id BIGINT PRIMARY KEY,
    title VARCHAR(64) NOT NULL,
    `key` VARCHAR(120) NOT NULL,
    source VARCHAR(32) NOT NULL
);
```

---

### 4. Manga-Tags Junction Table (`manga_tags`)

Many-to-many relationship between manga and tags.

```sql
CREATE TABLE manga_tags (
    manga_id BIGINT NOT NULL,
    tag_id BIGINT NOT NULL,
    PRIMARY KEY (manga_id, tag_id),
    INDEX idx_manga_tags_tag_id (tag_id),
    CONSTRAINT fk_manga_tags_tag_id FOREIGN KEY (tag_id) REFERENCES tags(id),
    CONSTRAINT fk_manga_tags_manga_id FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE
);
```

---

### 5. Categories Table (`categories`)

User-created collections for organizing favorites (e.g., "Reading", "Completed", "On Hold").

```sql
CREATE TABLE categories (
    id BIGINT NOT NULL,
    created_at BIGINT NOT NULL,
    sort_key INT NOT NULL,
    title VARCHAR(120) NOT NULL,
    `order` VARCHAR(16) NOT NULL,
    user_id BIGINT NOT NULL,
    track TINYINT(1) NOT NULL,
    show_in_lib TINYINT(1) NOT NULL,
    deleted_at BIGINT,
    PRIMARY KEY (id, user_id),
    CONSTRAINT fk_categories_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

**Key Columns:**

- `id`: Category ID (composite primary key with user_id)
- `user_id`: Owner of the category
- `created_at`: Unix timestamp (milliseconds) for creation
- `track`: Whether to show new chapters notifications
- `show_in_lib`: Whether to display in library
- `deleted_at`: Soft delete timestamp (NULL = active)
- `order`: Sort order strategy

---

### 6. Favourites Table (`favourites`)

Junction table linking manga to user categories with metadata.

```sql
CREATE TABLE favourites (
    manga_id BIGINT NOT NULL,
    category_id BIGINT NOT NULL,
    sort_key INT NOT NULL,
    pinned TINYINT(1) NOT NULL,
    created_at BIGINT NOT NULL,
    deleted_at BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    PRIMARY KEY (manga_id, category_id, user_id),
    INDEX idx_favourites_user_id (user_id),
    CONSTRAINT fk_favourites_manga_id FOREIGN KEY (manga_id) REFERENCES manga(id),
    CONSTRAINT fk_favourites_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_favourites_category FOREIGN KEY (category_id, user_id) REFERENCES categories(id, user_id) ON DELETE CASCADE
);
```

**Key Columns:**

- Composite primary key: `manga_id, category_id, user_id`
- `sort_key`: Position in the category
- `pinned`: Whether the manga is pinned
- `deleted_at`: Soft delete marker (0 = deleted, non-zero = timestamp)

---

### 7. History Table (`history`)

User's manga reading progress and history.

```sql
CREATE TABLE history (
    manga_id BIGINT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    chapter_id BIGINT NOT NULL,
    page SMALLINT NOT NULL,
    scroll DOUBLE NOT NULL,
    percent DOUBLE NOT NULL,
    chapters INT NOT NULL,
    deleted_at BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    PRIMARY KEY (user_id, manga_id),
    INDEX idx_manga_id (manga_id),
    CONSTRAINT fk_history_manga_id FOREIGN KEY (manga_id) REFERENCES manga(id),
    CONSTRAINT fk_history_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

**Key Columns:**

- `chapter_id`: Last read chapter ID
- `page`: Last read page number
- `scroll`: Scroll position (0-1 or pixels)
- `percent`: Reading progress percentage
- `chapters`: Total chapters count
- `deleted_at`: Soft delete marker

---

## Authentication

### JWT Token System

**Token Structure:**

- **Algorithm**: HMAC256
- **Secret**: Configured via `JWT_SECRET` environment variable
- **Issuer**: Configured via `jwt.issuer` (default: `http://0.0.0.0:9292/`)
- **Audience**: Configured via `jwt.audience` (default: `http://0.0.0.0:9292/resource`)
- **Lifetime**: 30 days
- **Claims**: `user_id` (int) - User's database ID

**Token Generation Flow:**

1. User sends email + password to `/auth` endpoint
2. Server validates credentials with Argon2
3. Server generates JWT token with 30-day expiration
4. Client stores token and includes in `Authorization: Bearer <token>` header

**Token Verification:**

- Applied to all `/resource/*` endpoints
- JWT middleware validates signature, issuer, audience, and expiration
- Extracts `user_id` claim and loads current user from database

---

## API Endpoints

### Base URL

```
http://localhost:9292
```

Default port is 9292, configurable via `PORT` environment variable.

---

### Health & Status

#### GET `/`

**Description:** Health check endpoint
**Authentication:** None
**Response:**

```
200 OK: "Alive"
```

---

### Authentication Endpoints

#### POST `/auth`

**Description:** Login or register a user

**Authentication:** None  
**Rate Limit:** `auth` (more restrictive)

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response (Success - 200):**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (Failure - 400):**

```
Wrong password
```

**Behavior:**

- If user exists: authenticate with password
- If user doesn't exist AND `ALLOW_NEW_REGISTER=true`: create new account
- If user doesn't exist AND `ALLOW_NEW_REGISTER=false`: reject login
- Passwords hashed with Argon2
- Password length: 2-24 characters

---

#### POST `/forgot-password`

**Description:** Request password reset email

**Authentication:** None  
**Rate Limit:** `forgotPassword` (more restrictive)

**Request Body:**

```json
{
  "email": "user@example.com"
}
```

**Response (Always - 200):**

```
A password reset email was sent
```

**Behavior:**

- Always returns 200 (no email enumeration)
- Only sends email if:
  - User exists
  - No active reset token (checked against `password_reset_token_expires_at`)
- Generates cryptographic reset token
- Sends email with deep link: `{BASE_URL}/deeplink/reset-password?token={TOKEN}`
- Token expires after configured time

---

#### POST `/reset-password`

**Description:** Reset user password with valid token

**Authentication:** None  
**Rate Limit:** `resetPassword` (more restrictive)

**Request Body:**

```json
{
  "resetToken": "token_from_email_link",
  "password": "newpassword123"
}
```

**Response (Success - 200):**

```
Password reset successfully
```

**Response (Failure - 400):**

```
Invalid or expired token
```

or

```
Password should be from 2 to 24 characters long
```

**Behavior:**

- Validates token against `password_reset_token_hash`
- Checks expiration timestamp
- Enforces password length: 2-24 chars
- Hashes new password with Argon2

---

### User Profile

#### GET `/me`

**Description:** Get current authenticated user info

**Authentication:** Required (JWT)  
**Rate Limit:** `globalApi`

**Response (Success - 200):**

```json
{
  "id": 1,
  "email": "user@example.com",
  "nickname": "MyNickname"
}
```

**Response (Unauthorized - 401):**

```
Unauthorized
```

---

### Manga Endpoints

#### GET `/manga?offset={offset}&limit={limit}`

**Description:** Get paginated list of manga

**Authentication:** None  
**Rate Limit:** `globalApi`

**Query Parameters:**

- `offset` (required, integer): Number of records to skip
- `limit` (required, integer): Number of records to return

**Response (Success - 200):**

```json
[
  {
    "id": 12345678,
    "title": "Manga Title",
    "altTitle": "Alternative Title",
    "url": "https://source.com/manga/manga-slug",
    "publicUrl": "https://source.com/manga/manga-slug",
    "rating": 8.5,
    "contentRating": "SAFE",
    "coverUrl": "https://cdn.example.com/cover.jpg",
    "largeCoverUrl": "https://cdn.example.com/cover-large.jpg",
    "state": "ONGOING",
    "author": "Author Name",
    "source": "mangadex"
  },
  ...
]
```

**Response (Invalid parameters - 400):**

```
Parameter "offset" is missing or invalid
```

or

```
Parameter "limit" is missing or invalid
```

---

#### GET `/manga/{id}`

**Description:** Get specific manga by ID

**Authentication:** None  
**Rate Limit:** `globalApi`

**Path Parameters:**

- `id` (required, long): Manga ID

**Response (Success - 200):**

```json
{
  "id": 12345678,
  "title": "Manga Title",
  "altTitle": "Alternative Title",
  "url": "https://source.com/manga/manga-slug",
  "publicUrl": "https://source.com/manga/manga-slug",
  "rating": 8.5,
  "contentRating": "SAFE",
  "coverUrl": "https://cdn.example.com/cover.jpg",
  "largeCoverUrl": "https://cdn.example.com/cover-large.jpg",
  "state": "ONGOING",
  "author": "Author Name",
  "source": "mangadex"
}
```

**Response (Not found - 404):**

```
Not Found
```

---

### Favorites Sync

#### GET `/resource/favourites`

**Description:** Get user's current favorites and categories

**Authentication:** Required (JWT)  
**Rate Limit:** `globalApi`

**Response (Success - 200):**

```json
{
  "categories": [
    {
      "id": 1,
      "createdAt": 1672531200000,
      "sortKey": 0,
      "title": "Reading",
      "order": "LATEST",
      "track": true,
      "showInLib": true,
      "deletedAt": null
    },
    {
      "id": 2,
      "createdAt": 1672531200000,
      "sortKey": 1,
      "title": "Completed",
      "order": "LATEST",
      "track": false,
      "showInLib": true,
      "deletedAt": null
    }
  ],
  "favourites": [
    {
      "mangaId": 12345678,
      "categoryId": 1,
      "sortKey": 0,
      "pinned": true,
      "createdAt": 1672531200000,
      "deletedAt": null
    }
  ],
  "timestamp": 1672531200000
}
```

**Response (Unauthorized - 401):**

```
Unauthorized
```

---

#### POST `/resource/favourites`

**Description:** Sync favorites from client

**Authentication:** Required (JWT)  
**Rate Limit:** `globalApi`

**Request Body:**

```json
{
  "categories": [
    {
      "id": 1,
      "createdAt": 1672531200000,
      "sortKey": 0,
      "title": "Reading",
      "order": "LATEST",
      "track": true,
      "showInLib": true,
      "deletedAt": null
    }
  ],
  "favourites": [
    {
      "mangaId": 12345678,
      "categoryId": 1,
      "sortKey": 0,
      "pinned": true,
      "createdAt": 1672531200000,
      "deletedAt": 1672531200000
    }
  ],
  "timestamp": 1672531200000
}
```

**Response (No changes - 204):**

```
(No Content)
```

**Response (With changes - 200):**

```json
{
  "categories": [...],
  "favourites": [...],
  "timestamp": 1672531200000
}
```

**Response (Unauthorized - 401):**

```
Unauthorized
```

**Behavior:**

- Server merges client data with server state
- Returns 204 if no conflicts/changes
- Returns 200 with merged state if conflicts detected
- Updates `favourites_sync_timestamp` after sync
- Uses timestamps to determine which version is newer

---

### History Sync

#### GET `/resource/history`

**Description:** Get user's current reading history

**Authentication:** Required (JWT)  
**Rate Limit:** `globalApi`

**Response (Success - 200):**

```json
{
  "history": [
    {
      "mangaId": 12345678,
      "createdAt": 1672531200000,
      "updatedAt": 1672617600000,
      "chapterId": 987654,
      "page": 15,
      "scroll": 0.45,
      "percent": 30.5,
      "chapters": 100,
      "deletedAt": null
    }
  ],
  "timestamp": 1672617600000
}
```

**Response (Unauthorized - 401):**

```
Unauthorized
```

---

#### POST `/resource/history`

**Description:** Sync history from client

**Authentication:** Required (JWT)  
**Rate Limit:** `globalApi`

**Request Body:**

```json
{
  "history": [
    {
      "mangaId": 12345678,
      "createdAt": 1672531200000,
      "updatedAt": 1672617600000,
      "chapterId": 987654,
      "page": 15,
      "scroll": 0.45,
      "percent": 30.5,
      "chapters": 100,
      "deletedAt": 0
    }
  ],
  "timestamp": 1672617600000
}
```

**Response (No changes - 204):**

```
(No Content)
```

**Response (With changes - 200):**

```json
{
  "history": [...],
  "timestamp": 1672617600000
}
```

**Response (Unauthorized - 401):**

```
Unauthorized
```

**Behavior:**

- Server merges client history with server state
- Returns 204 if no conflicts/changes
- Returns 200 with merged state if conflicts detected
- Updates `history_sync_timestamp` after sync
- Uses `updatedAt` timestamps to determine merges

---

### Deep Links

#### GET `/deeplink/reset-password?token={token}`

**Description:** HTML page with deep link for password reset in Kotatsu app

**Authentication:** None

**Query Parameters:**

- `token` (required, string): Reset token from email

**Response (Success - 200):**

```html
(HTML page with kotatsu:// deep link) Deep link format:
kotatsu://reset-password?base_url={BASE_URL}&token={TOKEN}
```

**Response (Missing token - 400):**

```
Missing token
```

---

## Data Models

### User Model

```kotlin
interface UserEntity : Entity<UserEntity> {
    var id: Int
    var email: String
    var passwordHash: String
    var passwordResetTokenHash: String?
    var passwordResetTokenExpiresAt: Long?
    var nickname: String?
    var favouritesSyncTimestamp: Long?
    var historySyncTimestamp: Long?
}
```

### Authentication Requests

```kotlin
@Serializable
data class AuthRequest(
    @SerialName("email") val email: String,
    @SerialName("password") val password: String,
)

@Serializable
data class ForgotPasswordRequest(
    @SerialName("email") val email: String,
)

@Serializable
data class ResetPasswordRequest(
    @SerialName("resetToken") val resetToken: String,
    @SerialName("password") val password: String,
)
```

### Manga Models

```kotlin
data class Manga(
    val id: Long,
    val title: String,
    val altTitle: String?,
    val url: String,
    val publicUrl: String,
    val rating: Float,
    val contentRating: ContentRating?,
    val coverUrl: String,
    val largeCoverUrl: String?,
    val state: MangaState,
    val author: String?,
    val source: String,
)

enum class ContentRating {
    SAFE, SUGGESTIVE, ADULT
}

enum class MangaState {
    ONGOING, FINISHED, ABANDONED, PAUSED, UPCOMING, RESTRICTED
}
```

### Sync Packages

```kotlin
@Serializable
data class FavouritesPackage(
    @SerialName("categories") val favouriteCategories: List<Category>,
    @SerialName("favourites") val favourites: List<Favourite>,
    @SerialName("timestamp") val timestamp: Long?,
)

@Serializable
data class HistoryPackage(
    @SerialName("history") val history: List<History>,
    @SerialName("timestamp") val timestamp: Long?,
)

data class Category(
    val id: Long,
    val createdAt: Long,
    val sortKey: Int,
    val title: String,
    val order: String,
    val track: Boolean,
    val showInLib: Boolean,
    val deletedAt: Long?,
)

data class Favourite(
    val mangaId: Long,
    val categoryId: Long,
    val sortKey: Int,
    val pinned: Boolean,
    val createdAt: Long,
    val deletedAt: Long?,
)

data class History(
    val mangaId: Long,
    val createdAt: Long,
    val updatedAt: Long,
    val chapterId: Long,
    val page: Int,
    val scroll: Double,
    val percent: Double,
    val chapters: Int,
    val deletedAt: Long?,
)
```

---

## Rate Limiting

The API implements rate limiting on different endpoint groups for abuse protection:

**Rate Limit Groups:**

1. **`auth`** - More restrictive for auth endpoints (`/auth`, `/forgot-password`, `/reset-password`)
2. **`forgotPassword`** - Most restrictive for password reset requests
3. **`resetPassword`** - Most restrictive for password reset submission
4. **`globalApi`** - Standard rate limit for general API endpoints

**Application:**

- Rate limits are applied via Ktor's built-in rate limiting plugin
- Configure limits via `application.conf` or environment variables
- Returns 429 Too Many Requests when limit exceeded

---

## Configuration

### Environment Variables

**Database Configuration:**

- `DATABASE_HOST` - Database server hostname (default: `localhost`)
- `DATABASE_PORT` - Database server port (default: `3306`)
- `DATABASE_NAME` - Database name (default: `kotatsu_db`)
- `DATABASE_USER` - Database username (required)
- `DATABASE_PASSWORD` - Database password (required)
- `DATABASE_DIALECT` - `mariadb` or `mysql` (default: `mariadb`)

**JWT Configuration:**

- `JWT_SECRET` - HMAC256 secret for JWT signing (required)

**Server Configuration:**

- `PORT` - HTTP port (default: `9292`)

**Application Configuration:**

- `ALLOW_NEW_REGISTER` - Enable/disable new user registration (`true`/`false`, default: `true`)
- `BASE_URL` - Base URL for deep links and emails (default: `http://localhost:9292`)

**Mail Configuration:**

- `MAIL_PROVIDER` - `console` or `smtp` (default: `console`)
- `SMTP_HOST` - SMTP server hostname
- `SMTP_PORT` - SMTP server port
- `SMTP_USERNAME` - SMTP authentication username
- `SMTP_PASSWORD` - SMTP authentication password
- `SMTP_FROM` - From email address for outgoing mail

### Configuration File

Located at `src/main/resources/application.conf`:

```hocon
ktor {
    deployment {
        port = 9292
        port = ${?PORT}
    }
    application {
        modules = [ org.kotatsu.ApplicationKt.module ]
    }
}
jwt {
    secret = ${?JWT_SECRET}
    issuer = "http://0.0.0.0:9292/"
    audience = "http://0.0.0.0:9292/resource"
}
database {
    name = "kotatsu_db"
    name = ${?DATABASE_NAME}
    dialect = "mariadb"
    dialect = ${?DATABASE_DIALECT}
    host = "localhost"
    host = ${?DATABASE_HOST}
    port = 3306
    port = ${?DATABASE_PORT}
    user = ${?DATABASE_USER}
    password = ${?DATABASE_PASSWORD}
}
kotatsu {
    allow_new_register = true
    allow_new_register = ${?ALLOW_NEW_REGISTER}
    mail_provider = "console"
    mail_provider = ${?MAIL_PROVIDER}
    base_url = "http://localhost:9292"
    base_url = ${?BASE_URL}
}
smtp {
    host = ${?SMTP_HOST}
    port = ${?SMTP_PORT}
    username = ${?SMTP_USERNAME}
    password = ${?SMTP_PASSWORD}
    from = ${?SMTP_FROM}
}
```

---

## Security & Password Handling

### Password Hashing

- **Algorithm**: Argon2 (recommended by OWASP)
- **Implementation**: `org.argon2.jvm.Argon2Factory`
- All passwords stored as Argon2 hashes, never plaintext

### Password Validation

- Email format validation
- Password length: 2-24 characters
- Passwords validated during registration and password reset

### Password Reset Tokens

- **Token Format**: Cryptographic random token (base64)
- **Storage**: SHA256 hash of token (token never stored plaintext)
- **Expiration**: Configurable length (default: check `password_reset_token_expires_at`)
- **Security**: Tokens tied to email verification for account recovery

### JWT Security

- **Signing Algorithm**: HMAC256 with secret key
- **Expiration**: 30 days
- **Claims**: Minimal (only `user_id`)
- **Verification**: Signature, issuer, audience, and expiration checked on every authenticated request

---

## Database Access Patterns

### ORM: Ktorm

The API uses **Ktorm** for type-safe database access:

**Key Features:**

- Type-safe SQL builders
- Entity mapping to Kotlin objects
- Transaction support
- Connection pooling via HikariCP

**Database Access Examples:**

```kotlin
// Find user by email
val user = database.users.find { it.email eq "user@example.com" }

// Find manga by ID
val manga = database.manga.find { x -> x.id eq mangaId }

// Query with pagination
val mangaList = database.manga.drop(offset).take(limit).map { it.toManga() }

// Database transaction
database.useTransaction {
    // Multiple operations in a transaction
}

// Update user
user.update()

// Delete with cascade (foreign key ON DELETE CASCADE)
```

### Connection Management

- **Connection Pool**: HikariCP
- **Idle Timeout**: Configurable
- **Max Pool Size**: Configurable
- **Initialization**: Via `configureDatabase()` in Application.kt

### Migrations

- **Tool**: Flyway
- **Location**: `src/main/resources/db/migration/`
- **Naming Convention**: `V{version}__description.sql`
- **Auto-run**: On application startup

**Existing Migrations:**

1. `V1__create_users.sql` - User base table
2. `V2__create_manga_and_tags.sql` - Manga and tags tables
3. `V3__create_history_category_and_favourites.sql` - Sync data tables
4. `V4__add_fk_favourites_category.sql` - Foreign key constraint
5. `V5__create_password_reset_tokens.sql` - Password reset functionality
6. `V6__add_additional_fields_to_manga_and_tags.sql` - Extra manga fields

---

## Error Handling

### HTTP Status Codes

- **200 OK** - Successful request
- **204 No Content** - Successful sync with no changes
- **400 Bad Request** - Invalid parameters, validation errors
- **401 Unauthorized** - Missing or invalid JWT token
- **404 Not Found** - Resource not found
- **429 Too Many Requests** - Rate limit exceeded
- **500 Internal Server Error** - Server error

### Error Response Format

Most errors return plain text error messages:

```
"Wrong password"
"Invalid or expired token"
"Parameter \"offset\" is missing or invalid"
```

### Status Pages

Custom error handling implemented in `plugins/StatusPages.kt`:

- Global exception handler
- Custom status responses
- Error logging

---

## External Services

### Email Service

Located in `org.kotatsu.mail.*`

**Supported Providers:**

1. **Console** (Development):
   - Prints emails to console
   - Set `MAIL_PROVIDER=console`

2. **SMTP** (Production):
   - Sends emails via SMTP server
   - Set `MAIL_PROVIDER=smtp`
   - Requires: `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, `SMTP_FROM`

**Usage:**

```kotlin
mailService.send(
    to = "user@example.com",
    subject = "Password reset",
    textBody = "Reset link: ...",
    htmlBody = "<html>...</html>"
)
```

**Email Templates:**

- `templates/mail/forgot-password.hbs` - Password reset email HTML template

---

## Application Architecture

### Plugin System

The application uses Ktor plugins for modular configuration:

1. **Serialization** (`plugins/Serialization.kt`)
   - JSON serialization/deserialization
   - Kotlinx Serialization

2. **Database** (`plugins/Database.kt`)
   - Connection pool setup
   - Ktorm database initialization

3. **Authentication** (`plugins/Authentication.kt`)
   - JWT configuration
   - Token validation

4. **Logging** (`plugins/Logging.kt`)
   - Request/response logging

5. **Mail** (`plugins/Mail.kt`)
   - Email service setup

6. **Compression** (`plugins/Compression.kt`)
   - Response compression (gzip, etc.)

7. **Rate Limiting** (`plugins/RateLimit.kt`)
   - Endpoint rate limit configuration

8. **Status Pages** (`plugins/StatusPages.kt`)
   - Error handling and custom responses

9. **Templating** (`plugins/Templating.kt`)
   - Mustache template engine for HTML emails

10. **Routing** (`plugins/Routing.kt`)
    - All route registration

### Route Organization

Routes are split by domain in `org.kotatsu.routes.*`:

- `AuthRoutes.kt` - Authentication
- `UserRoutes.kt` - User profile
- `MangaRoutes.kt` - Manga data
- `FavouriteRoutes.kt` - Favorites sync
- `HistoryRoutes.kt` - History sync
- `HealthRoutes.kt` - Health check
- `DeeplinkRoutes.kt` - Deep link support

### Resource Handlers

Located in `org.kotatsu.resources.*`:

- `User.kt` - User resource operations
- `Manga.kt` - Manga resource operations
- `Favourites.kt` - Favorites sync logic
- `History.kt` - History sync logic

---

## Deployment

### Docker

A Dockerfile is provided for containerization:

```bash
# Build image
docker build -t kotatsu-sync .

# Run container
docker run -d -p 9292:9292 \
  -e DATABASE_HOST=db_host \
  -e DATABASE_USER=db_user \
  -e DATABASE_PASSWORD=db_pass \
  -e DATABASE_NAME=kotatsu_db \
  -e JWT_SECRET=your_secret_key \
  --name kotatsu-sync kotatsu-sync
```

### Docker Compose

Provided compose files for multi-service setup:

- `docker-compose.yaml` - Base compose configuration
- `docker-compose.mysql.yaml` - MySQL service
- `docker-compose.common.yaml` - Common configurations

### Systemd Service

A systemd service file is provided: `kotatsu-sync.service`

---

## Key Development Notes for GO Rewrite

### Sync Algorithm Considerations

The favorites and history sync use timestamp-based merging:

- Client sends local data with `timestamp` field
- Server compares with server `syncTimestamp` in users table
- Newer data wins conflicts
- Soft deletes tracked via `deletedAt` field
- Both sides need to implement same merge logic

### Data Type Mappings for MariaDB

- **TINYINT(1)** → boolean (track, pinned, show_in_lib, etc.)
- **BIGINT** → int64 (timestamps in milliseconds/seconds)
- **SMALLINT** → int16 (page numbers)
- **DOUBLE** → float64 (scroll position, percent)
- **ENUM** → string (content_rating, state, order)
- **VARCHAR(n)** → string
- **FLOAT** → float32 (rating)

### Important Timestamps

- Most timestamps are **milliseconds** (created_at, updated_at in sync objects)
- Password reset expiration is **seconds** (password_reset_token_expires_at)
- Be consistent with one unit in new implementation

### Transaction Isolation

History sync uses `TransactionIsolation.READ_COMMITTED` for concurrent access

### URL/Path Conventions

- No trailing slashes on endpoints
- Path parameters in curly braces: `/manga/{id}`
- Query parameters for pagination: `?offset=0&limit=20`
- Consumer expects JSON request/response bodies (Content-Type: application/json)

### Authentication Flow for Clients

1. POST `/auth` with email/password → receive JWT token
2. Store token
3. Include `Authorization: Bearer {token}` in all authenticated requests
4. If 401 received → token expired, must re-authenticate
5. For password reset: POST `/forgot-password` → email sent → user clicks link → POST `/reset-password` with token

---

## Testing Recommendations for GO Implementation

1. **Unit Tests**
   - Password hashing and validation
   - Token generation and validation
   - Timestamp-based sync merging logic
   - Enum serialization/deserialization

2. **Integration Tests**
   - JWT token flow (generation, validation, expiration)
   - Password reset flow (request, token generation, reset)
   - Favorites sync (GET/POST merge logic)
   - History sync (GET/POST merge logic)
   - Pagination (offset/limit)

3. **Database Tests**
   - Foreign key cascades
   - Soft deletes
   - Composite primary keys
   - Unique constraints

4. **Load Tests**
   - Rate limiting effectiveness
   - Connection pooling under load
   - Concurrent sync operations

---

## Known Behavior Notes

1. **Registration Behavior**: New user registration creates account on first `/auth` call if `ALLOW_NEW_REGISTER=true`
2. **Forgotten Password Protection**: Always returns 200 for security (no email enumeration)
3. **No Content Response**: Sync endpoints return 204 when client data matches server exactly
4. **Soft Deletes**: Favorites and history use soft deletes with `deletedAt` field for sync consistency
5. **Category Composite Key**: Categories use composite key of `(id, user_id)` - important for foreign key in favorites
6. **Token Security**: Reset tokens hashed before storage, never exposed in logs/databases
7. **Sync Timestamps**: Each user has separate `favourites_sync_timestamp` and `history_sync_timestamp`

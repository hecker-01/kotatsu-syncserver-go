-- Create database and user
CREATE DATABASE kotatsu_db;
CREATE USER 'kotatsu'@'%' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON kotatsu_db.* TO 'kotatsu'@'%';
FLUSH PRIVILEGES;

USE kotatsu_db;

-- Users table
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

-- Manga table
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

-- Tags table
CREATE TABLE tags (
    id BIGINT PRIMARY KEY,
    title VARCHAR(64) NOT NULL,
    `key` VARCHAR(120) NOT NULL,
    source VARCHAR(32) NOT NULL
);

-- Manga-Tags junction table
CREATE TABLE manga_tags (
    manga_id BIGINT NOT NULL,
    tag_id BIGINT NOT NULL,
    PRIMARY KEY (manga_id, tag_id),
    INDEX idx_manga_tags_tag_id (tag_id),
    CONSTRAINT fk_manga_tags_tag_id FOREIGN KEY (tag_id) REFERENCES tags(id),
    CONSTRAINT fk_manga_tags_manga_id FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE
);

-- Categories table
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

-- Favourites table
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

-- History table
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
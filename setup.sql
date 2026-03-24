-- Create database and user
CREATE DATABASE kotatsugo;
CREATE USER 'kotatsugo'@'%' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON kotatsugo.* TO 'kotatsugo'@'%';
FLUSH PRIVILEGES;

USE kotatsugo;

-- Users table
CREATE TABLE users (
  id INT AUTO_INCREMENT PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";


CREATE TABLE users (
                       id            UUID PRIMARY KEY,
                       username      VARCHAR(255) NOT NULL UNIQUE,
                       email         VARCHAR(255) NOT NULL UNIQUE,
                       password_hash TEXT NOT NULL,
                       created_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                       updated_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


-- Таблица Метаданных файлов
CREATE TABLE file_metadata (
                               id          UUID PRIMARY KEY,
                               user_id     UUID NOT NULL,
                               filename    VARCHAR(255) NOT NULL,
                               stored_name VARCHAR(255) NOT NULL,
                               path        TEXT NOT NULL,
                               size        BIGINT NOT NULL,
                               mime_type   VARCHAR(100),
                               checksum    VARCHAR(255),
                               created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,


                               CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
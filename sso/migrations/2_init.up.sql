CREATE TABLE IF NOT EXISTS users (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    email       VARCHAR(50) UNIQUE,              -- NULL для Telegram авторизации
    pass_hash   VARCHAR(100),                    -- NULL для Telegram авторизации
    username    VARCHAR(50) UNIQUE,
    telegram_id BIGINT UNIQUE,                   -- Telegram user ID (NULL для email авторизации)
    first_name  VARCHAR(100),                    -- Имя пользователя Telegram
    last_name   VARCHAR(100),                    -- Фамилия пользователя Telegram
    photo_url   VARCHAR(2048) DEFAULT NULL,      -- URL фото профиля Telegram
    balance     REAL NOT NULL DEFAULT 0.0,       -- Баланс пользователя (float)
    auth_type   VARCHAR(20) NOT NULL DEFAULT 'email', -- 'email' или 'telegram'
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    -- Проверка: либо email+pass_hash, либо telegram_id должны быть заполнены
    CHECK (
        (auth_type = 'email' AND email IS NOT NULL AND pass_hash IS NOT NULL) OR
        (auth_type = 'telegram' AND telegram_id IS NOT NULL)
    ),
    -- Проверка: avatar_url должен быть валидной ссылкой (http:// или https://) и не слишком длинным
    CHECK (
        avatar_url IS NULL OR
        (LENGTH(avatar_url) <= 2048 AND (avatar_url LIKE 'http://%' OR avatar_url LIKE 'https://%'))
    )
);

-- Триггер для обновления updated_at при изменении записи
CREATE TRIGGER IF NOT EXISTS update_users_updated_at
    AFTER UPDATE ON users
    FOR EACH ROW
BEGIN
    UPDATE users
    SET updated_at = CURRENT_TIMESTAMP
    WHERE id = OLD.id AND created_at = OLD.created_at;
END;


CREATE TABLE IF NOT EXISTS apps
(
    id     INTEGER PRIMARY KEY,
    name   TEXT NOT NULL UNIQUE,
    secret TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS user_app_roles
(
    id      INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    app_id  INTEGER NOT NULL,
    role    TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE,
    UNIQUE (user_id, app_id)
);
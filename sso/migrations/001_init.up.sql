-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id              BIGSERIAL PRIMARY KEY,
    email           VARCHAR(50) UNIQUE,              -- NULL для Telegram авторизации
    pass_hash       VARCHAR(100),                    -- NULL для Telegram авторизации
    username        VARCHAR(50) UNIQUE,
    telegram_id     BIGINT UNIQUE,                   -- Telegram user ID (NULL для email авторизации)
    first_name      VARCHAR(100),                    -- Имя пользователя Telegram
    last_name       VARCHAR(100),                    -- Фамилия пользователя Telegram
    photo_url       VARCHAR(2048) DEFAULT NULL,      -- URL фото профиля Telegram
    balance         BIGINT NOT NULL DEFAULT 0,       -- Баланс пользователя в копейках
    reserve_balance BIGINT NOT NULL DEFAULT 0,      -- Замороженные средства в копейках
    auth_type       VARCHAR(20) NOT NULL DEFAULT 'email', -- 'email' или 'telegram'
    role            VARCHAR(20) NOT NULL DEFAULT 'user', -- Роль пользователя: 'user', 'moderator', 'admin', 'service'
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Проверка: либо email+pass_hash, либо telegram_id должны быть заполнены
    CONSTRAINT check_auth_type CHECK (
        (auth_type = 'email' AND email IS NOT NULL AND pass_hash IS NOT NULL) OR
        (auth_type = 'telegram' AND telegram_id IS NOT NULL)
    ),
    -- Проверка: photo_url должен быть валидной ссылкой
    CONSTRAINT check_photo_url CHECK (
        photo_url IS NULL OR
        (LENGTH(photo_url) <= 2048 AND (photo_url LIKE 'http://%' OR photo_url LIKE 'https://%'))
    ),
    -- Проверка: role должна быть одной из допустимых
    CONSTRAINT check_role CHECK (role IN ('user', 'moderator', 'admin', 'service'))
);

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггер для обновления updated_at при изменении записи
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();


-- Таблица приложений
CREATE TABLE IF NOT EXISTS apps (
    id      SERIAL PRIMARY KEY,
    name    TEXT NOT NULL UNIQUE,
    secret  TEXT NOT NULL UNIQUE
);

-- Таблица ролей пользователей в приложениях
CREATE TABLE IF NOT EXISTS user_app_roles (
    id      BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    app_id  INTEGER NOT NULL,
    role    TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE,
    UNIQUE (user_id, app_id)
);

-- ==================== BALANCE TRANSACTIONS ====================
-- Единая таблица для всех операций с балансом
CREATE TABLE IF NOT EXISTS transactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         BIGINT NOT NULL,                 -- ID пользователя
    app_id          INTEGER NOT NULL,                -- ID приложения
    reservation_id  UUID,                            -- ID резервирования (для commit/cancel ссылается на reserve)
    type            VARCHAR(20) NOT NULL,            -- 'deposit', 'reserve', 'commit', 'cancel', 'refund', 'withdrawal'
    amount          BIGINT NOT NULL,                 -- Сумма операции в копейках (всегда положительная)
    balance_before  BIGINT NOT NULL,                 -- Баланс до операции в копейках
    balance_after   BIGINT NOT NULL,                 -- Баланс после операции в копейках
    reserved_before BIGINT NOT NULL DEFAULT 0,       -- Reserved баланс до операции в копейках
    reserved_after  BIGINT NOT NULL DEFAULT 0,       -- Reserved баланс после операции в копейках
    description     TEXT,                            -- Описание операции
    idempotency_key VARCHAR(255) UNIQUE,             -- Ключ идемпотентности
    metadata        JSONB,                           -- JSON с дополнительными данными
    expires_at      TIMESTAMP,                       -- Время истечения резервирования (для type='reserve')
    status          VARCHAR(10) NOT NULL DEFAULT 'pending', -- Статус транзакции: 'pending', 'success', 'failed'
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE,

    -- Проверка типа транзакции
    CONSTRAINT check_transaction_type CHECK (type IN ('deposit', 'reserve', 'commit', 'cancel', 'refund', 'withdrawal')),
    -- Проверка что сумма положительная
    CONSTRAINT check_positive_amount CHECK (amount > 0),
    -- Проверка статуса транзакции
    CONSTRAINT check_transaction_status CHECK (status IN ('pending', 'success', 'failed'))
);

-- Индексы для оптимизации запросов
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_user_app ON transactions(user_id, app_id);
CREATE INDEX IF NOT EXISTS idx_transactions_created ON transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions(type);
CREATE INDEX IF NOT EXISTS idx_transactions_reservation ON transactions(reservation_id);
CREATE INDEX IF NOT EXISTS idx_transactions_idempotency ON transactions(idempotency_key);

-- Частичный индекс для активных резервирований
CREATE INDEX IF NOT EXISTS idx_transactions_pending_reserves ON transactions(user_id, type, expires_at)
    WHERE type = 'reserve';

-- Индексы для users
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- Вставка тестового приложения (только если таблица пуста)
INSERT INTO apps (name, secret)
SELECT
    'test-app-' || substring(md5(random()::text) from 1 for 8),
    'secret-' || md5(random()::text || clock_timestamp()::text)
WHERE NOT EXISTS (SELECT 1 FROM apps LIMIT 1);

-- Вставка администратора (только если нет пользователя с username 'admin')
-- Пароль: admin (bcrypt hash)
INSERT INTO users (email, pass_hash, username, auth_type, role, balance)
SELECT
    'admin@localhost',
    '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqbqeL6VZ5WQXQ4gqkuqL7auGYfnW',
    'admin',
    'email',
    'admin',
    100000
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'admin');

-- test jwt token for user admin
-- v2.public.eyJhcHBfaWQiOjEsImV4cCI6MTg1NDg5MDE3NSwicGhvdG9fdXJsIjoiIiwicm9sZSI6Miwic3ViIjoxLCJ1c2VybmFtZSI6Im1ha2hrZXRzIn2irI1mTOvDnETqtEgBAnBebr6SHE5yaQMKkGvrT7kMIaF6yjWJTBu2tEgscPAOTJ02x3pJKhyb54iVRbwcOD4E.bnVsbA


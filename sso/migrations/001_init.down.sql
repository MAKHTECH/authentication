-- Удаляем индексы
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_telegram_id;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_transactions_pending_reserves;
DROP INDEX IF EXISTS idx_transactions_idempotency;
DROP INDEX IF EXISTS idx_transactions_reservation;
DROP INDEX IF EXISTS idx_transactions_type;
DROP INDEX IF EXISTS idx_transactions_created;
DROP INDEX IF EXISTS idx_transactions_user_app;
DROP INDEX IF EXISTS idx_transactions_user_id;

-- Удаляем таблицы
DROP TABLE IF EXISTS balance_transactions;
DROP TABLE IF EXISTS user_app_roles;
DROP TABLE IF EXISTS apps;

-- Удаляем триггер и функцию
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS users;


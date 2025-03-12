11# 🔐 Auth Service (gRPC + PASETO)

![Go Version](https://img.shields.io/badge/go-1.21%2B-blue)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![gRPC](https://img.shields.io/badge/gRPC-Enabled-purple)
![SQLite](https://img.shields.io/badge/SQLite-Database-green)
![Redis](https://img.shields.io/badge/Redis-Cache-red)

🚀 **Auth Service** — это безопасный и масштабируемый сервис аутентификации на **gRPC**, основанный на **PASETO (TOKENS)** с защитой токенов и хранением сессий в **Redis**.

----

## 🌟 Функционал API

### 🛡️ **AuthService**
- 🔹 **Register** — регистрация нового пользователя
- 🔹 **Login** — вход в систему
- 🔹 **RefreshToken** — обновление токена (долгосрочные сессии)
- 🔹 **GetDevices** — просмотр всех активных сессий
- 🔹 **Logout** — выход из системы (инвалидация токена)
- 🔸 **Rate Limiter** — ограничение кол-во попыток запроса (блокировка по ip)

### 👥 **UserService**
- 🔹 **AssignRole** — назначение роли пользователю (для админов)

### 🌐 **Metrics**
- 🔹 **RequestDuration** — гистограмма времени выполнения хэндлера
- 🔹 **ErrorCounter** — счетчик ошибок


---

## 🔥 Технологический стек

| Компонент            | Описание                      |
|----------------------|-------------------------------|
| 📝 **Язык**          | Golang                        |
| ⚡ **gRPC**           | Высокопроизводительный API    |
| 🔐 **PASETO**        | Защищенная аутентификация     |
| 🗄️ **SQLite**       | Основная база данных          |
| 🔥 **Redis**         | Хранение активных сессий      |
| 📜 **slog**          | Структурированное логирование |
| 📈 **Prometheus**    | Метрики сервиса               |
| 💠 **kafka** (план)  | Авто обновление конфига        |


---
# 📌 Хранение refresh токенов в редисе:

- 🔹 **refreshToken** — сам токен
- 🔹 **fingerprint** — уникальный идентификатор устройства
- 🔹 **expiresIn** — время истечения токена
- 🔹 **ip** — IP-адрес пользователя
- 🔹 **createdAt** — время создания токена
- 🔹 **userId** — ID пользователя
- 🔹 **ua** — user-agent (информация об устройстве)

---


## 🔮 В планах:
- 🔄 Добавить авто обновление конфига через **Consul**

### Возможно будут добавлены
- 🔄 Добавить OAuth2 (google)
- 🔄 Добавить 2FA (authy)
___


## 🏗️ Структура проекта

```bash
sso/
│   📜 go.mod              # Модульные зависимости Go
│   📜 go.sum              # Контрольная сумма зависимостей
│   🚀 main.go             # Главная точка входа
│   📖 README.md           # Документация проекта
│
├── 📂 protos/             # gRPC-протофайлы
│   ├── 📂 gen/            # Сгенерированные файлы
│   │   └── 📂 go/
│   │       └── 📂 sso/
│   └── 📂 proto/          # Исходные протофайлы
│       └── 📂 sso/
│
├── 📂 cmd/                # Исполняемые файлы
│   ├── 📂 migrator/       # Мигратор базы данных
│   ├── 📂 sso/            # Основной сервис
│   │   └── 📂 tmp/        # Временные файлы
│   ├── 📂 test/           # Тестовые утилиты
│   └── 📂 tmp/            # Временные файлы
│
├── 📂 config/             # Конфигурация сервиса
│
├── 📂 internal/           # Внутренние модули
│   ├── 📂 app/            # Управление приложением
│   │   └── 📂 grpc/       # gRPC-сервер
│   ├── 📂 config/         # Внутренние настройки
│   ├── 📂 domain/         # Бизнес-логика
│   │   ├── 📂 custom_models/  # Кастомные модели
│   │   └── 📂 models/          # Основные модели
│   ├── 📂 grpc/           # gRPC-сервисы
│   │   ├── 📂 auth/
│   │   └── 📂 user/
│   ├── 📂 lib/            # Вспомогательные модули
│   │   ├── 🔐 jwt/        # Управление JWT-токенами
│   │   └── 📂 logger/     # Логирование
│   │       ├── 📂 handlers/
│   │       │   ├── 📝 slogdiscard/
│   │       │   └── 🎨 slogpretty/
│   │       └── 📂 sl/
│   ├── 📂 services/       # Основная бизнес-логика
│   │   ├── 🔑 auth/
│   │   └── 👥 user/
│   └── 📂 storage/        # Работа с базой данных
│       ├── 🛢️ logger/
│       ├── 🔥 redis/
│       └── 🗄️ sqlite/
│
├── 📂 migrations/         # Миграции базы данных
│
├── 📂 pkg/                # Пакеты утилит
│   ├── 📂 directories/
│   └── 📂 utils/
│
├── 📂 storage/            # Файлы хранения
│
└── 📂 tests/              # Тестирование
    ├── 🧪 migrations/
    └── 🔬 suite/

```


## 🗄️ Структура базы данных

```sql
CREATE TABLE IF NOT EXISTS users (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    email      VARCHAR(50) NOT NULL UNIQUE,
    pass_hash  VARCHAR(100) NOT NULL,
    username   VARCHAR(15) UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
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
```


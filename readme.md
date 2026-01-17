<p align="center">
  <img src="https://img.shields.io/badge/Go-1.24-00ADD8?style=for-the-badge&logo=go&logoColor=white" />
  <img src="https://img.shields.io/badge/gRPC-4285F4?style=for-the-badge&logo=google&logoColor=white" />
  <img src="https://img.shields.io/badge/PostgreSQL-336791?style=for-the-badge&logo=postgresql&logoColor=white" />
  <img src="https://img.shields.io/badge/Redis-DC382D?style=for-the-badge&logo=redis&logoColor=white" />
  <img src="https://img.shields.io/badge/Kafka-231F20?style=for-the-badge&logo=apachekafka&logoColor=white" />
</p>

# 🔐 Authentication / SSO Service

> **Production-grade** микросервис аутентификации и управления балансом.  
> gRPC-first · PASETO + Ed25519 · Idempotent Transactions · Event-Driven

---

## ⚡ Возможности

| Модуль | Описание |
|--------|----------|
| **Auth** | Регистрация, логин, refresh/logout, управление устройствами |
| **Telegram Login** | OAuth-like callback + синхронизация профиля |
| **User Management** | Роли, смена email / username / password / avatar |
| **Transactions** | Reserve → Commit / Cancel, Deposit, история операций |
| **Idempotency** | Гарантия отсутствия дублей в финансовых операциях |
| **Events** | Kafka-продюсер для интеграции с другими сервисами |
| **Background Jobs** | Cron-воркер для автоотмены протухших резервов |
| **Observability** | Структурированные логи (slog), метрики Prometheus |

---

## 🛠 Стек

```
Go 1.24  ·  gRPC / Protobuf  ·  PASETO (Ed25519)
PostgreSQL  ·  Redis  ·  Kafka (Sarama)  ·  Docker
```

---

## 📡 API

> Контракты: `protos/proto/sso/*.proto`

<table>
<tr>
<td valign="top">

**AuthService**
- `Register`
- `Login`
- `RefreshToken`
- `GetDevices`
- `Logout`

</td>
<td valign="top">

**UserService**
- `AssignRole`
- `ChangeAvatar`
- `ChangeUsername`
- `ChangePassword`
- `ChangeEmail`

</td>
<td valign="top">

**TransactionsService**
- `Reserve`
- `CommitReserve`
- `CancelReserve`
- `GetBalance`
- `Deposit`
- `GetTransactions`

</td>
</tr>
</table>

---

## 💾 Модели данных

<details>
<summary><b>Session</b> (Redis)</summary>

```json
{
  "refreshToken": "v4.public.eyJ...",
  "fingerprint": "device_abc123",
  "expiresIn": 1737158400,
  "ip": "192.168.1.42",
  "createdAt": 1734566400,
  "userId": "usr_7f3a2b",
  "userAgent": "Mozilla/5.0..."
}
```
</details>

<details>
<summary><b>Transaction</b> (PostgreSQL)</summary>

```json
{
  "id": "txn_b6716b6a",
  "type": "RESERVE",
  "amount": 1500,
  "balance_after": 8500,
  "reserved_after": 1500,
  "description": "Order #1234",
  "created_at": 1737158400,
  "reservation_id": "resv_9d1f4e"
}
```
</details>

---

## 📁 Структура

```
authentication/
├── deployments/docker/          # Docker & Compose
├── protos/
│   ├── gen/go/                  # Generated Go code
│   └── proto/sso/               # .proto definitions
└── sso/
    ├── cmd/                     # Entrypoints (sso, migrator, genkey, genjwt)
    ├── config/                  # local.json, prometheus.yml
    ├── internal/
    │   ├── app/                 # Wiring: gRPC, HTTP, Cron
    │   ├── domain/              # Domain models
    │   ├── gprc/                # gRPC handlers & middleware
    │   ├── http/                # HTTP (Telegram callback)
    │   ├── lib/                 # JWT, Kafka, Logger, RateLimiter
    │   ├── repository/          # Postgres & Redis repos
    │   └── services/            # Business logic
    ├── migrations/              # SQL migrations
    ├── pkg/                     # Shared utilities
    └── tests/                   # Integration tests
```

---

<p align="center">
  <sub>Built with ❤️ and Go</sub>
</p>

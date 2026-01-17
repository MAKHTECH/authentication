package repository

import (
	"context"
	"time"

	ssov1 "sso/protos/gen/go/sso"
	"sso/sso/internal/domain/models"
)

// AuthRepository интерфейс для работы с авторизацией пользователей
type AuthRepository interface {
	// SaveUser сохраняет пользователя с email авторизацией
	SaveUser(ctx context.Context, email, username, passHash string) (int64, error)

	// SaveTelegramUser сохраняет пользователя, авторизованного через Telegram
	SaveTelegramUser(ctx context.Context, telegramID int64, username, firstName, lastName, photoURL string) (int64, error)

	// UpdateTelegramUser обновляет данные пользователя Telegram
	UpdateTelegramUser(ctx context.Context, telegramID int64, username, firstName, lastName, photoURL string) error

	// User находит пользователя по username
	User(ctx context.Context, username string, appID int) (*models.User, error)

	// UserByID находит пользователя по ID
	UserByID(ctx context.Context, id int) (*models.User, error)

	// UserByTelegramID находит пользователя по Telegram ID
	UserByTelegramID(ctx context.Context, telegramID int64, appID int) (*models.User, error)
}

// UserRepository интерфейс для работы с пользователями
type UserRepository interface {
	// AssignRole assigns a role to a user
	AssignRole(ctx context.Context, userID uint32, appID int, role ssov1.Role) error

	// CheckPermission проверяет права пользователя
	CheckPermission(ctx context.Context, userID int, appID int) error

	// ChangePhoto обновляет URL фото профиля пользователя
	ChangePhoto(ctx context.Context, userID int, photoURL string) error

	// ChangeUsername обновляет username пользователя
	ChangeUsername(ctx context.Context, userID int, username string) error

	// ChangeEmail обновляет email пользователя
	ChangeEmail(ctx context.Context, userID int, newEmail string) error

	// ChangePassword обновляет пароль пользователя
	ChangePassword(ctx context.Context, userID int, newPassword string) error

	// UserByID находит пользователя по ID
	UserByID(ctx context.Context, id int) (*models.User, error)
}

// AppRepository интерфейс для работы с приложениями
type AppRepository interface {
	// App возвращает приложение по ID
	App(ctx context.Context, appID int32) (models.App, error)
}

// SessionRepository интерфейс для работы с сессиями (Redis)
type SessionRepository interface {
	// SaveRefreshSession сохраняет refresh сессию
	SaveRefreshSession(ctx context.Context, rs *models.RefreshSession, refreshTTL time.Duration) error

	// GetRefreshSession получает refresh сессию по fingerprint
	GetRefreshSession(ctx context.Context, fingerprint string) (*models.RefreshSession, error)

	// GetRefreshSessionsByUserId возвращает все сессии пользователя по его ID
	GetRefreshSessionsByUserId(ctx context.Context, userID string) ([]*models.RefreshSession, error)

	// DeleteRefreshSession удаляет refresh сессию
	DeleteRefreshSession(ctx context.Context, fingerprint, id string) error
}

type RTransactionRepository interface {
	// SaveIdempotentKey сохраняет идемпотентный ключ вместе с информацией о транзакции
	SaveIdempotentKey(ctx context.Context, transaction *models.RedisTransaction) error

	// GetIdempotentKey получает информацию о транзакции по идемпотентному ключу
	GetIdempotentKey(ctx context.Context, idempotentKey string) (*models.RedisTransaction, error)

	// SetIdempotentKeyStatus обновляет статус идемпотентного ключа
	SetIdempotentKeyStatus(ctx context.Context, idempotentKey string, status ssov1.TransactionStatus) error

	// DeleteIdempotentKey удаляет идемпотентный ключ
	DeleteIdempotentKey(ctx context.Context, idempotentKey string) error
}

// TransactionRepository интерфейс для работы с транзакциями баланса (PostgreSQL)
type TransactionRepository interface {
	// Reserve создает резервирование средств в одной транзакции
	// Возвращает созданную транзакцию или существующую при дубликате idempotency_key
	Reserve(ctx context.Context, userID int64, appID int32, amount int64, idempotencyKey string, description string, expiresAt time.Time) (*models.Transaction, error)

	// GetTransactionByIdempotencyKey получает транзакцию по idempotency_key
	GetTransactionByIdempotencyKey(ctx context.Context, idempotencyKey string) (*models.Transaction, error)

	// GetReservationByID получает резервирование по ID с блокировкой FOR UPDATE
	GetReservationByID(ctx context.Context, reservationID string) (*models.Transaction, error)

	// Commit подтверждает резервирование и списывает средства
	// Возвращает транзакцию commit или ошибку
	Commit(ctx context.Context, reservationID string, commitIdempotencyKey string) (*models.Transaction, error)

	// Cancel отменяет резервирование и возвращает средства
	// Возвращает транзакцию cancel или ошибку
	Cancel(ctx context.Context, reservationID string, cancelIdempotencyKey string) (*models.Transaction, error)

	// CancelExpiredReservation отменяет истёкшее резервирование и возвращает средства
	// Возвращает транзакцию отмены или ошибку
	CancelExpiredReservation(ctx context.Context, reservationID string) (*models.Transaction, error)

	// GetExpiredReservations возвращает список ID истёкших резервирований (status='pending', expires_at < NOW())
	// limit - максимальное количество записей для обработки за раз
	GetExpiredReservations(ctx context.Context, limit int) ([]string, error)

	// GetTransactionsByUserID возвращает список транзакций пользователя с пагинацией
	// Сортировка по дате создания (новые первыми)
	GetTransactionsByUserID(ctx context.Context, userID int64, limit, offset int) ([]*models.Transaction, int32, error)
}

// PostgresRepository объединяет все PostgreSQL репозитории
type PostgresRepository interface {
	AuthRepository
	UserRepository
	AppRepository
	TransactionRepository
	Close() error
}

// RedisRepository объединяет все Redis репозитории
type RedisRepository interface {
	SessionRepository
	RTransactionRepository
}

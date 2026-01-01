package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/sso/internal/config"
	"sso/sso/internal/domain/models"
	user_jwt "sso/sso/internal/lib/jwt"
	"sso/sso/internal/lib/kafka"
	"sso/sso/internal/lib/logger/sl"
	"sso/sso/internal/storage"
	"sso/sso/pkg/utils"
	"strconv"
	"time"
)

type Auth struct {
	log             *slog.Logger
	usrSaver        UserSaver
	usrProvider     UserProvider
	appProvider     AppProvider
	sessionProvider SessionsProvider
	cfg             *config.Config
	producer        *kafka.Producer
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidApp         = errors.New("invalid app id")
	InvalidRefreshToken   = errors.New("invalid refresh token")
)

type TelegramService interface {
	LoginTelegram(ctx context.Context, telegramUser models.TelegramAuthUser) (*models.TokenPair, error)
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		username string,
		passHash string,
	) (uid int64, err error)
	SaveTelegramUser(
		ctx context.Context,
		telegramID int64,
		username string,
		firstName string,
		lastName string,
		photoURL string,
	) (uid int64, err error)
	UpdateTelegramUser(
		ctx context.Context,
		telegramID int64,
		username string,
		firstName string,
		lastName string,
		photoURL string,
	) error
}

type SessionsProvider interface {
	SaveRefreshSession(ctx context.Context, rs *models.RefreshSession, refreshTTL time.Duration) error
	GetRefreshSession(ctx context.Context, fingerprint string) (*models.RefreshSession, error)
	GetRefreshSessionsByUserId(ctx context.Context, id string) ([]*models.RefreshSession, error)
	DeleteRefreshSession(ctx context.Context, fingerprint, id string) error
}

type UserProvider interface {
	User(ctx context.Context, username string, appID int) (*models.User, error)
	UserByID(ctx context.Context, id int) (*models.User, error)
	UserByTelegramID(ctx context.Context, telegramID int64, appID int) (*models.User, error)

	//CheckUserPermission(ctx context.Context, userID int, appID int32, permission string) error
}

type AppProvider interface {
	App(ctx context.Context, appID int32) (models.App, error)
}

func New(
	log *slog.Logger,
	cfg *config.Config,
	producer *kafka.Producer,
	userSaver UserSaver, userProvider UserProvider, appProvider AppProvider, sessionProvider SessionsProvider,
) *Auth {
	return &Auth{
		log:             log,
		usrSaver:        userSaver,
		usrProvider:     userProvider,
		appProvider:     appProvider,
		sessionProvider: sessionProvider,
		cfg:             cfg,
		producer:        producer,
	}
}

func (a *Auth) Login(ctx context.Context, user models.AuthUser) (*models.TokenPair, error) {
	const op = "auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", user.Username),
		slog.Int("app_id", int(user.AppID)),
	)

	log.Info("attempting to login user")

	//get fingerprint and clientIP with context
	fingerprint := ctx.Value("fingerprint").(string)
	clientIp := ctx.Value("ip").(string)
	userAgent := ctx.Value("user-agent").(string)

	// Check App ID
	app, err := a.appProvider.App(ctx, user.AppID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("invalid app id", sl.Err(err))
			return nil, ErrInvalidApp
		}

		log.Error("failed to get app", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	userObj, err := a.usrProvider.User(ctx, user.Username, int(user.AppID))
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))
			return nil, ErrUserNotFound
		}
		a.log.Error("failed to get user", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Проверяем, что пользователь зарегистрирован через email
	if !userObj.IsEmailUser() {
		a.log.Error("user registered via telegram, cannot login with password")
		return nil, ErrInvalidCredentials
	}

	// Проверяем, что PassHash не nil
	if userObj.PassHash == nil {
		a.log.Error("invalid credentials", sl.Err(errors.New("password hash is nil")))
		return nil, ErrInvalidCredentials
	}

	if ok := utils.ComparePasswordHash(user.Password, a.cfg.Secret, *userObj.PassHash); !ok {
		a.log.Error("invalid credentials", sl.Err(errors.New("invalid password")))
		return nil, ErrInvalidCredentials
	}

	// Удаляем из redis токен refresh
	ctx, _ = context.WithTimeout(context.Background(), 3*time.Second)
	if err := a.sessionProvider.DeleteRefreshSession(ctx, fingerprint, strconv.Itoa(int(userObj.ID))); err != nil {
		a.log.Error("failed to delete refresh token from Redis", sl.Err(err))
		return nil, err
	}

	// Генерируем access и refresh токен, попутно занеся в redis
	tokenPair, exp, error := user_jwt.CreateTokenPair(userObj, a.cfg)
	if error != nil {
		return nil, error
	}

	// Заносим Refresh Token в Redis хранилище
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	if err = a.sessionProvider.SaveRefreshSession(ctx, &models.RefreshSession{
		RefreshToken: tokenPair.RefreshToken,
		UserId:       strconv.FormatInt(userObj.ID, 10),
		Ua:           userAgent,
		Ip:           clientIp,
		Fingerprint:  fingerprint,
		ExpiresIn:    time.Duration(exp),
		CreatedAt:    time.Now(),
	}, a.cfg.Jwt.RefreshTokenTTL); err != nil {
		return nil, err
	}
	log.Debug("Refresh token saved to Redis")

	log.Info("user logged in successfully", slog.String("app_name", app.Name))

	return &models.TokenPair{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, user models.AuthUser) (*models.TokenPair, error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", user.Email),
	)

	log.Info("registering user")

	//get fingerprint and clientIP with context
	fingerprint := ctx.Value("fingerprint").(string)
	clientIp := ctx.Value("ip").(string)
	userAgent := ctx.Value("user-agent").(string)

	// Check App ID
	app, err := a.appProvider.App(ctx, user.AppID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("invalid app id", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, ErrInvalidApp)
		}

		log.Error("failed to get app", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// save user to storage
	hashPassword := utils.PasswordToHash(user.Password, a.cfg.Secret)
	log.Debug("СОХРАНЯЮ ПОЛЬЗОВАТЕЛЯ, ЗАХОЖУ В usrSaver.SaveUser")
	id, err := a.usrSaver.SaveUser(ctx, user.Email, user.Username, hashPassword)
	log.Debug("ВЫХОЖУ ИЗ, ЗАХОЖУ В usrSaver.SaveUser")
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", sl.Err(err))

			return nil, fmt.Errorf("%s: %w", op, ErrUserExists)
		}

		log.Error("failed to save user", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Create User Object
	userObj := &models.User{
		ID:       id,
		Email:    &user.Email,
		Username: user.Username,
		PassHash: &hashPassword,
		AuthType: models.AuthTypeEmail,
		AppID:    user.AppID,
	}

	tokenPair, exp, err := user_jwt.CreateTokenPair(userObj, a.cfg)
	if err != nil {
		return nil, err
	}

	// Заносим Refresh Token в Redis хранилище
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	if err = a.sessionProvider.SaveRefreshSession(ctx, &models.RefreshSession{
		RefreshToken: tokenPair.RefreshToken,
		UserId:       strconv.FormatInt(userObj.ID, 10),
		Ua:           userAgent,
		Ip:           clientIp,
		Fingerprint:  fingerprint,
		ExpiresIn:    time.Duration(exp),
		CreatedAt:    time.Now(),
	}, a.cfg.Jwt.RefreshTokenTTL); err != nil {
		return nil, err
	}
	log.Debug("Refresh token saved to Redis")

	log.Info("user registered", slog.String("app", app.Name))

	return tokenPair, nil
}

func (a *Auth) RefreshToken(ctx context.Context, refreshToken string) (*models.TokenPair, error) {
	const op = "auth.RefreshToken"

	log := a.log.With(
		slog.String("op", op),
		slog.String("refresh_token", refreshToken),
	)
	log.Info("refreshing token")

	//get fingerprint and clientIP with context
	fingerprint := ctx.Value("fingerprint").(string)
	clientIp := ctx.Value("ip").(string)
	userAgent := ctx.Value("user-agent").(string)

	// Get Refresh Session from Redis
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	rs, err := a.sessionProvider.GetRefreshSession(ctx, fingerprint)
	if err != nil {
		return nil, err
	}

	// Check Refresh Token
	fmt.Println("refresh", refreshToken)
	fmt.Println("rs.refresh", rs.RefreshToken)
	if rs.RefreshToken != refreshToken {
		return nil, InvalidRefreshToken
	}

	// Get User
	idToInt, _ := strconv.ParseInt(rs.UserId, 10, 64)
	userObj, err := a.usrProvider.UserByID(ctx, int(idToInt))
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Generate new token pair
	tokenPair, exp, err := user_jwt.CreateTokenPair(userObj, a.cfg)
	if err != nil {
		return nil, err
	}

	// Заносим Refresh Token в Redis хранилище
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	if err = a.sessionProvider.SaveRefreshSession(ctx, &models.RefreshSession{
		RefreshToken: tokenPair.RefreshToken,
		UserId:       rs.UserId,
		Ua:           userAgent,
		Ip:           clientIp,
		Fingerprint:  fingerprint,
		ExpiresIn:    time.Duration(exp),
		CreatedAt:    time.Now(),
	}, a.cfg.Jwt.RefreshTokenTTL); err != nil {
		return nil, err
	}
	log.Debug("Refresh token saved to Redis")

	log.Info("token refreshed")

	return tokenPair, nil
}

func (a *Auth) Logout(ctx context.Context, accessToken string) (bool, error) {
	const op = "auth.Logout"

	log := a.log.With(
		slog.String("op", op),
		slog.String("access_token", accessToken),
	)

	// parse access token
	accessTokenData, err := user_jwt.ParseToken(accessToken, true, a.cfg.Secret)
	if err != nil {
		a.log.Error("failed to parse access token", sl.Err(err))
		return false, err
	}

	// delete refresh token from Redis
	fingerprint := ctx.Value("fingerprint").(string)
	userID := strconv.FormatUint(uint64(accessTokenData.UserID), 10)
	if err := a.sessionProvider.DeleteRefreshSession(ctx, fingerprint, userID); err != nil {
		a.log.Error("failed to delete refresh token from Redis", sl.Err(err))
		return false, err
	}

	log.Info("user logout")

	return true, nil
}

func (a *Auth) GetDevices(ctx context.Context, userID int32) ([]*models.RefreshSession, error) {
	const op = "auth.GetDevices"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("user_id", int(userID)),
	)
	log.Info("getting devices")

	// Get Refresh Sessions from Redis
	sessions, err := a.sessionProvider.GetRefreshSessionsByUserId(ctx, strconv.Itoa(int(userID)))
	if err != nil {
		log.Error("failed to get refresh sessions from Redis", sl.Err(err))
		return nil, err
	}

	return sessions, nil
}

// LoginTelegram авторизует пользователя через Telegram.
// Если пользователь не найден, создает нового.
func (a *Auth) LoginTelegram(ctx context.Context, telegramUser models.TelegramAuthUser) (*models.TokenPair, error) {
	const op = "auth.LoginTelegram"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("telegram_id", telegramUser.TelegramID),
		slog.String("username", telegramUser.Username),
		slog.Int("app_id", int(telegramUser.AppID)),
	)

	log.Info("attempting to login user via Telegram")

	// Получаем fingerprint и clientIP из контекста
	fingerprint := ctx.Value("fingerprint").(string)
	clientIp := ctx.Value("ip").(string)
	userAgent := ctx.Value("user-agent").(string)

	// Проверяем App ID
	app, err := a.appProvider.App(ctx, telegramUser.AppID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("invalid app id", sl.Err(err))
			return nil, ErrInvalidApp
		}
		log.Error("failed to get app", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	_ = app

	// Пытаемся найти пользователя по Telegram ID
	userObj, err := a.usrProvider.UserByTelegramID(ctx, telegramUser.TelegramID, int(telegramUser.AppID))
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			// Пользователь не найден - регистрируем нового
			log.Info("telegram user not found, creating new user")
			return a.registerTelegramUser(ctx, telegramUser, fingerprint, clientIp, userAgent)
		}
		log.Error("failed to get user by telegram id", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Пользователь найден - обновляем данные и авторизуем
	if err := a.usrSaver.UpdateTelegramUser(ctx, telegramUser.TelegramID,
		telegramUser.Username, telegramUser.FirstName, telegramUser.LastName, telegramUser.PhotoURL); err != nil {
		log.Warn("failed to update telegram user data", sl.Err(err))
		// Не возвращаем ошибку, продолжаем авторизацию
	}

	// Удаляем из redis старый токен refresh
	ctx2, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := a.sessionProvider.DeleteRefreshSession(ctx2, fingerprint, strconv.Itoa(int(userObj.ID))); err != nil {
		a.log.Error("failed to delete refresh token from Redis", sl.Err(err))
		return nil, err
	}

	// Генерируем access и refresh токен
	tokenPair, exp, err := user_jwt.CreateTokenPair(userObj, a.cfg)
	if err != nil {
		return nil, err
	}

	// Заносим Refresh Token в Redis хранилище
	ctx3, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	if err = a.sessionProvider.SaveRefreshSession(ctx3, &models.RefreshSession{
		RefreshToken: tokenPair.RefreshToken,
		UserId:       strconv.FormatInt(userObj.ID, 10),
		Ua:           userAgent,
		Ip:           clientIp,
		Fingerprint:  fingerprint,
		ExpiresIn:    time.Duration(exp),
		CreatedAt:    time.Now(),
	}, a.cfg.Jwt.RefreshTokenTTL); err != nil {
		return nil, err
	}

	log.Info("telegram user logged in successfully")

	return tokenPair, nil
}

// registerTelegramUser регистрирует нового пользователя через Telegram
func (a *Auth) registerTelegramUser(ctx context.Context, telegramUser models.TelegramAuthUser, fingerprint, clientIp, userAgent string) (*models.TokenPair, error) {
	const op = "auth.registerTelegramUser"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("telegram_id", telegramUser.TelegramID),
	)

	// Сохраняем нового пользователя
	id, err := a.usrSaver.SaveTelegramUser(ctx, telegramUser.TelegramID,
		telegramUser.Username, telegramUser.FirstName, telegramUser.LastName, telegramUser.PhotoURL)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("telegram user already exists", sl.Err(err))
			return nil, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		log.Error("failed to save telegram user", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Создаем объект пользователя
	userObj := &models.User{
		ID:         id,
		Username:   telegramUser.Username,
		TelegramID: &telegramUser.TelegramID,
		FirstName:  &telegramUser.FirstName,
		LastName:   &telegramUser.LastName,
		PhotoURL:   &telegramUser.PhotoURL,
		AuthType:   models.AuthTypeTelegram,
		AppID:      telegramUser.AppID,
	}

	// Генерируем токены
	tokenPair, exp, err := user_jwt.CreateTokenPair(userObj, a.cfg)
	if err != nil {
		return nil, err
	}

	// Заносим Refresh Token в Redis хранилище
	ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = a.sessionProvider.SaveRefreshSession(ctx2, &models.RefreshSession{
		RefreshToken: tokenPair.RefreshToken,
		UserId:       strconv.FormatInt(userObj.ID, 10),
		Ua:           userAgent,
		Ip:           clientIp,
		Fingerprint:  fingerprint,
		ExpiresIn:    time.Duration(exp),
		CreatedAt:    time.Now(),
	}, a.cfg.Jwt.RefreshTokenTTL); err != nil {
		return nil, err
	}

	log.Info("telegram user registered and logged in successfully")

	return tokenPair, nil
}

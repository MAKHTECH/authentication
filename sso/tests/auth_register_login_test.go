package tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ssov1 "sso/protos/gen/go/sso"
	"sso/sso/tests/suite"
	"testing"
)

const (
	emptyAppID = 0
	appID      = 1
	appSecret  = "test-secret"

	passDefaultLen = 10
)

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	username := gofakeit.Username()
	pass := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Username: username,
		Password: pass,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetTokens())

	respLog, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Username: username,
		Password: pass,
		AppId:    appID,
	})
	require.NoError(t, err)

	//loginTime := time.Now()

	token := respLog.GetTokens()
	require.NotEmpty(t, token)

	//tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
	//	return []byte(appSecret), nil
	//})
	//require.NoError(t, err)
	//
	//claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	//assert.True(t, ok)
	//
	//assert.Equal(t, respReg.GetUserId(), int64(claims["uid"].(float64)))
	//assert.Equal(t, email, claims["email"].(string))
	//assert.Equal(t, appID, int(claims["app_id"].(float64)))

	//const deltaSecond = 1
	//
	//assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSecond)
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, true, passDefaultLen)
}

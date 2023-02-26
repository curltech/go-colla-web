package controller

import (
	"crypto/rand"
	"errors"
	"github.com/curltech/go-colla-biz/rbac/entity"
	"github.com/curltech/go-colla-biz/rbac/service"
	"github.com/curltech/go-colla-core/config"
	"github.com/curltech/go-colla-core/logger"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
	"golang.org/x/crypto/ed25519"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type UserClaims struct {
	UserName string `json:"userName"`
}

var (
	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
	signer     *jwt.Signer
	verifier   *jwt.Verifier
)

func init() {
	var err error
	_, err = os.Stat(config.RbacParams.PrivateKeyFileName)
	if err == nil {
		v, err := ioutil.ReadFile(config.RbacParams.PrivateKeyFileName)
		if err == nil {
			privateKey = ed25519.PrivateKey(v)
		}
		_, err = os.Stat(config.RbacParams.PublicKeyFileName)
		if err == nil {
			v, err = ioutil.ReadFile(config.RbacParams.PublicKeyFileName)
			if err == nil {
				publicKey = ed25519.PublicKey(v)
			}
		}
	}
	if privateKey == nil {
		publicKey, privateKey, err = ed25519.GenerateKey(rand.Reader)
		if err == nil {
			ioutil.WriteFile(config.RbacParams.PrivateKeyFileName, privateKey, 0644)
			ioutil.WriteFile(config.RbacParams.PublicKeyFileName, publicKey, 0644)
		}
	}
	signer = jwt.NewSigner(jwt.EdDSA, privateKey, time.Duration(config.RbacParams.AccessTokenMaxAge))
	verifier = jwt.NewVerifier(jwt.EdDSA, publicKey)
}

/**
产生新token，并设置cooki
*/
func GenerateToken(ctx iris.Context, user *entity.User) []byte {
	// Enable payload encryption with:
	// signer.WithEncryption([]byte(claims.Password), nil)
	token, err := CreateToken(user)
	if err != nil {
		logger.Sugar.Error(err.Error())
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())

		return nil
	}
	if token != nil && len(token) > 0 {
		ctx.SetCookieKV("token", string(token), iris.CookieExpires(time.Duration(config.RbacParams.AccessTokenMaxAge)), iris.CookieHTTPOnly(false))
	}

	return token
}

func CreateToken(user *entity.User) ([]byte, error) {
	//signer = jwt.NewSigner(jwt.EdDSA, privateKey, time.Duration(config.RbacParams.AccessTokenMaxAge))
	//token, err := signer.Sign(claims.UserName)
	claims := &UserClaims{UserName: user.UserName}
	token, err := jwt.Sign(jwt.EdDSA, privateKey, claims, jwt.MaxAge(time.Duration(config.RbacParams.AccessTokenMaxAge)))

	return token, err
}

func Protected(ctx iris.Context) {
	if !config.AppParams.EnableJwt {
		ctx.Next()
		return
	}
	// 检查是否是豁免地址
	_, ok := checkNone(ctx)
	if ok {
		ctx.Next()
		return
	}
	//检查会话中的当前用户
	currentUser := CurrentUser(ctx)
	//取得token
	token, done := getToken(ctx, "token")
	if !done {
		logger.Sugar.Error("NoToken")
		ctx.StopWithJSON(iris.StatusUnauthorized, "NoToken")

		return
	}
	// 校验token的有效性，并获取用户名
	userName, expiresAtString, timeLeft, err := VerifyToken([]byte(token), currentUser)
	if err != nil {
		logger.Sugar.Error(err.Error())
		ctx.StopWithJSON(iris.StatusUnauthorized, err.Error())

		return
	}
	logger.Sugar.Infof("%v:%v:%v", userName, expiresAtString, timeLeft)

	//缓存中必须有token中用户名的会话
	svc := service.GetUserService()
	user := svc.GetUser(userName)
	if user == nil {
		logger.Sugar.Error("NoCurrentUser")
		ctx.StopWithJSON(iris.StatusUnauthorized, "NoCurrentUser")

		return
	}
	//token中的剩余时间如果小于阀值，刷新token
	if timeLeft < time.Duration(config.RbacParams.RefreshLeftAge) {
		GenerateToken(ctx, user)
	}

	/**
	检查用户权限
	*/
	err = Check(ctx, user)
	if err != nil {
		logger.Sugar.Error(err.Error())
		ctx.StopWithJSON(iris.StatusForbidden, err.Error())

		return
	}
	ctx.Next()
}

/**
从会话中取当前用户，必须启动会话功能
*/
func CurrentUser(ctx iris.Context) *entity.User {
	var currentUser *entity.User
	if config.AppParams.EnableSession {
		return userController.CurrentUser(ctx)
	}

	return currentUser
}

func getToken(ctx iris.Context, key string) (string, bool) {
	token := ctx.URLParam(key)
	if token == "" {
		token = ctx.GetCookie(key)
		if token == "" {
			if key == "token" {
				authorization := ctx.GetHeader("Authorization")
				if authorization == "" {
					logger.Sugar.Error("authorization NoToken")

					return "", false
				} else {
					token = strings.TrimPrefix(authorization, "Bearer ")
					if token == "" {
						logger.Sugar.Error("Bearer NoToken")

						return "", false
					}
				}
			}
		}
	}
	return token, true
}

func VerifyToken(token []byte, currentUser *entity.User) (string, string, time.Duration, error) {
	//verifiedToken, err := verifier.VerifyToken(token)
	verifiedToken, err := jwt.Verify(jwt.EdDSA, publicKey, token)
	if err != nil {
		return "", "", 0, err
	}
	logger.Sugar.Infof("%v", verifiedToken)
	claims := &UserClaims{}
	err = verifiedToken.Claims(claims)
	if err != nil {
		return "", "", 0, err
	}
	if currentUser != nil && claims.UserName != currentUser.UserName {

		return claims.UserName, "", 0, errors.New("ErrorUserName")
	}
	// Just an example on how you can retrieve all the standard claims (set by jwt.MaxAge, "exp").
	standardClaims := verifiedToken.StandardClaims
	expiresAtString := standardClaims.ExpiresAt().Format(time.RFC3339Nano)
	timeLeft := standardClaims.Timeleft()
	logger.Sugar.Infof("%v:%v:%v", claims.UserName, expiresAtString, timeLeft)

	return claims.UserName, expiresAtString, timeLeft, nil
}

func GenerateTokenPair(ctx iris.Context, user *entity.User) *jwt.TokenPair {
	// Generates a Token Pair, long-live for refresh tokens, e.g. 1 hour.
	// First argument is the access claims,
	// second argument is the refresh claims,
	// third argument is the refresh max age.
	// Send the generated token pair to the client.
	// The tokenPair looks like: {"access_token": $token, "refresh_token": $token}
	tokenPair, err := CreateTokenPair(user)
	if err != nil {
		logger.Sugar.Errorf("token pair: %v", err)
		ctx.StopWithJSON(iris.StatusInternalServerError, err.Error())

		return nil
	}
	ctx.SetCookieKV("token", string(tokenPair.AccessToken), iris.CookieExpires(time.Duration(config.RbacParams.AccessTokenMaxAge)), iris.CookieHTTPOnly(false))
	ctx.SetCookieKV("refresh_token", string(tokenPair.RefreshToken), iris.CookieExpires(time.Duration(config.RbacParams.RefreshTokenMaxAge)), iris.CookieHTTPOnly(false))

	return &tokenPair
}

func CreateTokenPair(user *entity.User) (jwt.TokenPair, error) {
	accessClaims := &UserClaims{UserName: user.UserName}
	refreshClaims := jwt.Claims{Subject: user.UserName}
	tokenPair, err := signer.NewTokenPair(accessClaims, refreshClaims, time.Duration(config.RbacParams.RefreshTokenMaxAge))

	return tokenPair, err
}

// There are various methods of refresh token, depending on the application requirements.
// In this example we will accept a refresh token only, we will verify only a refresh token
// and we re-generate a whole new pair. An alternative would be to accept a token pair
// of both access and refresh tokens, verify the refresh, verify the access with a Leeway time
// and check if its going to expire soon, then generate a single access token.
func RefreshToken(ctx iris.Context, user *entity.User) *jwt.TokenPair {
	// Assuming you have access to the current user, e.g. sessions.

	// Get the refresh token from ?refresh_token=$token OR
	// the request body's JSON{"refresh_token": "$token"}.
	var refreshToken []byte
	v, ok := getToken(ctx, "refresh_token")
	if v != "" && ok {
		refreshToken = []byte(v)
	}
	if len(refreshToken) == 0 {
		// You can read the whole body with ctx.GetBody/ReadBody too.
		var tokenPair jwt.TokenPair
		if err := ctx.ReadJSON(&tokenPair); err != nil {
			logger.Sugar.Error(err.Error())
			ctx.StopWithJSON(iris.StatusUnauthorized, err.Error())

			return nil
		}

		refreshToken = tokenPair.RefreshToken
	}

	// Verify the refresh token, which its subject MUST match the "currentUserID".
	verifiedToken, err := verifier.VerifyToken(refreshToken, jwt.Expected{Subject: user.UserName})
	if err != nil {
		logger.Sugar.Errorf("verify refresh token: %v", err)
		ctx.StopWithJSON(iris.StatusUnauthorized, err.Error())

		return nil
	}

	/* Custom validation checks can be performed after Verify calls too:
	currentUserID := "53afcf05-38a3-43c3-82af-8bbbe0e4a149"
	*/
	userName := verifiedToken.StandardClaims.Subject
	if userName != user.UserName {
		logger.Sugar.Error(err.Error())
		ctx.StopWithJSON(iris.StatusUnauthorized, "username is not match")

		return nil
	}

	// All OK, re-generate the new pair and send to client,
	// we could only generate an access token as well.
	return GenerateTokenPair(ctx, user)
}

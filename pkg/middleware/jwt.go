package middleware

import (
	"fmt"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	v4 "github.com/golang-jwt/jwt/v4"
	"github.com/golang-module/carbon"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
	"net/http"
	"strings"
	"time"
)

func Jwt(options ...func(*JwtOptions)) gin.HandlerFunc {
	ops := getJwtOptionsOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	mw := initJwt(*ops)
	return func(c *gin.Context) {
		claims, err := mw.GetClaimsFromJWT(c)
		if err != nil {
			unauthorized(c, http.StatusUnauthorized, err, *ops)
			return
		}

		if claims["exp"] == nil {
			unauthorized(c, http.StatusBadRequest, jwt.ErrMissingExpField, *ops)
			return
		}

		if _, ok := claims["exp"].(float64); !ok {
			unauthorized(c, http.StatusBadRequest, jwt.ErrWrongFormatOfExp, *ops)
			return
		}

		if int64(claims["exp"].(float64)) < mw.TimeFunc().Unix() {
			unauthorized(c, http.StatusUnauthorized, jwt.ErrExpiredToken, *ops)
			return
		}

		c.Set("JWT_PAYLOAD", claims)
		identity := identity(c)

		if identity != nil {
			c.Set(mw.IdentityKey, identity)
		}

		if !authorizator(identity, c) {
			unauthorized(c, http.StatusForbidden, jwt.ErrForbidden, *ops)
			return
		}

		c.Next()
	}
}

func JwtLogin(options ...func(*JwtOptions)) gin.HandlerFunc {
	ops := getJwtOptionsOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	mw := initJwt(*ops)
	if len(ops.privateBytes) == 0 {
		panic("jwt login private bytes is empty")
	}
	return func(c *gin.Context) {
		data, err := login(c, *ops)

		if err != nil {
			unauthorized(c, http.StatusUnauthorized, err, *ops)
			return
		}

		token := v4.New(v4.GetSigningMethod(mw.SigningAlgorithm))
		claims := token.Claims.(v4.MapClaims)

		for key, value := range payload(data) {
			claims[key] = value
		}

		expire := mw.TimeFunc().Add(mw.Timeout)
		claims["exp"] = expire.Unix()
		claims["orig_iat"] = mw.TimeFunc().Unix()
		tokenString, err := signedString(mw.Key, token)

		if err != nil {
			unauthorized(c, http.StatusUnauthorized, jwt.ErrFailedTokenCreation, *ops)
			return
		}

		// set cookie
		if mw.SendCookie {
			expireCookie := mw.TimeFunc().Add(mw.CookieMaxAge)
			maxage := int(expireCookie.Unix() - mw.TimeFunc().Unix())

			if mw.CookieSameSite != 0 {
				c.SetSameSite(mw.CookieSameSite)
			}

			c.SetCookie(
				mw.CookieName,
				tokenString,
				maxage,
				"/",
				mw.CookieDomain,
				mw.SecureCookie,
				mw.CookieHTTPOnly,
			)
		}

		loginResponse(c, http.StatusOK, tokenString, expire, *ops)
	}
}

func JwtLogout(options ...func(*JwtOptions)) gin.HandlerFunc {
	ops := getJwtOptionsOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	mw := initJwt(*ops)
	return func(c *gin.Context) {
		if mw.SendCookie {
			if mw.CookieSameSite != 0 {
				c.SetSameSite(mw.CookieSameSite)
			}

			c.SetCookie(
				mw.CookieName,
				"",
				-1,
				"/",
				mw.CookieDomain,
				mw.SecureCookie,
				mw.CookieHTTPOnly,
			)
		}

		logoutResponse(c, http.StatusOK, *ops)
		c.Next()
	}
}

func JwtRefresh(options ...func(*JwtOptions)) gin.HandlerFunc {
	ops := getJwtOptionsOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	mw := initJwt(*ops)
	return func(c *gin.Context) {
		claims, err := mw.CheckIfTokenExpire(c)
		if err != nil {
			unauthorized(c, http.StatusUnauthorized, err, *ops)
			return
		}

		newToken := v4.New(v4.GetSigningMethod(mw.SigningAlgorithm))
		newClaims := newToken.Claims.(v4.MapClaims)

		for key := range claims {
			newClaims[key] = claims[key]
		}

		expire := mw.TimeFunc().Add(mw.Timeout)
		newClaims["exp"] = expire.Unix()
		newClaims["orig_iat"] = mw.TimeFunc().Unix()
		tokenString, err := signedString(mw.Key, newToken)

		if err != nil {
			unauthorized(c, http.StatusUnauthorized, err, *ops)
			return
		}

		// set cookie
		if mw.SendCookie {
			expireCookie := mw.TimeFunc().Add(mw.CookieMaxAge)
			maxage := int(expireCookie.Unix() - mw.TimeFunc().Unix())

			if mw.CookieSameSite != 0 {
				c.SetSameSite(mw.CookieSameSite)
			}

			c.SetCookie(
				mw.CookieName,
				tokenString,
				maxage,
				"/",
				mw.CookieDomain,
				mw.SecureCookie,
				mw.CookieHTTPOnly,
			)
		}

		refreshResponse(c, http.StatusOK, tokenString, expire, *ops)
	}
}

// init jwt with option
func initJwt(ops JwtOptions) *jwt.GinJWTMiddleware {
	j, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:         ops.realm,                                 // jwt flag
		Key:           []byte(ops.key),                           // server secret key
		Timeout:       time.Hour * time.Duration(ops.timeout),    // token expires
		MaxRefresh:    time.Hour * time.Duration(ops.maxRefresh), // token max refresh interval(RefreshToken=Timeout+MaxRefresh)
		TokenLookup:   ops.tokenLookup,                           // where to find token
		TokenHeadName: ops.tokenHeaderName,                       // header name
		SendCookie:    ops.sendCookie,                            // send cookie flag
		CookieName:    ops.cookieName,                            // cookie name
		TimeFunc:      time.Now,                                  // now time
	})
	if err != nil {
		panic(err)
	}
	return j
}

// check auth failed
func unauthorized(c *gin.Context, code int, err error, ops JwtOptions) {
	ops.logger.Debug(c, "jwt auth check failed, err: %d, %v", code, err)
	msg := fmt.Sprintf("%v", err)
	if msg == resp.LoginCheckErrorMsg || msg == resp.ForbiddenMsg || msg == resp.UserDisabledMsg {
		ops.failWithMsg(msg)
		return
	}
	ops.failWithCodeAndMsg(resp.Unauthorized, msg)
}

// check auth success
func authorizator(data interface{}, c *gin.Context) bool {
	if v, ok := data.(map[string]interface{}); ok {
		if userIdStr, ok := v["user"].(string); ok {
			userId := utils.Str2Int64(userIdStr)
			c.Set("user", userId)
			return true
		}
	}
	return false
}

// parse claims handler
func identity(c *gin.Context) interface{} {
	claims := jwt.ExtractClaims(c)
	return map[string]interface{}{
		"IdentityKey": claims[jwt.IdentityKey],
		"user":        claims["user"],
	}
}

func payload(data interface{}) jwt.MapClaims {
	if v, ok := data.(map[string]interface{}); ok {
		return jwt.MapClaims{
			jwt.IdentityKey: v["user"],
			"user":          v["user"],
		}
	}
	return jwt.MapClaims{}
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// custom login check
func login(c *gin.Context, ops JwtOptions) (interface{}, error) {
	var r User
	c.ShouldBind(&r)
	r.Username = strings.TrimSpace(r.Username)
	r.Password = strings.TrimSpace(r.Password)
	if r.Username == "" {
		return nil, fmt.Errorf("username is empty")
	}
	if r.Password == "" {
		return nil, fmt.Errorf("password is empty")
	}

	decodePwd, err := utils.RSADecrypt([]byte(r.Password), ops.privateBytes)
	if err != nil {
		return nil, err
	}

	// custom password check
	userId, pass := ops.loginPwdCheck(r.Username, string(decodePwd))
	if !pass {
		return nil, fmt.Errorf(resp.LoginCheckErrorMsg)
	}
	return map[string]interface{}{
		"user": fmt.Sprintf("%d", userId),
	}, nil
}

// login response
func loginResponse(c *gin.Context, code int, token string, expires time.Time, ops JwtOptions) {
	ops.successWithData(map[string]interface{}{
		"token":   token,
		"expires": carbon.Time2Carbon(expires),
	})
}

// logout response
func logoutResponse(c *gin.Context, code int, ops JwtOptions) {
	ops.success()
}

// refresh token response
func refreshResponse(c *gin.Context, code int, token string, expires time.Time, ops JwtOptions) {
	ops.successWithData(map[string]interface{}{
		"token":   token,
		"expires": carbon.Time2Carbon(expires),
	})
}

func signedString(key []byte, token *v4.Token) (string, error) {
	var tokenString string
	var err error
	tokenString, err = token.SignedString(key)
	return tokenString, err
}

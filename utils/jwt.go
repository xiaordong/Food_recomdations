package utils

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"time"
)

var MySecret = []byte("5cq2up0t")

const aTokenTime = 8 * time.Hour
const rTokenTime = 14 * 24 * time.Hour

func keyFunc(token *jwt.Token) (interface{}, error) {
	return MySecret, nil
}

// GenToken 生成 access_token 和 refresh_token
func GenToken(id string) (aToken, rToken string, err error) {
	claims := MyClaims{
		ID: id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(aTokenTime).Unix(),
			Issuer:    "Food",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	aToken, err = token.SignedString(MySecret)
	// rToken 不需要存储任何自定义数据
	rToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(rTokenTime).Unix(), // 过期时间
		Issuer:    "Food",                            // 签发人
	}).SignedString(MySecret)
	return aToken, rToken, nil
}

type MyClaims struct {
	ID string `json:"id"`
	jwt.StandardClaims
}

// ParasToken 解析 access_token
func ParasToken(aToken string) (claims *MyClaims, err error) {
	var token *jwt.Token
	claims = new(MyClaims)
	token, err = jwt.ParseWithClaims(aToken, claims, keyFunc)
	if err != nil {
		return nil, err
	}
	if !token.Valid { // token 是否有效
		err = errors.New("invalidToken")
		return nil, err
	}
	return claims, nil
}

// NewToken 刷新token
func NewToken(aToken, rToken string) (newToken, newRToken string, err error) {
	if _, err = jwt.Parse(rToken, keyFunc); err != nil {
		return "", "", err
	}
	var claims MyClaims
	token, err := jwt.ParseWithClaims(aToken, &claims, keyFunc)
	if err != nil || !token.Valid {
		return "", "", errors.New("invalid access_token")
	}
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return GenToken(claims.ID)
	}
	return "", "", nil
}

func ParseSet(c *gin.Context) string {
	ID, exists := c.Get("id")
	if !exists {
		return ""
	}
	return ID.(string)
}

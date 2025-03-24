package utils

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"log"
	"time"
)

var MundoSecret []byte
var OffercatSecret []byte

func InitSecret() {
	MundoSecret = []byte(viper.GetString("jwt.jwt_sec") + "mundo")
	//log.Println("MundoSecret: ", MundoSecret)
	//OffercatSecret = []byte(config.GetConfig().Jwt.Offercat + "offercat")
	//log.Println("OffercatSecret: ", OffercatSecret)
}

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	// StandardClaims 已经弃用，使用 RegisteredClaims
	jwt.RegisteredClaims
}

// 生成 JWT Token
func GenerateToken(userID int64, username, role string, service string) (string, error) {
	now := time.Now()
	expireTime := now.Add(24 * 7 * time.Hour) // Token 有效期 一周

	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime), // 转换为 *NumericDate
			IssuedAt:  jwt.NewNumericDate(now),        // 转换为 *NumericDate
			Issuer:    viper.GetString("jwt.issuer"),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	if service == "mundo" {
		return token.SignedString(MundoSecret)
	} else {
		return token.SignedString(OffercatSecret)
	}
}

// 验证 JWT Token
func ParseToken(service string, tokenString string) (*Claims, error) {
	if service == "mundo" {
		return parseToken(MundoSecret, tokenString)
	} else {
		return parseToken(OffercatSecret, tokenString)
	}
}

func parseToken(secret []byte, tokenString string) (*Claims, error) {
	log.Println("tokenString: ", tokenString)
	//log.Println("OffercatSecret: ", OffercatSecret)
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

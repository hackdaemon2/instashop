package util

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/hackdaemon2/instashop/config"
	"github.com/hackdaemon2/instashop/model"
)

type JwtData struct {
	Issuer     string `json:"issuer"`
	Token      string `json:"token"`
	Expiry     int64  `json:"expires"`
	DateIssued int64  `json:"issued"`
	UserID     string `json:"user_id"`
}

func newJwtData(strToken, userID string, exp int64, iss time.Time) JwtData {
	return JwtData{
		Token:      strToken,
		Expiry:     exp,
		Issuer:     "instashop",
		DateIssued: iss.Unix(),
		UserID:     userID,
	}
}

func GenerateJWT(userID string, role model.Role) (JwtData, error) {
	iss := time.Now()
	exp := iss.Add(time.Hour * 24).Unix()
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     exp,
		"role":    role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	strToken, err := token.SignedString([]byte(config.GetEnv("SECRET_KEY")))
	return newJwtData(strToken, userID, exp, iss), err
}

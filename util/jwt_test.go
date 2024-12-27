package util

import (
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/hackdaemon2/instashop/config"
	"github.com/hackdaemon2/instashop/model"
	"github.com/stretchr/testify/assert"
)

func TestNewJwtData(t *testing.T) {
	token := "sampleTokenString"
	userID := "user123"
	expiry := time.Now().Add(time.Hour * 24).Unix()
	issuedAt := time.Now()
	result := newJwtData(token, userID, expiry, issuedAt)
	assert.Equal(t, token, result.Token, "Token should match")
	assert.Equal(t, expiry, result.Expiry, "Expiry should match")
	assert.Equal(t, "instashop", result.Issuer, "Issuer should be 'instashop'")
	assert.Equal(t, issuedAt.Unix(), result.DateIssued, "IssuedAt should match")
	assert.Equal(t, userID, result.UserID, "UserID should match")
}

func TestGenerateJWT(t *testing.T) {
	userID := "user123"
	role := model.Role("admin")
	jwtData, err := GenerateJWT(userID, role)

	assert.NoError(t, err, "Expected no error while generating JWT")
	assert.NotEmpty(t, jwtData.Token, "Token should not be empty")
	assert.Equal(t, "instashop", jwtData.Issuer, "Issuer should be 'instashop'")
	assert.Equal(t, userID, jwtData.UserID, "User ID should match the input")

	token, err := jwt.Parse(jwtData.Token, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.NewValidationError("unexpected signing method", jwt.ValidationErrorSignatureInvalid)
		}
		return []byte(config.GetEnv("SECRET_KEY")), nil
	})

	assert.NoError(t, err, "Expected no error while parsing the JWT")
	claims, ok := token.Claims.(jwt.MapClaims)
	assert.True(t, ok, "Expected claims to be of type jwt.MapClaims")
	assert.Equal(t, userID, claims["user_id"], "User ID should be in claims")
	assert.Equal(t, float64(jwtData.Expiry), claims["exp"], "Expiration time in claims should match")
	assert.Equal(t, string(role), claims["role"], "Role in claims should match")
}

func TestGenerateJWTExpiration(t *testing.T) {
	userID := "user123"
	role := model.Role("admin")
	jwtData, err := GenerateJWT(userID, role)
	assert.NoError(t, err)
	expiration := time.Unix(jwtData.Expiry, 0)
	assert.WithinDuration(t, time.Now().Add(24*time.Hour), expiration, time.Second, "Expiration should be 24 hours from now")
}

func TestGenerateJWTInvalidRole(t *testing.T) {
	userID := "user123"
	role := model.Role("invalidRole")
	_, err := GenerateJWT(userID, role)
	assert.NoError(t, err, "Expected no error with valid role")
}

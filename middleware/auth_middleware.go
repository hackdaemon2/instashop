package middleware

import (
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/hackdaemon2/instashop/model"
)

const (
	AUTHORIZATION_HEADER_ERROR = "Authorization header missing"
	INVALID_TOKEN_ERROR        = "Invalid token"
	FORBIDDEN_ACCESS_ERROR     = "You do not have the right to access this resource"
)

func parseToken(ctx *gin.Context) (*jwt.Token, error) {
	authHeader := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, http.ErrNoLocation
	}

	tokenStr := authHeader[7:]
	return jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		return []byte("secret_key"), nil
	})
}

func respondUnauthorized(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusUnauthorized, gin.H{"error": true, "message": message})
	ctx.Abort()
}

func IsAdmin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := parseToken(ctx)
		if err == http.ErrNoLocation {
			respondUnauthorized(ctx, AUTHORIZATION_HEADER_ERROR)
			return
		}
		if err != nil || !token.Valid {
			respondUnauthorized(ctx, INVALID_TOKEN_ERROR)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if role, ok := claims["role"].(string); ok && role == string(model.AdminRole) {
				ctx.Next()
				return
			}
		}
		ctx.JSON(http.StatusForbidden, gin.H{"error": true, "message": FORBIDDEN_ACCESS_ERROR})
		ctx.Abort()
	}
}

func Authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := parseToken(ctx)
		if err == http.ErrNoLocation {
			respondUnauthorized(ctx, AUTHORIZATION_HEADER_ERROR)
			return
		}

		if err != nil || !token.Valid {
			respondUnauthorized(ctx, INVALID_TOKEN_ERROR)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			ctx.Set("user_id", claims["user_id"])
			ctx.Next()
			return
		}

		respondUnauthorized(ctx, INVALID_TOKEN_ERROR)
	}
}

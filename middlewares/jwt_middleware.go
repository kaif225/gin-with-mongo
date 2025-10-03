package middlewares

import (
	"fmt"
	"moviestreaming/utils"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenCookie, err := c.Request.Cookie("Bearer")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Authorization cookie is missing"})
			return
		}

		jwtSecret := os.Getenv("SECRET_KEY")
		claims := &utils.SignedDetails{}

		token, err := jwt.ParseWithClaims(tokenCookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
			// Ensure signing method is HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		if claims.ExpiresAt.Time.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
			return
		}

		c.Set("role", claims.Role)
		c.Set("userID", claims.UserID)

		c.Next()
	}
}

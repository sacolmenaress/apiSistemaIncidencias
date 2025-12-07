package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/sacolmenaress/apiTesisISUM/database"
	"github.com/sacolmenaress/apiTesisISUM/models"
)

var jwtKey = []byte("clave_secreta_super_dificil")

func RequireAuth(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Se requiere autorización"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Formato de token inválido"})
		return
	}

	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token inválido o expirado"})
		return
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expirado"})
		return
	}

	var user models.User
	result := database.DB.First(&user, "email = ?", claims.Subject)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Usuario no encontrado"})
		return
	}

	c.Set("user", user)

	c.Next()
}

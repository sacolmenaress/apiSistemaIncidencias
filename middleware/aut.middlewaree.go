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

// Clave secreta (debe ser la MISMA que usaste en user.controller.go)
var jwtKey = []byte("clave_secreta_super_dificil")

// RequireAuth es nuestro "guardia de seguridad"
func RequireAuth(c *gin.Context) {
	// 1. Obtener el token del header "Authorization"
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Se requiere autorización"})
		return
	}

	// El token debe venir como "Bearer <token>"
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Formato de token inválido"})
		return
	}

	// 2. Validar el token
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token inválido o expirado"})
		return
	}

	// 3. Verificar que el token no haya expirado
	if claims.ExpiresAt.Time.Before(time.Now()) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expirado"})
		return
	}

	// 4. Buscar al usuario en la BD usando el "Subject" (email) del token
	var user models.User
	result := database.DB.First(&user, "email = ?", claims.Subject)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Usuario no encontrado"})
		return
	}

	// 5. ¡ÉXITO! Guardamos al usuario en el "contexto" de Gin.
	//    Esto permite que el *controlador* sepa quién es el usuario.
	c.Set("user", user)

	// 6. Permitir que la petición continúe hacia el controlador
	c.Next()
}

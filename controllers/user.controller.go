package controllers

import (
	"errors" // <-- ¡AÑADIR ESTE!
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/smtp"
	"strings" // <-- ¡AÑADIR ESTE!
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sacolmenaress/apiTesisISUM/database"
	"github.com/sacolmenaress/apiTesisISUM/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Clave secreta para firmar los tokens.
var jwtKey = []byte("clave_secreta_super_dificil")

// Responde a: POST /api/v1/auth/register
func Register(c *gin.Context) {
	var body struct {
		FirstName string
		LastName  string
		Email     string
		Password  string
		Role      string // "contador", "tecnico", "admin"
	}

	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	// 1. Hashear la contraseña
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al hashear la contraseña"})
		return
	}

	// 2. Crear el usuario
	user := models.User{
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Email:     body.Email,
		Password:  string(hash), // Guardamos el hash
		Role:      body.Role,
	}

	// 3. Guardar en la BD
	result := database.DB.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear el usuario"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Usuario creado exitosamente"})
}

// --- Crear Usuario por Administrador ---
// Responde a: POST /api/v1/users
func CreateUserByAdmin(c *gin.Context) {
	// 1. Obtener el usuario autenticado (el Admin)
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autenticado."})
		return
	}
	currentUser, _ := userInterface.(models.User)

	// 2. Control de Acceso: Solo 'admin' puede crear usuarios
	if currentUser.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acceso denegado. Solo administradores pueden crear usuarios."})
		return
	}

	// 3. Recibir los datos del nuevo usuario
	var body struct {
		FirstName string `json:"firstName" binding:"required"`
		LastName  string `json:"lastName" binding:"required"`
		Email     string `json:"email" binding:"required"`
		Role      string `json:"role" binding:"required"` // Rol del nuevo usuario
	}

	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: Nombre, Apellido, Email y Rol son obligatorios."})
		return
	}

	// 4. Generar Contraseña Temporal y Hash
	defaultPassword := GenerateRandomPassword(12) // Generar clave de 12 caracteres
	hash, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno al hashear la contraseña."})
		return
	}

	// 5. Crear el objeto User
	newUser := models.User{
		FirstName:          body.FirstName,
		LastName:           body.LastName,
		Email:              body.Email,
		Password:           string(hash),
		Role:               body.Role,
		MustChangePassword: true, // 👈 ¡CLAVE! Se le forzará a cambiarla en el primer login
	}

	// 6. Guardar el nuevo usuario en la DB
	result := database.DB.Create(&newUser)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear el usuario. El email podría estar duplicado."})
		return
	}

	// 7. Enviar Correo de Bienvenida (Se ejecuta en una goroutine para no bloquear la API)
	// Solo loguea un error si falla, pero el usuario ya está creado.
	// NOTA: ¡No olvides configurar tu SMTP real en SendWelcomeEmail!
	go SendWelcomeEmail(newUser.Email, newUser.FirstName, defaultPassword)

	// 8. Respuesta Exitosa
	c.JSON(http.StatusOK, gin.H{
		"message": "Usuario creado exitosamente. Se ha enviado un correo de bienvenida.",
		"userId":  newUser.ID,
	})
}

// 3. PUT /api/v1/users/:id -> Actualizar un usuario existente
func UpdateUserByAdmin(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	var body struct {
		FirstName *string `json:"firstName"` // Usar punteros para permitir campos opcionales
		LastName  *string `json:"lastName"`
		Email     *string `json:"email" binding:"omitempty,email"`
		Role      *string `json:"role" binding:"omitempty,oneof=contador tecnico admin"`
		Password  *string `json:"password" binding:"omitempty,min=6"` // Para resetear la contraseña
	}

	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: Asegúrate que el email sea correcto o la contraseña tenga mínimo 6 caracteres."})
		return
	}

	// Aplicar actualizaciones solo si el campo está presente en el body
	if body.FirstName != nil {
		user.FirstName = *body.FirstName
	}
	if body.LastName != nil {
		user.LastName = *body.LastName
	}
	if body.Email != nil {
		user.Email = *body.Email
	}
	if body.Role != nil {
		user.Role = *body.Role
	}
	if body.Password != nil {
		// Si se resetea la contraseña, la hasheamos y forzamos el cambio
		newHash, err := bcrypt.GenerateFromPassword([]byte(*body.Password), 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al hashear la nueva contraseña"})
			return
		}
		user.Password = string(newHash)
		user.MustChangePassword = true
	}

	database.DB.Save(&user)

	c.JSON(http.StatusOK, gin.H{"message": "Usuario actualizado con éxito."})
}

// --- INICIAR SESIÓN (LOGIN) ---
// Responde a: POST /api/v1/auth/login
func Login(c *gin.Context) {
	var body struct {
		Email    string
		Password string
	}

	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	// 1. Buscar al usuario por email
	var user models.User
	result := database.DB.First(&user, "email = ?", body.Email)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email o contraseña incorrectos"})
		return
	}

	// 2. Comparar la contraseña hasheada
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email o contraseña incorrectos"})
		return
	}

	// 3. Generar un token JWT
	expirationTime := time.Now().Add(24 * time.Hour) // Token válido por 24 horas
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		Subject:   user.Email, // Identificador del usuario
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo generar el token"})
		return
	}

	// 4. Enviar el token al cliente
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user":  user,
	})
}

// --- OBTENER PERFIL DE USUARIO (GET) ---
// Responde a: GET /api/v1/auth/profile
func GetProfile(c *gin.Context) {
	// 1. Obtener el usuario del contexto (que puso el middleware)
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}

	// 2. Convertir el usuario a su tipo
	user, ok := userInterface.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al procesar la información del usuario"})
		return
	}

	// 3. Responder con el usuario (sin la contraseña)
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// Actualizar la contraseña del usuario
func UpdatePassword(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}
	user, _ := userInterface.(models.User)

	// 2. Leer el JSON que nos envía el frontend
	var body struct {
		// Ya no es 'binding:"required"' para permitir el flujo de clave forzada
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword" binding:"required"`
	}

	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "La nueva contraseña es obligatoria"})
		return
	}

	// 3. Verificar la contraseña antigua (SOLO si NO está forzado a cambiar)
	// Si user.MustChangePassword es TRUE, el sistema ignora la clave antigua.
	if !user.MustChangePassword {
		// Si no está forzado, la clave antigua es obligatoria
		if body.OldPassword == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "La contraseña actual es obligatoria."})
			return
		}

		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.OldPassword))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "La contraseña actual es incorrecta"})
			return
		}
	}

	// 4. Hashear la nueva contraseña
	newHash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al hashear la nueva contraseña"})
		return
	}

	// 5. Guardar la nueva contraseña
	user.Password = string(newHash)
	user.MustChangePassword = false // <--- ¡Desactivamos el forzado!
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo actualizar la contraseña"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contraseña actualizada con éxito.", "mustChangePassword": false})
}

// --- Funciones de Utilidad ---

// Genera una contraseña alfanumérica y con símbolos de una longitud dada
func GenerateRandomPassword(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	// Inicializar la semilla para asegurar la aleatoriedad (solo una vez por proceso)
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		// Seleccionar un caracter aleatorio del conjunto 'chars'
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

// Envía el correo de bienvenida con las credenciales temporales.
// NOTA: Debes reemplazar los placeholders con tus credenciales reales de SMTP.
func SendWelcomeEmail(toEmail, firstName, password string) error {
	// ⚠️ CONFIGURACIÓN CLAVE: Reemplaza con tus datos de SMTP ⚠️
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	fromEmail := "sacolmenaress@gmail.com" // Correo del cual se envia
	smtpPassword := "xyul dizq ixmq zgnn"  // Contraseña de aplicación

	auth := smtp.PlainAuth("", fromEmail, smtpPassword, smtpHost)

	subject := "Subject: Autorización de Acceso a Sistema de Incidencias\r\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n"

	bodyContent := fmt.Sprintf(`
        <html>
        <body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
            <div style="max-width: 600px; margin: 0 auto; border: 1px solid #ddd; padding: 20px;">
                <h2 style="color: #007bff;">¡Bienvenido(a) al Sistema de Incidencias!</h2>
                <p><strong>Hola %s,</strong></p>
                <p>Nos complace informarte que has sido <strong>autorizado(a)</strong> para ingresar al sistema de gestión de incidencias de LC Consultores.</p>
                <p>A continuación, encontrarás tus datos de acceso inicial:</p>
                <ul style="list-style: none; padding: 0; background-color: #f8f9fa; padding: 10px; border-radius: 5px;">
                    <li><strong>Correo de Ingreso:</strong> <code style="font-weight: bold;">%s</code></li>
                    <li><strong>Contraseña Temporal:</strong> <code style="font-weight: bold; color: #dc3545;">%s</code></li>
                </ul>
                <p style="color: #dc3545; font-weight: bold;">⚠️ IMPORTANTE: Por tu seguridad, el sistema te pedirá que <strong>cambies esta contraseña de inmediato</strong> al iniciar sesión por primera vez. Deberás ir a la sección "Perfil".</p>
                <p><a href="http://localhost:5173" style="display: inline-block; padding: 10px 20px; background-color: #007bff; color: white; text-decoration: none; border-radius: 5px;">Ir al Sistema de Incidencias</a></p>
                <p>Saludos cordiales,</p>
                <p>El Equipo de Administración de LC Consultores.</p>
            </div>
        </body>
        </html>`, firstName, toEmail, password)

	msg := []byte(subject + mime + "\r\n" + bodyContent)

	addr := smtpHost + ":" + smtpPort

	err := smtp.SendMail(addr, auth, fromEmail, []string{toEmail}, msg)
	if err != nil {
		// Logueamos el error pero no lo hacemos fatal para que el usuario se cree
		log.Printf("ERROR: No se pudo enviar el correo a %s: %v", toEmail, err)
		return fmt.Errorf("error al enviar el correo: %w", err)
	}
	return nil
}

// controllers/user.controller.go

// ... (después de la función GetUsers)

// controllers/user.controller.go

// DeleteUser elimina un usuario por ID. Solo permitido para el rol 'admin'.
// Responde a: DELETE /api/v1/users/:id
func DeleteUser(c *gin.Context) {
	// 1. Obtener el usuario logueado del contexto
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no encontrado en el contexto"})
		return
	}
	loggedInUser, _ := userInterface.(models.User)

	// 2. Verificar el rol (Solo Admin puede eliminar)
	if loggedInUser.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acceso denegado. Solo administradores pueden eliminar usuarios."})
		return
	}

	// 3. Obtener el ID del parámetro de la URL
	id := c.Param("id")

	// 4. Prevención: No permitir que un usuario se elimine a sí mismo
	if fmt.Sprintf("%d", loggedInUser.ID) == id {
		c.JSON(http.StatusForbidden, gin.H{"error": "No puedes eliminar tu propia cuenta mientras estás logueado."})
		return
	}

	// 5. MODIFICACIÓN CRÍTICA: Eliminar en una transacción para manejar claves foráneas (tickets)
	err := database.DB.Transaction(func(tx *gorm.DB) error {

		// 5a. Eliminar todos los tickets asociados a este usuario
		// Importante: Eliminar tickets ANTES que el usuario que los creó
		if err := tx.Where("created_by_id = ?", id).Delete(&models.Ticket{}).Error; err != nil {
			return fmt.Errorf("error al eliminar tickets asociados: %w", err)
		}

		// 5b. Eliminar al usuario
		result := tx.Delete(&models.User{}, id)
		if result.Error != nil {
			return fmt.Errorf("error al eliminar usuario: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			// Se usa errors.New para poder diferenciar si el error fue que no lo encontró
			return errors.New("Usuario no encontrado para eliminar")
		}

		return nil // Si llegamos aquí, la transacción se completa con éxito (COMMIT)
	})

	// 6. Verificar el resultado de la transacción
	if err != nil {
		if strings.Contains(err.Error(), "Usuario no encontrado") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado para eliminar."})
			return
		}
		// Este error captura cualquier fallo en la BD, incluyendo fallos en la eliminación de tickets.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno al eliminar usuario. Verifica la consola del servidor Go."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usuario eliminado con éxito."})
}

// controllers/user.controller.go

// ... (después de las funciones Login, GetProfile, UpdatePassword y CreateUserByAdmin)

// GetUsers trae la lista de todos los usuarios
// Responde a: GET /api/v1/users
func GetUsers(c *gin.Context) {
	// 1. Obtener el usuario del contexto (establecido por RequireAuth)
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no encontrado en el contexto"})
		return
	}
	loggedInUser, _ := userInterface.(models.User)

	// 2. Verificar el rol
	if loggedInUser.Role != "admin" && loggedInUser.Role != "tecnico" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acceso denegado. Solo administradores y técnicos pueden ver la lista de usuarios."})
		return
	}

	// 3. Obtener todos los usuarios de la base de datos
	var users []models.User
	// GORM automáticamente excluye el campo Password por el tag `json:"-"` en el modelo.
	result := database.DB.Find(&users)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener la lista de usuarios"})
		return
	}

	// 4. Devolver la lista de usuarios
	// El frontend espera un array de objetos, se devuelve directamente 'users'.
	c.JSON(http.StatusOK, users)
}

// ... (continúa con la siguiente función)

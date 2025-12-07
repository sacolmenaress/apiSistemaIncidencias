package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sacolmenaress/apiTesisISUM/database"
	"github.com/sacolmenaress/apiTesisISUM/models"
)

func CreateIncidencia(c *gin.Context) {
	var body struct {
		Title    string `json:"title" binding:"required"`
		Solution string `json:"solution" binding:"required"`
		Category string `json:"category"`
		IsPublic bool   `json:"isPublic"`
	}

	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: Título y solución son obligatorios."})
		return
	}

	incidencia := models.Incidencia{
		Title:    body.Title,
		Solution: body.Solution,
		Category: body.Category,
		IsPublic: body.IsPublic,
	}

	result := database.DB.Create(&incidencia)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear la incidencia."})
		return
	}

	c.JSON(http.StatusCreated, incidencia)
}

func GetIncidencias(c *gin.Context) {
	var incidencias []models.Incidencia

	result := database.DB.Find(&incidencias)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al cargar las incidencias."})
		return
	}

	c.JSON(http.StatusOK, incidencias)
}

// PUT /api/v1/incidencias/:id
func UpdateIncidencia(c *gin.Context) {
	id := c.Param("id")
	var incidencia models.Incidencia

	if err := database.DB.First(&incidencia, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Incidencia no encontrada"})
		return
	}

	var body struct {
		Title    *string `json:"title"`
		Solution *string `json:"solution"`
		Category *string `json:"category"`
		IsPublic *bool   `json:"isPublic"`
	}

	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de actualización inválidos"})
		return
	}

	updates := make(map[string]interface{})
	if body.Title != nil {
		updates["Title"] = *body.Title
	}
	if body.Solution != nil {
		updates["Solution"] = *body.Solution
	}
	if body.Category != nil {
		updates["Category"] = *body.Category
	}
	if body.IsPublic != nil {
		updates["IsPublic"] = *body.IsPublic
	}

	database.DB.Model(&incidencia).Updates(updates)

	c.JSON(http.StatusOK, incidencia)
}

// GET /api/v1/incidencias/:id
func GetIncidencia(c *gin.Context) {
	id := c.Param("id")
	var incidencia models.Incidencia

	if err := database.DB.First(&incidencia, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Incidencia no encontrada"})
		return
	}

	c.JSON(http.StatusOK, incidencia)
}

// DELETE /api/v1/incidencias/:id
func DeleteIncidencia(c *gin.Context) {
	id := c.Param("id")

	result := database.DB.Delete(&models.Incidencia{}, id)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al eliminar la incidencia."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Incidencia eliminada con éxito (soft delete)."})
}

// GET /api/v1/incidencias/public
func GetPublicIncidencias(c *gin.Context) {
	var incidencias []models.Incidencia

	result := database.DB.Where("is_public = ?", true).Find(&incidencias)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al cargar las incidencias públicas de la base de datos."})
		return
	}
	c.JSON(http.StatusOK, incidencias)
}

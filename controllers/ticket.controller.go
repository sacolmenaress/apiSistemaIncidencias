package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sacolmenaress/apiTesisISUM/database"
	"github.com/sacolmenaress/apiTesisISUM/models"
)

func CreateTicket(c *gin.Context) {
	var body struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description" binding:"required"`
		Priority    string `json:"priority"` // Baja, media, alta
	}

	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos. El título y la descripción son obligatorios."})
		return
	}

	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no encontrado en el contexto"})
		return
	}
	user, _ := userInterface.(models.User)

	ticket := models.Ticket{
		Title:       body.Title,
		Description: body.Description,
		Priority:    body.Priority,
		CreatedByID: user.ID,
	}

	result := database.DB.Create(&ticket)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear el ticket"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"ticket": ticket})
}

func GetTickets(c *gin.Context) {
	var tickets []models.Ticket

	result := database.DB.Preload("User").Find(&tickets)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener los tickets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tickets": tickets})
}

func GetMyTickets(c *gin.Context) {

	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no encontrado en el contexto"})
		return
	}
	user, _ := userInterface.(models.User)

	var tickets []models.Ticket

	result := database.DB.Preload("User").Where("created_by_id = ?", user.ID).Find(&tickets)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener mis tickets"})
		return
	}

	c.JSON(http.StatusOK, tickets)
}

func GetTicket(c *gin.Context) {
	id := c.Param("id")
	var ticket models.Ticket
	if err := database.DB.Preload("User").First(&ticket, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket no encontrado"})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

func UpdateTicket(c *gin.Context) {
	id := c.Param("id")
	var ticket models.Ticket
	if err := database.DB.First(&ticket, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket no encontrado"})
		return
	}

	var body struct {
		Title       *string `json:"title"`       // Para que el usuario lo edite si está abierto
		Description *string `json:"description"` // Para que el usuario lo edite si está abierto
		Status      *string `json:"status"`      // Para que el técnico lo cambie
		Priority    *string `json:"priority"`    // Para que el técnico lo cambie
	}

	if err := c.Bind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de actualización inválidos"})
		return
	}

	updates := make(map[string]interface{})
	if body.Title != nil {
		updates["Title"] = *body.Title
	}
	if body.Description != nil {
		updates["Description"] = *body.Description
	}
	if body.Status != nil {
		updates["Status"] = *body.Status
	}
	if body.Priority != nil {
		updates["Priority"] = *body.Priority
	}

	database.DB.Model(&ticket).Updates(updates)

	c.JSON(http.StatusOK, gin.H{"message": "Ticket actualizado con éxito"})
}

func DeleteTicket(c *gin.Context) {
	id := c.Param("id")

	result := database.DB.Delete(&models.Ticket{}, id)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al eliminar el ticket"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ticket eliminado con éxito"})
}

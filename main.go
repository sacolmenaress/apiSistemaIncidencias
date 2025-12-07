package main

import (
	"github.com/gin-gonic/gin"

	"github.com/sacolmenaress/apiTesisISUM/controllers"
	"github.com/sacolmenaress/apiTesisISUM/database"
	"github.com/sacolmenaress/apiTesisISUM/middleware"
)

func init() {
	database.ConnectDB()
}

func main() {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api/v1")
	{
		//Rutas de Autenticación
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", controllers.Register)
			authRoutes.POST("/login", controllers.Login)
			authRoutes.GET("/profile", middleware.RequireAuth, controllers.GetProfile)
			authRoutes.PUT("/password", middleware.RequireAuth, controllers.UpdatePassword)
		}

		//Rutas de Usuarios (Admin/Técnico)
		userRoutes := api.Group("/users")
		userRoutes.Use(middleware.RequireAuth) // Protegida por el middleware de JWT
		{
			userRoutes.POST("", controllers.CreateUserByAdmin)
			userRoutes.DELETE("/:id", controllers.DeleteUser)
			userRoutes.GET("", controllers.GetUsers)
		}

		// Rutas de Tickets
		ticketRoutes := api.Group("/tickets").Use(middleware.RequireAuth)
		{
			ticketRoutes.POST("/", controllers.CreateTicket)  // Crear
			ticketRoutes.GET("/", controllers.GetTickets)     // Todos
			ticketRoutes.GET("/me", controllers.GetMyTickets) // Historial/Usuario
			ticketRoutes.GET("/:id", controllers.GetTicket)
			ticketRoutes.PUT("/:id", controllers.UpdateTicket)
			ticketRoutes.DELETE("/:id", controllers.DeleteTicket)
		}

		//Rutas de biblioetca de incidencias
		incidenciaRoutes := api.Group("/incidencias").Use(middleware.RequireAuth)
		{
			incidenciaRoutes.GET("/", controllers.GetIncidencias)
			incidenciaRoutes.POST("/", controllers.CreateIncidencia)
			incidenciaRoutes.GET("/:id", controllers.GetIncidencia)
			incidenciaRoutes.PUT("/:id", controllers.UpdateIncidencia)
			incidenciaRoutes.DELETE("/:id", controllers.DeleteIncidencia)
		}

		publicIncidenciaRoutes := api.Group("/incidencias")
		{
			publicIncidenciaRoutes.GET("/public", controllers.GetPublicIncidencias)
		}

	}

	r.Run(":8000")
}

package main

import (
	"log"
	"net/http"

	"event-registration-api/config"
	"event-registration-api/handlers"
	"event-registration-api/models"
	"event-registration-api/repositories"
	"event-registration-api/services"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate schema
	err = db.AutoMigrate(&models.User{}, &models.Event{}, &models.Registration{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize Repositories
	userRepo := repositories.NewUserRepository(db)
	eventRepo := repositories.NewEventRepository(db)
	regRepo := repositories.NewRegistrationRepository(db)

	// Initialize Services
	userService := services.NewUserService(userRepo)
	eventService := services.NewEventService(eventRepo, userRepo)
	regService := services.NewRegistrationService(regRepo, eventRepo, userRepo)

	// Initialize Handlers
	userHandler := handlers.NewUserHandler(userService)
	eventHandler := handlers.NewEventHandler(eventService)
	regHandler := handlers.NewRegistrationHandler(regService)

	// Setup Gin router
	r := gin.Default()
	r.SetTrustedProxies(nil)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	r.POST("/users", userHandler.CreateUser)
	r.GET("/users", userHandler.GetUsers)

	r.POST("/events", eventHandler.CreateEvent)
	r.GET("/events", eventHandler.GetEvents)
	r.GET("/events/:id", eventHandler.GetEventByID)

	r.POST("/events/:id/register", regHandler.RegisterForEvent)
	r.DELETE("/registrations/:id", regHandler.CancelRegistration)

	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

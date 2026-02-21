package handlers

import (
	"net/http"

	"event-registration-api/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RegistrationHandler struct {
	service services.RegistrationService
}

func NewRegistrationHandler(service services.RegistrationService) *RegistrationHandler {
	return &RegistrationHandler{service: service}
}

func (h *RegistrationHandler) RegisterForEvent(c *gin.Context) {
	eventIDParam := c.Param("id")
	eventID, err := uuid.Parse(eventIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID format"})
		return
	}

	var req struct {
		UserID string `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	registration, err := h.service.RegisterForEvent(c.Request.Context(), eventID, userID)
	if err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, registration)
}

func (h *RegistrationHandler) CancelRegistration(c *gin.Context) {
	regIDParam := c.Param("id")
	regID, err := uuid.Parse(regIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid registration ID format"})
		return
	}

	if err := h.service.CancelRegistration(c.Request.Context(), regID); err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "registration cancelled"})
}

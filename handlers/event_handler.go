package handlers

import (
	"net/http"

	"event-registration-api/models"
	"event-registration-api/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EventHandler struct {
	service services.EventService
}

func NewEventHandler(service services.EventService) *EventHandler {
	return &EventHandler{service: service}
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
	var event models.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// For simplicity, we assume the user ID is passed in the payload as organizer_id
	// In a real app, this should come from the auth token
	if err := h.service.CreateEvent(c.Request.Context(), event.OrganizerID, &event); err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, event)
}

func (h *EventHandler) GetEvents(c *gin.Context) {
	events, err := h.service.GetEvents(c.Request.Context())
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, events)
}

func (h *EventHandler) GetEventByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID format"})
		return
	}

	event, err := h.service.GetEventByID(c.Request.Context(), id)
	if err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, event)
}

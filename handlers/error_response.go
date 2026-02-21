package handlers

import (
	"errors"
	"net/http"

	"event-registration-api/services"

	"github.com/gin-gonic/gin"
)

func respondWithError(c *gin.Context, err error) {
	if errors.Is(err, services.ErrEventNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, services.ErrUserNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, services.ErrNoSeatsAvailable) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, services.ErrAlreadyRegistered) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, services.ErrNotOrganizer) {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// default 500
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

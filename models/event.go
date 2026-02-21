package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Event represents an event that users can register for
type Event struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Title          string         `gorm:"not null" json:"title"`
	Description    string         `json:"description"`
	OrganizerID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"organizer_id"`
	TotalCapacity  int            `gorm:"not null" json:"total_capacity"`
	AvailableSeats int            `gorm:"not null" json:"available_seats"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate generates a UUID and initialises AvailableSeats = TotalCapacity.
func (e *Event) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.AvailableSeats == 0 {
		e.AvailableSeats = e.TotalCapacity
	}
	return nil
}

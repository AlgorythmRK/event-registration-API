package services

import "errors"

// Sentinel errors as required
var (
	ErrNoSeatsAvailable  = errors.New("no seats available")
	ErrAlreadyRegistered = errors.New("already registered")
	ErrEventNotFound     = errors.New("event not found")
	ErrUserNotFound      = errors.New("user not found")
	ErrNotOrganizer      = errors.New("user is not an organizer")
)

package test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"

	"event-registration-api/models"
	"event-registration-api/repositories"
	"event-registration-api/services"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestConcurrencyRegistration(t *testing.T) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		getEnv("DB_HOST", "localhost"), getEnv("DB_USER", "postgres"), getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "event_registration"), getEnv("DB_PORT", "5432"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("Skipping integration test; db not reachable: %v", err)
	}

	// Migrate schema
	db.AutoMigrate(&models.User{}, &models.Event{}, &models.Registration{})

	userRepo := repositories.NewUserRepository(db)
	eventRepo := repositories.NewEventRepository(db)
	regRepo := repositories.NewRegistrationRepository(db)
	regService := services.NewRegistrationService(regRepo, eventRepo, userRepo)

	ctx := context.Background()

	// 1. Create Organizer (idempotent via FirstOrCreate)
	organizer := &models.User{Name: "Org", Email: "org@example.com", Role: "organizer"}
	db.Where("email = ?", organizer.Email).FirstOrCreate(organizer)

	// 2. Create a fresh event with exactly 1 seat for this test run
	event := &models.Event{
		Title:          "Concurrency Event",
		TotalCapacity:  1,
		AvailableSeats: 1,
		OrganizerID:    organizer.ID,
	}
	db.Create(event)

	// Register cleanup so each test run starts with clean registration data
	t.Cleanup(func() {
		db.Where("event_id = ?", event.ID).Delete(&models.Registration{})
		db.Delete(event)
	})

	// 3. Create 10 Users (idempotent via FirstOrCreate)
	var users []models.User
	for i := 0; i < 10; i++ {
		email := fmt.Sprintf("testuser_%d@example.com", i)
		u := models.User{Name: "TestUser", Email: email, Role: "user"}
		db.Where("email = ?", email).FirstOrCreate(&u)
		users = append(users, u)
	}

	// 4. Launch 10 goroutines simultaneously — all racing for 1 seat
	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0
	failureCount := 0

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(u models.User) {
			defer wg.Done()

			_, err := regService.RegisterForEvent(ctx, event.ID, u.ID)

			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				successCount++
			} else if errors.Is(err, services.ErrNoSeatsAvailable) ||
				errors.Is(err, services.ErrAlreadyRegistered) {
				// Both are valid "seat taken" outcomes:
				// - ErrNoSeatsAvailable: saw 0 seats after acquiring the lock
				// - ErrAlreadyRegistered: hit the unique constraint (edge case on re-run)
				failureCount++
			} else {
				t.Logf("Unexpected error: %v", err)
			}
		}(users[i])
	}

	wg.Wait()

	// 5. Assert exactly 1 success and 9 failures
	if successCount != 1 {
		t.Errorf("Expected exactly 1 success, got %d", successCount)
	}
	if failureCount != 9 {
		t.Errorf("Expected exactly 9 failures (seat taken), got %d", failureCount)
	}

	// 6. Verify DB state: available_seats must be exactly 0
	var finalEvent models.Event
	db.First(&finalEvent, "id = ?", event.ID)
	if finalEvent.AvailableSeats != 0 {
		t.Errorf("Expected available_seats = 0, got %d", finalEvent.AvailableSeats)
	}

	// 7. Print summary
	fmt.Printf("\n--- Concurrency Test Result ---\n")
	fmt.Printf("%d succeeded, %d failed — concurrency safe\n", successCount, failureCount)
	fmt.Println("-------------------------------")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

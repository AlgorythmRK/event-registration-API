# Services Layer

## Overview

The `services/` package is the business logic layer of the application. It sits between the HTTP handlers and the database repositories. It is the only layer that knows about business rules — handlers and repositories have no cross-knowledge.

Services are defined as **interfaces**, which allows them to be easily tested and swapped without changing their consumers.

---

## Files

### `user_service.go`

Manages user-related business operations.

| Function | Description |
|---|---|
| `CreateUser` | Delegates user creation to the repository |
| `GetUsers` | Returns all users from the repository |
| `GetUserByID` | Looks up a user by UUID; returns `ErrUserNotFound` if not found |

### `event_service.go`

Manages event-related business operations.

| Function | Description |
|---|---|
| `CreateEvent` | Validates that the organizer exists and has the `organizer` role; creates the event |
| `GetEvents` | Returns all events |
| `GetEventByID` | Returns a single event; returns `ErrEventNotFound` if not found |

**Business rule enforced here:**
- Only users with `role = "organizer"` may create events. If the user exists but is not an organizer, `ErrNotOrganizer` is returned.

### `registration_service.go`

Orchestrates event registration and cancellation.

| Function | Description |
|---|---|
| `RegisterForEvent` | Verifies user and event exist, then delegates to the registration repository |
| `CancelRegistration` | Delegates cancellation to the registration repository |

**Note:** The heavy concurrency logic (row-level locking, seat decrement) lives in the repository layer. The service's role here is to verify preconditions (user exists, event exists) before handing off.

### `errors.go`

Defines all domain-level sentinel errors used across the application:

```go
var (
    ErrNoSeatsAvailable  = errors.New("no seats available")
    ErrAlreadyRegistered = errors.New("already registered")
    ErrEventNotFound     = errors.New("event not found")
    ErrUserNotFound      = errors.New("user not found")
    ErrNotOrganizer      = errors.New("user is not an organizer")
)
```

These errors are returned by services (and repositories), then mapped to HTTP status codes exclusively within the handler layer via `respondWithError`.

### `repository_interfaces.go`

Declares the repository interfaces that services depend on:

```go
type UserRepository interface { ... }
type EventRepository interface { ... }
type RegistrationRepository interface { ... }
```

This file exists to prevent an **import cycle**. Without it, `services` would import `repositories`, and `repositories` would import `services` (for sentinel errors), creating a circular dependency. By defining the interfaces here, `services` is self-contained and `repositories` can implement them without importing the service package.

---

## Design Principles

- **No database code in services.** Services call repository interfaces only.
- **No HTTP concerns in services.** Services return Go errors; the handler layer decides HTTP status codes.
- **Dependency injection via interfaces.** Services accept repository interfaces in their constructors, making them decoupled and testable.
- **Sentinel errors as the error contract.** All meaningful failure states are represented by typed sentinel errors, not raw strings.

---

## Dependency Flow

```
handlers → services → repositories (via interfaces)
                ↑
          errors.go (shared sentinel errors)
```

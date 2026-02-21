# Handlers Layer

## Overview

The `handlers/` package is the HTTP presentation layer of the application. It sits at the outermost boundary and is the entry point for all incoming HTTP requests. It is built using the [Gin](https://github.com/gin-gonic/gin) framework.

Handlers do **not** contain any business logic. Their sole responsibility is to:
1. Parse and validate incoming HTTP request bodies and URL parameters.
2. Delegate work to the `services` layer.
3. Map returned errors to appropriate HTTP status codes.
4. Format and write JSON responses.

---

## Files

### `user_handler.go`

Handles user-related HTTP endpoints.

| Method | Route | Handler Function | Description |
|--------|-------|-----------------|-------------|
| `POST` | `/users` | `CreateUser` | Parses request body and creates a new user |
| `GET` | `/users` | `GetUsers` | Returns a list of all registered users |

### `event_handler.go`

Handles event-related HTTP endpoints.

| Method | Route | Handler Function | Description |
|--------|-------|-----------------|-------------|
| `POST` | `/events` | `CreateEvent` | Creates a new event; validates organizer role via service |
| `GET` | `/events` | `GetEvents` | Returns a list of all events |
| `GET` | `/events/:id` | `GetEventByID` | Returns a single event by UUID |

### `registration_handler.go`

Handles event registration and cancellation endpoints.

| Method | Route | Handler Function | Description |
|--------|-------|-----------------|-------------|
| `POST` | `/events/:id/register` | `RegisterForEvent` | Registers a user for the given event |
| `DELETE` | `/registrations/:id` | `CancelRegistration` | Cancels an existing registration by UUID |

### `error_response.go`

Centralised error response utility. The `respondWithError` function maps sentinel errors from the `services` package to the correct HTTP status codes:

| Sentinel Error | HTTP Status |
|---|---|
| `ErrEventNotFound` | `404 Not Found` |
| `ErrUserNotFound` | `404 Not Found` |
| `ErrNoSeatsAvailable` | `409 Conflict` |
| `ErrAlreadyRegistered` | `409 Conflict` |
| `ErrNotOrganizer` | `403 Forbidden` |
| *(any other error)* | `500 Internal Server Error` |

---

## Design Principles

- **No business logic in handlers.** All validation of business rules (e.g., organizer role check, seat availability) happens in the service layer.
- **Always use `respondWithError`.** All error paths go through the centralised helper to ensure consistent JSON error responses.
- **UUID validation at the boundary.** URL parameters (`:id`) are parsed and validated using `uuid.Parse()` before being passed downstream. If an ID is malformed, a `400 Bad Request` is returned immediately without hitting the database.

---

## Example: Register for Event Request Flow

```
POST /events/:id/register
        |
        ▼
RegisterForEvent (handler)
  - Parse :id (UUID)
  - Parse body: { "user_id": "..." }
  - Validate UUID format
  - Call service.RegisterForEvent(ctx, eventID, userID)
        |
        ▼
  On Success → 201 Created + Registration JSON
  On Error   → respondWithError → 404 / 409 / 500
```

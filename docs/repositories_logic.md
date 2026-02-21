# Repositories Layer

## Overview

The `repositories/` package is the data access layer of the application. It is the only layer that communicates directly with the PostgreSQL database via [GORM](https://gorm.io/). All SQL queries, transactions, and locking operations live here.

Repositories are defined as **interfaces** in the `services/` package (`repository_interfaces.go`) and implemented here, allowing the service layer to remain decoupled from the database implementation.

---

## Files

### `user_repository.go`

Handles all database operations for the `users` table.

| Function | SQL Operation | Description |
|---|---|---|
| `CreateUser` | `INSERT` | Inserts a new user record |
| `GetUserByID` | `SELECT ... WHERE id = ?` | Fetches a single user by primary key UUID |
| `GetUsers` | `SELECT *` | Fetches all user records |

### `event_repository.go`

Handles all database operations for the `events` table.

| Function | SQL Operation | Description |
|---|---|---|
| `CreateEvent` | `INSERT` | Inserts a new event with GORM `BeforeCreate` hook initialising `available_seats = total_capacity` |
| `GetEventByID` | `SELECT ... WHERE id = ?` | Fetches a single event by UUID |
| `GetEvents` | `SELECT *` | Fetches all events |

### `registration_repository.go`

The most critical file in the repository layer. Handles transactional, concurrency-safe event registration and cancellation.

---

## Booking Transaction — Step by Step

`RegisterForEvent` wraps the entire booking in a GORM transaction with the following steps:

```
BEGIN TRANSACTION
│
├── 1. SELECT * FROM events WHERE id = ? FOR UPDATE
│       → Acquires a row-level lock on this event row.
│         All other concurrent transactions attempting to lock the
│         same row will BLOCK here until this transaction commits or rolls back.
│
├── 2. Check available_seats > 0
│       → If 0, return ErrNoSeatsAvailable (transaction rolled back automatically)
│
├── 3. INSERT INTO registrations (event_id, user_id, status)
│       → If (event_id, user_id) unique index is violated,
│         return ErrAlreadyRegistered
│
├── 4. UPDATE events SET available_seats = available_seats - 1 WHERE id = ?
│       → Atomic SQL expression avoids lost updates
│
COMMIT
```

---

## Cancellation Transaction — Step by Step

`CancelRegistration` also runs inside a full transaction:

```
BEGIN TRANSACTION
│
├── 1. SELECT registration WHERE id = ?
│       → Fetch the registration to be cancelled
│
├── 2. If status == "cancelled" → return nil (idempotent)
│
├── 3. SELECT * FROM events WHERE id = ? FOR UPDATE
│       → Lock the event row before modifying seat count
│
├── 4. UPDATE registrations SET status = "cancelled"
│
├── 5. UPDATE events SET available_seats = available_seats + 1
│       → Atomic increment, mirrors the booking decrement
│
COMMIT
```

---

## Unique Violation Detection

The `isUniqueViolation` helper detects duplicate registration attempts at the DB level:

```go
func isUniqueViolation(err error) bool {
    if errors.Is(err, gorm.ErrDuplicatedKey) {
        return true
    }
    errStr := strings.ToLower(err.Error())
    return strings.Contains(errStr, "23505") ||
           strings.Contains(errStr, "unique constraint") ||
           strings.Contains(errStr, "duplicate key")
}
```

`23505` is the PostgreSQL SQLSTATE code for unique constraint violation. The string fallbacks handle driver versions that may not wrap the error using `gorm.ErrDuplicatedKey`.

---

## Why SELECT FOR UPDATE and Not sync.Mutex?

| Approach | Problem |
|---|---|
| `sync.Mutex` | Only works within a single process. Fails under horizontal scaling (multiple pods). |
| Optimistic locking | High retry/rollback rate under heavy ticket-drop concurrency. |
| **`SELECT FOR UPDATE`** ✅ | Enforced at the database level. Works across all server instances simultaneously. Crash-safe — DB rolls back automatically on failure. |

---

## Database Constraints Relied Upon

| Constraint | Table | Purpose |
|---|---|---|
| `PRIMARY KEY (id)` | all tables | UUID primary key, auto-generated via `gen_random_uuid()` |
| `UNIQUE (email)` | users | Prevents duplicate user accounts |
| `UNIQUE (event_id, user_id)` | registrations | Second line of defence against double-booking |
| `NOT NULL` | all required fields | Schema-enforced data integrity |

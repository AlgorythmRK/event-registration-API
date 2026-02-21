# Architecture & Design Decisions

## Clean Architecture
This project follows a layered architecture to ensure separation of concerns and maintainability.

```text
[HTTP Clients]
      |
[Gin Handlers (Controller Layer)] <-- Parses JSON, Maps Errors to HTTP Status Codes
      |
[Services (Business Layer)]       <-- Validates rules, Defines Sentinel Errors
      |
[Repositories (Data Access)]      <-- GORM DB Operations, Transaction Mgmt, FOR UPDATE locks
      |
[PostgreSQL DB]
```

## Database Schema Constraints
* **Users**: Email is `UNIQUE`. Role is restricted via `CHECK` constraints to 'user' or 'organizer'.
* **Events**: `total_capacity` must be > 0. `available_seats` must be >= 0 (`CHECK (available_seats >= 0)`). This provides a hard DB-level guard against overselling, even if application logic fails.
* **Registrations**: A composite unique index on `(event_id, user_id)` guarantees a user can only book one ticket for the same event.

## Avoiding Race Conditions during Booking
Booking a ticket is essentially: `read seats`, `check > 0`, `decrement seats`, `save`. Under high concurrency, this is susceptible to race conditions.

### Scenario: Without Locking (What Goes Wrong)
1. Thread A reads `available_seats = 1`.
2. Thread B reads `available_seats = 1`.
3. Thread A checks `1 > 0`, saves `available_seats = 0`.
4. Thread B checks `1 > 0`, saves `available_seats = 0`.
**Outcome:** Total seats exceeded. Two tickets given for 1 seat.

### Scenario: With Pessimistic Locking (What Goes Right)
1. Thread A invokes `SELECT * FROM events FOR UPDATE`. DB acquires lock on Event row.
2. Thread B invokes `SELECT FOR UPDATE`. Thread B is immediately blocked, waiting for Thread A to finish.
3. Thread A checks `1 > 0`, saves `available_seats = 0`, commits transaction. DB releases lock.
4. Thread B is unblocked, executes SELECT. Thread B reads `available_seats = 0`.
5. Thread B checks `0 > 0` and fails safely throwing `ErrNoSeatsAvailable`.

### Why GORM Transactions vs Application Mutex?
We chose PostgreSQL row-level locks via GORM transactions rather than a Go language-level `sync.Mutex` because:
1. **Horizontal Scalability:** An in-memory mutex only works if the application runs as a single instance. In a real-world, horizontally scaled API (e.g., Kubernetes with 5 pods), an app-level mutex is useless because requests route to different pods. The database lock applies universally across all connected clients.
2. **Crash Resilience:** If the application crashes midway, the DB automatically rolls back the transaction.

### Optimistic vs Pessimistic Locking Trade-offs
1. **Optimistic Locking**: Could be implemented using a `version` column. We attempt to update `WHERE version = old_version`. If another thread updated it first, it fails and we retry.
   * *Pros*: Faster for read-heavy apps with low collision.
   * *Cons*: With high volume "ticket drops" (1000 people racing for 50 tickets), there is incredibly high collision leading to massive retry spikes or timeouts.
2. **Pessimistic Locking**: The approach chosen (`FOR UPDATE`).
   * *Pros*: Serializes requests at the database level. Guarantees safety without retries. Highly reliable for ticketing.
   * *Cons*: Can slightly bottleneck throughput for a single event row, but entirely worth it for the integrity.

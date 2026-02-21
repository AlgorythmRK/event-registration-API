# Event Registration & Ticketing System API

A production-ready REST API built in Go for managing events, users, and seat registrations — with concurrency-safe booking enforced at the database level using PostgreSQL row-level locking.

---

## 📌 Project Overview

This system allows **organizers** to create events with a fixed seat capacity, and **users** to register for those events. The core challenge this project solves is **overbooking under concurrent load** — a classic distributed systems problem where multiple users register for the last remaining seat simultaneously.

The API handles this using:
- **PostgreSQL `SELECT FOR UPDATE`** to acquire row-level locks during booking
- A **composite unique constraint** `(event_id, user_id)` as a database-level guard against duplicate registrations
- **Sentinel errors** for clean service layer communication

It is designed for **academic evaluation** and **industry-level review**, following clean architecture principles with separated layers for models, repositories, services, and handlers.

---

## 🛠️ Tech Stack

- **Language**: Go 1.21+
- **Web Framework**: [Gin](https://github.com/gin-gonic/gin)
- **Database**: PostgreSQL
- **ORM**: [GORM](https://gorm.io/)
- **UUID Generation**: [google/uuid](https://github.com/google/uuid)
- **Config Management**: [godotenv](https://github.com/joho/godotenv)

---

## Architecture

This system follows **Clean Architecture**, with strict separation between layers:

```
Handler → Service → Repository → Database
```

| Layer | Responsibility |
|---|---|
| **Handlers** | Parse HTTP requests, validate input, return JSON responses (Gin) |
| **Services** | Enforce business rules, orchestrate repository calls, map sentinel errors |
| **Repositories** | Execute SQL queries, manage transactions, apply `SELECT FOR UPDATE` locking |
| **Database** | PostgreSQL enforces schema constraints (unique indexes, NOT NULL, foreign keys) |

---

## 📁 Project Structure

```
event-registration-api/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── config/
│   └── config.go                # Environment config loader
├── docs/                        # Layer-level documentation
│   ├── handlers_logic.md
│   ├── repositories_logic.md
│   └── services_logic.md
├── handlers/                    # HTTP request/response layer
│   ├── error_response.go
│   ├── event_handler.go
│   ├── registration_handler.go
│   └── user_handler.go
├── models/                      # GORM data models
│   ├── event.go
│   ├── registration.go
│   └── user.go
├── repositories/                # Database access layer
│   ├── event_repository.go
│   ├── registration_repository.go
│   └── user_repository.go
├── services/                    # Business logic layer
│   ├── errors.go
│   ├── event_service.go
│   ├── registration_service.go
│   ├── repository_interfaces.go
│   └── user_service.go
├── test/
│   └── concurrency_test.go      # Goroutine-based concurrency test
├── .env.example
├── go.mod
├── go.sum
├── setup.sh                     # Automated setup & smoke test script
├── DESIGN.md
└── README.md
```

---

## ⚙️ Setup Instructions

### 1. Clone the Repository

```bash
git clone https://github.com/your-username/event-registration-api.git
cd event-registration-api
```

### 2. Install Go Dependencies

```bash
go mod tidy
```

### 3. Set Up PostgreSQL

Ensure PostgreSQL is running locally. Then create the database:

```bash
psql -U postgres -c "CREATE DATABASE event_registration;"
```

### 4. Configure Environment Variables

```bash
cp .env.example .env
```

Edit `.env` with your credentials (see [Environment Variables](#-environment-variables) below).

### 5. Run Migrations

Migrations run automatically on server startup via GORM `AutoMigrate`.

### 6. Start the Server

```bash
go run cmd/server/main.go
```

The server starts on `http://localhost:8080` by default.

---

### ⚡ Automated Setup (One Command)

Alternatively, use the provided setup script to do **everything** in one shot:

```bash
chmod +x setup.sh && ./setup.sh
```

This script handles: prerequisites check → `.env` creation → database setup → dependency install → build → concurrency test → API smoke tests.

---

## 🔐 Environment Variables

Create a `.env` file at the project root:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=event_registration
SERVER_PORT=8080
```

---

## 📡 API Endpoints

| Method | Route | Description |
|--------|-------|-------------|
| `GET` | `/health` | Health check |
| `POST` | `/users` | Create a new user |
| `GET` | `/users` | List all users |
| `POST` | `/events` | Create a new event (organizer only) |
| `GET` | `/events` | List all events |
| `GET` | `/events/:id` | Get event by ID |
| `POST` | `/events/:id/register` | Register a user for an event |
| `DELETE` | `/registrations/:id` | Cancel a registration |

---

### 🔧 Sample Requests & Responses

#### Create a User

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "email": "alice@example.com", "role": "organizer"}'
```

```json
{
  "id": "96c937ff-6a91-4afb-a8b7-459755864561",
  "name": "Alice",
  "email": "alice@example.com",
  "role": "organizer"
}
```

---

#### Create an Event

```bash
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Go Workshop",
    "description": "Learn Go",
    "total_capacity": 50,
    "organizer_id": "96c937ff-6a91-4afb-a8b7-459755864561"
  }'
```

```json
{
  "id": "02bb3968-6880-4ebb-b87f-fd7658e8d62d",
  "title": "Go Workshop",
  "description": "Learn Go",
  "organizer_id": "96c937ff-6a91-4afb-a8b7-459755864561",
  "total_capacity": 50,
  "available_seats": 50,
  "created_at": "2026-02-21T08:26:23Z",
  "updated_at": "2026-02-21T08:26:23Z"
}
```

---

#### Register for an Event

```bash
curl -X POST http://localhost:8080/events/02bb3968-6880-4ebb-b87f-fd7658e8d62d/register \
  -H "Content-Type: application/json" \
  -d '{"user_id": "265681da-8475-4d7e-912c-d41edc625536"}'
```

```json
{
  "id": "c91a759e-cfb0-406d-88a8-f12a7205ef8d",
  "event_id": "02bb3968-6880-4ebb-b87f-fd7658e8d62d",
  "user_id": "265681da-8475-4d7e-912c-d41edc625536",
  "status": "confirmed",
  "created_at": "2026-02-21T08:26:25Z",
  "updated_at": "2026-02-21T08:26:25Z"
}
```

**Duplicate registration attempt:**
```json
{ "error": "already registered" }
```

---

#### Cancel a Registration

```bash
curl -X DELETE http://localhost:8080/registrations/c91a759e-cfb0-406d-88a8-f12a7205ef8d
```

```json
{ "message": "registration cancelled" }
```

---

## 🔒 Concurrency Strategy

### The Problem: Race Conditions in Booking

Without locking, two concurrent requests can both read `available_seats = 1`, both pass the seat check, and both create a registration — causing **overbooking**.

```
Goroutine A: reads available_seats = 1 → proceeds
Goroutine B: reads available_seats = 1 → proceeds  ← both win the race!
Goroutine A: creates registration, decrements to 0
Goroutine B: creates registration, decrements to -1  ← overbooking!
```

### The Fix: SELECT FOR UPDATE (Row-Level Locking)

When a booking begins, we lock the event row using PostgreSQL's `SELECT FOR UPDATE`:

```go
tx.Set("gorm:query_option", "FOR UPDATE").First(&event, "id = ?", eventID)
```

This forces concurrent transactions to **queue up** and wait. Once the first transaction commits and releases the lock, the next one reads the updated `available_seats` value — by which point it may be `0`, and the booking is correctly rejected.

### First Line of Defense — Pessimistic Locking

`SELECT FOR UPDATE` is the primary mechanism. It prevents multiple goroutines from ever reading stale seat counts simultaneously.

### Second Line of Defense — Unique Constraint

A composite unique index on `(event_id, user_id)` in the `registrations` table ensures that even in edge cases (e.g., two registrations for same user/event slip through), the **database itself** rejects duplicates at the INSERT level:

```go
EventID uuid.UUID `gorm:"uniqueIndex:idx_event_user"`
UserID  uuid.UUID `gorm:"uniqueIndex:idx_event_user"`
```

The application catches this unique violation and returns `ErrAlreadyRegistered`.

### Why Not an Application-Level Mutex?

| Approach | Problem |
|---|---|
| `sync.Mutex` in Go | Only works in a single process — fails under horizontal scaling |
| Optimistic Locking | High contention → many retries/rollbacks under load |
| **DB Row-Level Lock** ✅ | Works across multiple server instances; enforced at the data layer |

### Atomic Seat Decrement

Instead of read-modify-write with `Save()`, we use an atomic SQL expression:

```go
tx.Model(&event).Update("available_seats", gorm.Expr("available_seats - 1"))
```

This avoids lost updates if other fields on the `event` struct were modified in-flight.

### Sentinel Errors

The `services` package defines typed sentinel errors that are returned from the business logic layer. Handlers map these to appropriate HTTP status codes:

| Sentinel Error | HTTP Status | Description |
|---|---|---|
| `ErrNoSeatsAvailable` | `409 Conflict` | Event has no remaining seats |
| `ErrAlreadyRegistered` | `409 Conflict` | User is already registered for this event |
| `ErrEventNotFound` | `404 Not Found` | No event exists with the given ID |
| `ErrUserNotFound` | `404 Not Found` | No user exists with the given ID |
| `ErrNotOrganizer` | `403 Forbidden` | User does not have organizer role |

This pattern keeps error handling consistent and avoids leaking internal error strings to the client.

---

## 🧪 Running the Concurrency Test

The test in `test/concurrency_test.go` launches **10 goroutines simultaneously** to register for an event with only **1 seat**, then asserts exactly **1 success** and **9 failures**.

```bash
go test ./test/... -v -timeout 30s
```

Expected output:

```
--- Concurrency Test Result ---
1 succeeded, 9 failed — concurrency safe
-------------------------------
--- PASS: TestConcurrencyRegistration (0.60s)
```

> **Note:** The test requires a running PostgreSQL instance. If the DB is unreachable, the test is automatically skipped with `t.Skip(...)`.

---

## 🚀 Future Improvements

| Feature | Description |
|---|---|
| **JWT Authentication** | Protect endpoints with token-based auth |
| **Event Categories & Tags** | Allow filtering and discovery of events |
| **Pagination** | Support `limit` and `offset` query params for large datasets |
| **Redis Caching** | Cache event listings to reduce DB load |
| **Dockerization** | `Dockerfile` + `docker-compose.yml` for container-based deployment |
| **Waitlist System** | Queue users when seats are full; auto-confirm on cancellation |
| **Email Notifications** | Send confirmation and cancellation emails via SMTP |
| **Admin Dashboard** | Role-based management interface |

---

## 📄 License

MIT License. See [LICENSE](LICENSE) for details.

---

## 👤 Author

# AI Prompts Used — Event Registration & Ticketing API

This document contains all prompts used with AI tools (Claude by Anthropic) during 
the development of this project, as required for academic submission transparency.

Each prompt is documented with its purpose, the exact prompt text, and what it produced.

**AI Tool Used:** Claude (Anthropic)  
**Project:** Event Registration & Ticketing API  
**Language:** Go  
**Total Prompts Used:** 6

## Prompt 1 — Service Layer with Business Logic & Sentinel Errors

**Purpose:**
To implement business rules and define the error contract between layers.

**Prompt Used:**
Implement the complete service layer along with the domain error contract. Create an errors.go file defining all sentinel errors for the application (e.g., ErrNoSeatsAvailable, ErrAlreadyRegistered, ErrEventNotFound, ErrUserNotFound, ErrNotOrganizer). Create repository_interfaces.go in the core domain to define the interfaces the repository layer must implement, effectively preventing import cycles. Implement the business logic in the services, specifically enforcing organizer role validation during event creation, and performing strict pre-validation checks for user and event existence prior to processing any booking.

**Output Received:**
A highly decoupled service layer containing all core business logic, a comprehensive set of defined sentinel errors for standardized error handling, and correctly abstracted repository interfaces.

## Prompt 2 — Handler Layer with Error Mapping

**Purpose:**
To implement the HTTP presentation layer with proper status code mapping.

**Prompt Used:**
Implement all HTTP handler files strictly adhering to the rule that handlers contain zero business logic. Implement UUID validation at the HTTP boundary before passing parameters to the service layer. Create a centralized respondWithError function that consistently maps all sentinel domain errors to their correct HTTP status codes (e.g., 404 Not Found, 409 Conflict, 403 Forbidden, 500 Internal Server). Ensure all route handlers return clean, standard JSON error responses.

**Output Received:**
A clean presentation layer built with Gin, featuring robust HTTP request parsing, boundary validation, and a centralized, predictable error mapping system returning standardized JSON payloads.

## Prompt 3 — Main Entry Point & Dependency Wiring

**Purpose:**
To wire all layers together correctly and configure the server.

**Prompt Used:**
Implement the main.go entry point for the application. Write the logic to load configuration variables via godotenv. Establish a database connection to PostgreSQL using GORM, explicitly setting the logger to the Warn log level. Run auto-migrations for all core models. Manually wire all dependencies by injecting the database instance into the repositories, the repositories into the services, and the services into the HTTP handlers. Register all 8 application routes on the Gin engine router, set trusted proxies to nil to resolve warnings, and gracefully start the HTTP server.

**Output Received:**
The fully functional application entry point, completely wiring the clean architecture components via manual dependency injection, configuring the database, and bootstrapping the Gin server.

## Prompt 4 — Goroutine-Based Concurrency Test

**Purpose:**
To prove with a real test that the SELECT FOR UPDATE implementation prevents overbooking.

**Prompt Used:**
Write a comprehensive Go test in test/concurrency_test.go to validate the concurrent booking implementation. The test must connect directly to the test database, create a single event with exactly available_seats = 1, and create 10 distinct users. Launch 10 goroutines simultaneously to attempt booking the single seat, utilizing sync.WaitGroup for execution synchronization and sync.Mutex to safely collect the results. The test must directly call the registration service and rigorously assert that exactly 1 registration succeeds and exactly 9 fail with the specific ErrNoSeatsAvailable error. Include a clear summary print line at the end of the test.

**Output Received:**
A robust integration test utilizing goroutines to simulate a high-concurrency race condition, definitively proving the pessimistic locking mechanism correctly prevents overbooking.

## Prompt 5 — Bug Fixes & Production Hardening

**Purpose:**
To fix all identified issues after initial implementation and make the project production-grade.

**Prompt Used:**
Audit and apply necessary production hardening fixes across the codebase. Address the GORM default log level issue that prints expected constraint violation errors as panics by explicitly setting the GORM logger to Warn level. Resolve the Gin trusted proxies warning by explicitly executing SetTrustedProxies(nil). Update the User model to ensure standard timestamp and soft delete fields are present. Implement Go-level UUID generation within the BeforeCreate hooks on all three models (User, Event, Registration) to completely remove the database-level dependency on the pgcrypto extension.

**Output Received:**
Code modifications across models, configuration, and the entry point that eliminated runtime warnings, decoupled the application from specific Postgres extensions, and improved overall production readiness.

## Prompt 6 — DESIGN.md Architecture & Race Condition Analysis

**Purpose:**
To document the architectural decisions and provide a deep technical explanation of the concurrency solution.

**Prompt Used:**
Create a comprehensive DESIGN.md document focused on the architectural decisions and race condition analysis. Include a detailed clean architecture diagram. Document the full database schema, providing explanations for the necessity of all constraints. Write an in-depth race condition walkthrough demonstrating exactly what happens in a high-concurrency environment without locking, contrasted against the safe execution flow using SELECT FOR UPDATE. Provide a detailed technical justification for choosing pessimistic locking over optimistic locking, detailing the trade-offs of both approaches, and comprehensively explain why a Go-level sync.Mutex is insufficient for a horizontally scaled distributed system.

**Output Received:**
An advanced technical design document containing architectural visualizations, schema definitions, and a senior-level analysis of distributed locking strategies and race condition mitigation.

---

## Summary

These prompts were used iteratively to build the project in a logical,
layered order — through architecture, implementation,
testing, hardening, and finally documentation. Each prompt built on the output
of the previous one, ensuring a coherent and production-grade final result.


Suitable for academic review and production backend evaluation.

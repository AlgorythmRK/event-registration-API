# AI Prompts Used — Event Registration & Ticketing API

This document contains all prompts used with AI tools (Claude by Anthropic) during 
the development of this project, as required for academic submission transparency.

Each prompt is documented with its purpose, the exact prompt text, and what it produced.

**AI Tool Used:** Claude (Anthropic)  
**Project:** Event Registration & Ticketing API  
**Language:** Go  
**Total Prompts Used:** 13

## Prompt 1 — Initial Project Planning & Requirements Analysis

**Purpose:**
To define the full scope, tech stack, and requirements before writing any code.

**Prompt Used:**
Design a production-ready Go REST API for an event registration and ticketing system. The tech stack must be Go, Gin, PostgreSQL, GORM, and UUIDs for primary keys. The project must strictly follow clean architecture principles with clearly separated layers: Handler, Service, Repository, and Model. Provide a comprehensive list of all required API endpoints, the database schema including all necessary constraints, and detail the strategy for handling concurrent bookings using SELECT FOR UPDATE. List all expected deliverables for this project.

**Output Received:**
A comprehensive project specification outlining the clean architecture layers, a complete list of RESTful endpoints, the initial database schema design, and a high-level explanation of the pessimistic locking strategy for concurrency control.

## Prompt 2 — Database Schema Design

**Purpose:**
To design a normalized PostgreSQL schema with proper constraints before implementation.

**Prompt Used:**
Design a complete, normalized PostgreSQL database schema for the users, events, and registrations tables. Use UUIDs for all primary keys. Implement specific CHECK constraints on the role field (to ensure valid user roles) and status fields. Add a UNIQUE constraint on the email column, and a composite UNIQUE constraint on (event_id, user_id) to prevent duplicate registrations. Define all foreign keys with appropriate cascading actions. Finally, provide a technical explanation for why each constraint is necessary for data integrity.

**Output Received:**
SQL-like schema definitions for the three core tables, complete with UUID configurations, data type specifications, and robust constraints, along with technical justifications for the enforced data integrity rules.

## Prompt 3 — Clean Architecture Folder Structure

**Purpose:**
To establish the project structure before writing any code.

**Prompt Used:**
Define the complete folder and file structure for the Go REST API project strictly adhering to clean architecture principles. Provide a detailed breakdown of the directories including cmd, internal/api/handlers, internal/core/services, internal/core/domain, and internal/repository. Explain exactly what responsibilities belong to each layer, and explicitly define the strict rules restricting what each layer is and is not allowed to import or do.

**Output Received:**
A robust, scalable directory tree structure with comprehensive documentation on the boundaries, dependencies, and responsibilities of the Handler, Service, Repository, and Model layers.

## Prompt 4 — Model Layer Implementation

**Purpose:**
To implement all GORM models with proper tags, hooks, and constraints.

**Prompt Used:**
Implement the three core GORM model files (User, Event, Registration) based on the designed schema. Use UUID primary keys and generate them via BeforeCreate hooks in the Go code, explicitly avoiding reliance on the pgcrypto extension. Include proper GORM tags for column definitions, constraints, and relationships. Implement standard timestamp fields (CreatedAt, UpdatedAt) and soft delete functionality (DeletedAt) on the User and Event models. Define the composite unique index on the registration model, and include the initialization logic for AvailableSeats within the Event's BeforeCreate hook.

**Output Received:**
Production-ready Go struct models representing the database entities, complete with all specified GORM tags, UUID generation logic in BeforeCreate hooks, and robust index definitions.

## Prompt 5 — Repository Layer with Concurrency-Safe Booking

**Purpose:**
To implement the most critical part of the system — the data access layer with SELECT FOR UPDATE locking.

**Prompt Used:**
Implement all three repository interfaces and their concrete implementations, focusing primarily on registration_repository.go. Implement the RegisterForEvent method within a full GORM transaction. Crucially, implement row-level pessimistic locking using SELECT FOR UPDATE when querying the event, and include detailed inline comments explaining the necessity of this lock and the race conditions that would occur without it. Implement the seat availability check, atomic seat decrement using gorm.Expr for safe concurrency, and robust unique violation detection handling both the specific GORM error and SQLSTATE 23505, with a string fallback. Additionally, implement the CancelRegistration method with its own proper transaction and locking mechanism.

**Output Received:**
The complete data access layer implementation, featuring robust transaction management, rock-solid pessimistic locking for concurrent ticket booking, and comprehensive error handling for database constraint violations.

## Prompt 6 — Service Layer with Business Logic & Sentinel Errors

**Purpose:**
To implement business rules and define the error contract between layers.

**Prompt Used:**
Implement the complete service layer along with the domain error contract. Create an errors.go file defining all sentinel errors for the application (e.g., ErrNoSeatsAvailable, ErrAlreadyRegistered, ErrEventNotFound, ErrUserNotFound, ErrNotOrganizer). Create repository_interfaces.go in the core domain to define the interfaces the repository layer must implement, effectively preventing import cycles. Implement the business logic in the services, specifically enforcing organizer role validation during event creation, and performing strict pre-validation checks for user and event existence prior to processing any booking.

**Output Received:**
A highly decoupled service layer containing all core business logic, a comprehensive set of defined sentinel errors for standardized error handling, and correctly abstracted repository interfaces.

## Prompt 7 — Handler Layer with Error Mapping

**Purpose:**
To implement the HTTP presentation layer with proper status code mapping.

**Prompt Used:**
Implement all HTTP handler files strictly adhering to the rule that handlers contain zero business logic. Implement UUID validation at the HTTP boundary before passing parameters to the service layer. Create a centralized respondWithError function that consistently maps all sentinel domain errors to their correct HTTP status codes (e.g., 404 Not Found, 409 Conflict, 403 Forbidden, 500 Internal Server). Ensure all route handlers return clean, standard JSON error responses.

**Output Received:**
A clean presentation layer built with Gin, featuring robust HTTP request parsing, boundary validation, and a centralized, predictable error mapping system returning standardized JSON payloads.

## Prompt 8 — Main Entry Point & Dependency Wiring

**Purpose:**
To wire all layers together correctly and configure the server.

**Prompt Used:**
Implement the main.go entry point for the application. Write the logic to load configuration variables via godotenv. Establish a database connection to PostgreSQL using GORM, explicitly setting the logger to the Warn log level. Run auto-migrations for all core models. Manually wire all dependencies by injecting the database instance into the repositories, the repositories into the services, and the services into the HTTP handlers. Register all 8 application routes on the Gin engine router, set trusted proxies to nil to resolve warnings, and gracefully start the HTTP server.

**Output Received:**
The fully functional application entry point, completely wiring the clean architecture components via manual dependency injection, configuring the database, and bootstrapping the Gin server.

## Prompt 9 — Goroutine-Based Concurrency Test

**Purpose:**
To prove with a real test that the SELECT FOR UPDATE implementation prevents overbooking.

**Prompt Used:**
Write a comprehensive Go test in test/concurrency_test.go to validate the concurrent booking implementation. The test must connect directly to the test database, create a single event with exactly available_seats = 1, and create 10 distinct users. Launch 10 goroutines simultaneously to attempt booking the single seat, utilizing sync.WaitGroup for execution synchronization and sync.Mutex to safely collect the results. The test must directly call the registration service and rigorously assert that exactly 1 registration succeeds and exactly 9 fail with the specific ErrNoSeatsAvailable error. Include a clear summary print line at the end of the test.

**Output Received:**
A robust integration test utilizing goroutines to simulate a high-concurrency race condition, definitively proving the pessimistic locking mechanism correctly prevents overbooking.

## Prompt 10 — Bug Fixes & Production Hardening

**Purpose:**
To fix all identified issues after initial implementation and make the project production-grade.

**Prompt Used:**
Audit and apply necessary production hardening fixes across the codebase. Address the GORM default log level issue that prints expected constraint violation errors as panics by explicitly setting the GORM logger to Warn level. Resolve the Gin trusted proxies warning by explicitly executing SetTrustedProxies(nil). Update the User model to ensure standard timestamp and soft delete fields are present. Implement Go-level UUID generation within the BeforeCreate hooks on all three models (User, Event, Registration) to completely remove the database-level dependency on the pgcrypto extension.

**Output Received:**
Code modifications across models, configuration, and the entry point that eliminated runtime warnings, decoupled the application from specific Postgres extensions, and improved overall production readiness.

## Prompt 11 — Full Project Audit

**Purpose:**
To do a complete end-to-end verification that every file is correct, every interface is satisfied, every import is valid, and the project compiles and runs without issues.

**Prompt Used:**
Perform a full, exhaustive audit of the entire codebase. Verify the correctness of all import paths and ensure zero circular dependencies. Confirm that all repository and service interfaces are fully and correctly satisfied by their implementations. Validate the integrity of the Handler → Service → Repository call chain architecture. Verify the correct implementation of the BeforeCreate hooks on all models. Double-check the SELECT FOR UPDATE implementation and the robustness of the isUniqueViolation detection logic. Validate the assertions in the concurrency test, inspect the go.mod dependencies, and verify the structural integrity of the folder architecture while identifying any potential runtime bugs.

**Output Received:**
Confirmation of architectural compliance, interface satisfaction, import resolution, and overall project stability, ensuring the application compiles and functions precisely as designed.

## Prompt 12 — README.md Documentation

**Purpose:**
To create professional documentation explaining the project, setup, and concurrency strategy.

**Prompt Used:**
Write a complete, professional README.md for the Event Registration & Ticketing API project. Provide a high-level project overview and explicitly detail the utilized tech stack. Include a mermaid or text clean architecture breakdown. Provide exhaustive, step-by-step local setup instructions. Document all available API endpoints, including accurate curl examples and standard sample JSON responses. Finally, write a detailed explanation of the concurrency strategy focusing on the pessimistic locking mechanism, and provide instructions on how to run the integrated concurrency test.

**Output Received:**
A comprehensive, developer-friendly README.md file covering project setup, architectural overviews, API consumption guides, and technical explanations of core concurrency features.

## Prompt 13 — DESIGN.md Architecture & Race Condition Analysis

**Purpose:**
To document the architectural decisions and provide a deep technical explanation of the concurrency solution.

**Prompt Used:**
Create a comprehensive DESIGN.md document focused on the architectural decisions and race condition analysis. Include a detailed clean architecture diagram. Document the full database schema, providing explanations for the necessity of all constraints. Write an in-depth race condition walkthrough demonstrating exactly what happens in a high-concurrency environment without locking, contrasted against the safe execution flow using SELECT FOR UPDATE. Provide a detailed technical justification for choosing pessimistic locking over optimistic locking, detailing the trade-offs of both approaches, and comprehensively explain why a Go-level sync.Mutex is insufficient for a horizontally scaled distributed system.

**Output Received:**
An advanced technical design document containing architectural visualizations, schema definitions, and a senior-level analysis of distributed locking strategies and race condition mitigation.

---

## Summary

These prompts were used iteratively to build the project from scratch in a logical,
layered order — starting from requirements, through architecture, implementation,
testing, hardening, and finally documentation. Each prompt built on the output
of the previous one, ensuring a coherent and production-grade final result.

All generated code was reviewed, understood, and verified to compile and run correctly
before submission.

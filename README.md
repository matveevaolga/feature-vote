# Feature Vote

Backend service for team decision-making: create groups, invite members,
and vote on project features with real-time concurrent voting
processing.

## Features

### Group Management

- Create groups with owner privileges
- Invite users by username (owner only)
- Member roles: owner, admin, member
- Delegate voting creation rights to admins
- Transfer ownership to another member

### Invitation System

- Send pending invitations to users
- Accept/decline invitations
- Automatic group membership upon acceptance
- Track invitation status (pending/accepted/declined)

### Voting System

- Start feature voting with customizable duration
- Concurrent vote processing using goroutines
- Real-time vote collection via channels
- Automatic vote counting with deadline handling
- Results calculation with percentages
- Stop voting prematurely (owner/admin only)
- View results in real-time

### Access Control

- Role-based permissions (owner, admin, member)
- Owner: full control (delete group, manage roles, transfer ownership)
- Admin: can manage members and create votings
- Member: can participate in votings and view results

## Tech Stack

- **Language:** Go 1.24
- **Database:** PostgreSQL 15 with pgx driver
- **Architecture:** Clean Architecture (domain, service, repository, transport)
- **HTTP Router:** chi (lightweight, composable router)
- **Concurrency:** Goroutines, channels, contexts, mutexes, sync.Map
- **Validation:** go-playground/validator
- **Logging:** Structured logging with slog
- **Migrations:** golang-migrate
- **Container:** Docker + docker-compose
- **Configuration:** Environment variables (12-factor app)
- **UUID:** gofrs/uuid

## Prerequisites

- Go 1.24
- Docker and docker-compose
- Make (optional, for Makefile commands)
- golang-migrate (for database migrations)

## Quick Start

### 1. Clone the repository

```bash
git clone https://github.com/matveevaolga/feature-vote.git
cd feature-vote
```

### 2. Set up environment variables

```bash
cp .env.example .env
```

### 3. Start PostgreSQL with Docker

```bash
make docker-up
```

### 4. Run migrations

```bash
make migrate-up
```

### 5. Build and run

```bash
make run
```

The server will start on `http://localhost:8080`

## API Endpoints

### Health (public)
- **GET** `/health` - Full service health check (database connectivity, active votings, uptime)
- **GET** `/readiness` - Service readiness check (database availability)

### Authentication (public)
- **POST** `/auth/register` - Register a new user with email and password
- **POST** `/auth/login` - Login with email and password

### Users (public)
- **POST** `/users` - Legacy user creation (username only, for backward compatibility)

### Groups (authenticated)
- **POST** `/groups` - Create a new group
- **GET** `/groups/{id}` - Get detailed information about a specific group
- **PUT** `/groups/{id}` - Update group details (owner only)
- **DELETE** `/groups/{id}` - Delete a group (owner only)
- **GET** `/users/groups` - List all groups the authenticated user belongs to

### Members (authenticated)
- **POST** `/groups/{id}/invite` - Invite a user to join a group (owner/admin only)
- **GET** `/groups/{id}/members` - List all members of a specific group
- **DELETE** `/groups/{id}/members/{userID}` - Remove a member from a group (owner only)
- **POST** `/groups/{id}/leave` - Leave a group (members except owner)
- **PUT** `/groups/{id}/members/{userID}/role` - Update a member's role within a group (owner only)
- **POST** `/groups/{id}/transfer` - Transfer group ownership to another member (owner only)

### Invitations (authenticated)
- **GET** `/users/invitations` - Get all pending invitations for the authenticated user
- **POST** `/invitations/{id}/accept` - Accept a pending invitation
- **POST** `/invitations/{id}/decline` - Decline a pending invitation

### Votings (authenticated)
- **POST** `/votings` - Create a new voting in a group (owner/admin only)
- **GET** `/votings/{id}` - Get the current status of a voting
- **GET** `/votings/{id}/results` - Get the results of a voting
- **POST** `/votings/{id}/votes` - Cast a vote
- **POST** `/votings/{id}/stop` - Stop an active voting prematurely (owner/admin only)

## Project Structure

The project follows **Clean Architecture** principles and is organized as follows:

### Entry Point
- **`cmd/server/`** — contains `main.go`, the application entry point. All dependencies (config, logger, database connection, repositories, services, handlers) are initialized here, and the HTTP server is started with chi router.

### Internal Logic (`internal/`)
All core application logic resides in the `internal/` package and is not accessible for external imports:

- **`internal/domain/`** — business entities (User, Group, Voting, etc.) and repository interfaces that define contracts for data access.
- **`internal/repository/`** — PostgreSQL implementations of the repository interfaces. Each entity has its own repository (userRepository, groupRepository, votingRepository).
- **`internal/service/`** — business logic layer. Coordinates repositories, implements voting rules, access control, and concurrent vote processing (goroutines, channels, contexts).
- **`internal/transport/`** — delivery layer:
  - **`handler/`** — HTTP handlers that receive requests, validate them, call service methods, and format responses. Path parameters are extracted using `chi.URLParam()`.
    - **`dto/`** — request and response structures (Data Transfer Objects).
  - **`middleware/`** — HTTP middleware components (logging, authentication).
- **`internal/config/`** — configuration loading and validation from environment variables.
- **`internal/logger/`** — structured logging setup (slog).

### Migrations and Configuration
- **`migrations/`** — SQL migration files for database schema (table creation, indexes, etc.).

### Root Files
- **`docker-compose.yml`** — service definitions for local development (PostgreSQL).
- **`Dockerfile`** — instructions for building the application Docker image.
- **`Makefile`** — command automation (run, test, migrations, Docker).
- **`.env.example`** — example environment variables file.
- **`README.md`** — project documentation.

## Authentication

For development/demo purposes, authentication is simplified:
- Include `X-User-ID` header with valid user UUID
- Public endpoint `POST /users` doesn't require authentication
- All other endpoints require valid `X-User-ID`

Example:

```bash
curl -X POST http://localhost:8080/groups \
  -H "X-User-ID: 123e4567-e89b-12d3-a456-426614174000" \
  -H "Content-Type: application/json" \
  -d '{"name":"Gophers"}'
```

## Makefile Commands

- **`make run`** - Run the server
- **`make build`** - Build the binary
- **`make test`** - Run tests with race detection
- **`make docker-up`** - Start PostgreSQL in Docker
- **`make docker-down`** - Stop PostgreSQL
- **`make docker-logs`** - View Docker logs
- **`make migrate-up`** - Apply database migrations
- **`make migrate-down`** - Rollback migrations
- **`make migrate-create`** - Create a new migration

## Database Schema

### Main Tables

- **users** - application users
- **groups** - groups with owner reference
- **group_members** - many-to-many with roles
- **invitations** - pending invitations
- **votings** - feature votings with status
- **votes** - user votes with yes/no
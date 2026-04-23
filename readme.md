# Taskforge

Task management REST API built with Go.

## About

Taskforge is an API that allows users to create teams, manage members, and organize tasks with permission controls. Users register, create teams, invite members by email, and manage tasks with assignment and status tracking.

## Stack

- **Go 1.22+** — main language
- **Chi** — HTTP router and middlewares
- **PostgreSQL** — relational database
- **pgxpool** — Postgres connection pooling
- **SQLc** — type-safe Go code generation from SQL queries
- **Tern** — database migrations
- **Redis** — caching and rate limiting
- **JWT** — stateless authentication
- **bcrypt** — password hashing
- **Docker Compose** — local environment (Postgres + Redis)
- **slog** — structured logging

## Features

**Authentication**
- User registration with password hashing (bcrypt)
- Login with JWT token generation
- Authentication middleware for protected routes

**Teams**
- Create team (creator becomes owner and member automatically)
- List teams for the logged-in user
- Get team by ID
- Add member by email (owner only)
- Remove member (owner only, cannot remove self)
- List team members (members only)

**Tasks**
- Create task in a team (members only)
- List team tasks (members only)
- Get task by ID
- Update status (`pending`, `in_progress`, `done`)
- Assign task to a team member
- Delete task

**Infrastructure**
- Redis cache for task and member listings with automatic invalidation on writes
- IP-based rate limiting using Redis (100 requests/minute)
- Request ID, recovery, and logging via Chi middlewares

## Project Structure

```
taskforge/
├── cmd/api/              # Application entrypoint
├── internal/
│   ├── api/              # HTTP response helpers (SendJSON)
│   ├── auth/             # JWT generation and validation
│   ├── handler/          # HTTP handlers (users, teams, tasks)
│   └── middleware/        # Auth middleware and rate limiter
├── db/
│   ├── migrations/       # SQL migrations (Tern)
│   ├── queries/          # SQL queries for SQLc
│   └── sqlc/             # SQLc generated code
├── docker-compose.yml
├── .env.example
└── go.mod
```

## Routes

### Public
| Method | Route | Description |
|--------|-------|-------------|
| GET | `/status` | Health check |
| POST | `/register` | User registration |
| POST | `/login` | Login (returns JWT) |

### Protected (requires Bearer token)
| Method | Route | Description |
|--------|-------|-------------|
| POST | `/teams/` | Create team |
| GET | `/teams/` | List user's teams |
| GET | `/teams/{id}` | Get team by ID |
| POST | `/teams/{id}/members` | Add member |
| DELETE | `/teams/{id}/members` | Remove member |
| GET | `/teams/{id}/members` | List members |
| POST | `/teams/{id}/tasks` | Create task |
| GET | `/teams/{id}/tasks` | List tasks |
| GET | `/teams/{id}/tasks/{taskID}` | Get task |
| PATCH | `/teams/{id}/tasks/{taskID}/status` | Update status |
| PATCH | `/teams/{id}/tasks/{taskID}/assign` | Assign task |
| DELETE | `/teams/{id}/tasks/{taskID}` | Delete task |

## Getting Started

**Prerequisites:** Go 1.22+, Docker

1. Clone the repository
```bash
git clone https://github.com/jmarcosnsf/taskforge.git
cd taskforge
```

2. Copy `.env.example` and fill in your values
```bash
cp .env.example .env
```

3. Start Postgres and Redis
```bash
docker compose up -d
```

4. Run migrations
```bash
cd db/migrations
tern migrate
cd ../..
```

5. Start the server
```bash
go run cmd/api/main.go
```

Server runs at `http://localhost:8080`.

## Data Model

```
users ──────┐
            ├── team_members (N:N)
teams ──────┘
  │
  └── tasks (1:N, with optional user_id for assignment)
```

- **users**: id, name, email, password, created_at, updated_at
- **teams**: id, name, description, owner_id (FK users), created_at, updated_at
- **team_members**: user_id (FK users), team_id (FK teams) — composite PK
- **tasks**: id, title, description, status, team_id (FK teams), user_id (FK users, nullable), created_at, updated_at
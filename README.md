# Todo Backend Service

This is a backend service for a Flutter to-do app that accepts text input, extracts tasks using an LLM, and stores them in a PostgreSQL database. It's built with Go using the Gin framework.

## Features

- User Authentication (JWT)
- Task CRUD operations
- LLM-powered task extraction from text
- PostgreSQL database
- Dockerized deployment
- CORS middleware
- Centralized error handling and logging

## Technology Stack

- **Backend Language:** Go (Golang)
- **Web Framework:** Gin
- **Database:** PostgreSQL (via GORM)
- **LLM Integration:** OpenAI (GPT models)
- **Containerization:** Docker, Docker Compose
- **Migrations:** golang-migrate
- **Logging:** zerolog

## Project Structure

```
/todo-backend
  /cmd/server           # Main application entry point
  /internal/api         # API handlers, routes, and middleware
  /internal/services    # Business logic layer
  /internal/repositories # Database access layer
  /internal/models      # Data structures/models
  /internal/config      # Configuration loading
  /internal/llm         # LLM (Large Language Model) integration for task extraction
  /internal/middleware  # Custom Gin middlewares (logging, recovery)
  /migrations           # SQL migration files for PostgreSQL
  Dockerfile            # Dockerfile for building the Go application
  docker-compose.yml    # Docker Compose for multi-service setup (backend, db)
  go.mod                # Go module definition
  go.sum                # Go module checksums
  README.md             # Project README (this file)
  .env.example          # Environment variable template
```

## Setup and Running Locally

### Prerequisites

- Go (1.21 or higher)
- Docker & Docker Compose
- `golang-migrate` CLI (for manual migrations, if not using `make` commands)
  ```bash
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```

### 1. Clone the Repository

```bash
git clone https://github.com/your-username/todo-backend.git
cd todo-backend
```

### 2. Configure Environment Variables

Create a `.env` file by copying `.env.example` and fill in the necessary values.

```bash
cp .env.example .env
```

Edit the `.env` file:

```ini
# Server Configuration
PORT=8080

# Database Configuration
DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=todo
DB_SSLMODE=disable

# JWT Configuration
JWT_SECRET=your-32-char-secret-key-for-jwt-signing # IMPORTANT: Change this to a strong, random key!

# OpenAI API Key
OPENAI_API_KEY=sk-your-openai-api-key # Get from OpenAI platform
```

**Note:** For `JWT_SECRET`, generate a strong random string (e.g., `openssl rand -base64 32`).

### 3. Start Docker Services (Database)

Start the PostgreSQL database using Docker Compose:

```bash
docker-compose up -d db
```

Wait a few seconds for the database to initialize.

### 4. Run Database Migrations

Apply the database schema migrations. Ensure you have `golang-migrate` installed (see Prerequisites).

```bash
# Wait for the DB to be ready if you just started it.
# You might need to adjust the DB_HOST to 'localhost' if running migrate from outside docker-compose network
migrate -path migrations -database "postgres://postgres:postgres@localhost:5433/todo?sslmode=disable" up
```
**Note:** If `localhost:5433` doesn't work, ensure your `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, and `DB_NAME` in the `.env` file match the connection string.

### 5. Build and Run the Backend

#### Using Docker Compose (Recommended)

```bash
docker-compose up --build backend
```

The backend service will be accessible at `http://localhost:8080`.

#### Running Directly (for Development)

```bash
go run cmd/server/main.go
```

The backend service will be accessible at `http://localhost:8080`.

## API Endpoints

All endpoints respond with JSON. Errors are returned in the format: `{"error": "message"}`.

### Authentication

- `POST /auth/register`
  - **Request:**
    ```json
    {
      "email": "user@example.com",
      "password": "strongpassword"
    }
    ```
  - **Response (201 Created):**
    ```json
    {
      "id": "a-uuid-string",
      "email": "user@example.com",
      "created_at": "2023-10-26T10:00:00Z"
    }
    ```
- `POST /auth/login`
  - **Request:**
    ```json
    {
      "email": "user@example.com",
      "password": "strongpassword"
    }
    ```
  - **Response (200 OK):**
    ```json
    {
      "token": "your.jwt.token"
    }
    ```
- `GET /auth/me`
  - **Headers:** `Authorization: Bearer <your.jwt.token>`
  - **Response (200 OK):**
    ```json
    {
      "id": "a-uuid-string",
      "email": "user@example.com",
      "created_at": "2023-10-26T10:00:00Z"
    }
    ```

### Tasks

All task endpoints require JWT authentication. Include `Authorization: Bearer <your.jwt.token>` in the request headers.

- `POST /tasks/from-text`
  - Extracts tasks from a given text using an LLM and creates them.
  - **Request:**
    ```json
    {
      "text": "Tomorrow buy groceries, call mom, and schedule dentist next week"
    }
    ```
  - **Response (201 Created):** Array of created tasks
    ```json
    [
      {
        "id": "a-uuid",
        "user_id": "user-uuid",
        "title": "Buy groceries",
        "description": "Buy groceries tomorrow",
        "due_date": "2025-11-20T00:00:00Z",
        "priority": "medium",
        "raw_text": "Tomorrow buy groceries, call mom, and schedule dentist next week",
        "created_at": "2023-10-26T10:00:00Z"
      },
      {
        "id": "another-uuid",
        "user_id": "user-uuid",
        "title": "Call mom",
        "description": "Call mom tomorrow",
        "due_date": "2025-11-20T00:00:00Z",
        "priority": "medium",
        "raw_text": "Tomorrow buy groceries, call mom, and schedule dentist next week",
        "created_at": "2023-10-26T10:00:00Z"
      }
    ]
    ```
- `GET /tasks`
  - Returns all tasks for the authenticated user.
  - **Response (200 OK):** Array of tasks
- `POST /tasks`
  - Creates a new task manually.
  - **Request:**
    ```json
    {
      "title": "New Task Title",
      "description": "Optional description",
      "due_date": "2025-12-01T10:30:00Z",
      "priority": "high"
    }
    ```
  - **Response (201 Created):** The created task object.
- `GET /tasks/:id`
  - Returns a specific task by ID.
  - **Response (200 OK):** The task object.
- `PUT /tasks/:id`
  - Updates an existing task.
  - **Request:** (partial update allowed)
    ```json
    {
      "title": "Updated Task Title",
      "priority": "low"
    }
    ```
  - **Response (200 OK):** The updated task object.
- `DELETE /tasks/:id`
  - Deletes a task by ID.
  - **Response (204 No Content)**

## Example cURL Commands

First, register a user and get an authentication token:

```bash
# 1. Register
curl -X POST -H "Content-Type: application/json" -d '{"email": "test@example.com", "password": "password123"}' http://localhost:8080/auth/register

# 2. Login to get token
AUTH_TOKEN=$(curl -X POST -H "Content-Type: application/json" -d '{"email": "test@example.com", "password": "password123"}' http://localhost:8080/auth/login | jq -r .token)
echo "Auth Token: $AUTH_TOKEN"

# 3. Get user info (requires token)
curl -X GET -H "Authorization: Bearer $AUTH_TOKEN" http://localhost:8080/auth/me

# 4. Extract and create tasks from text
curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $AUTH_TOKEN" -d '{"text": "Buy milk and eggs tomorrow, call mom on Friday"}' http://localhost:8080/tasks/from-text

# 5. Get all tasks
curl -X GET -H "Authorization: Bearer $AUTH_TOKEN" http://localhost:8080/tasks

# 6. Create a task manually
curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $AUTH_TOKEN" -d '{"title": "Read Go book", "priority": "high", "due_date": "2025-11-25T18:00:00Z"}' http://localhost:8080/tasks

# (Assuming TASK_ID is obtained from a previous response)
# TASK_ID="a-task-id-from-response"

# 7. Get a specific task
# curl -X GET -H "Authorization: Bearer $AUTH_TOKEN" http://localhost:8080/tasks/$TASK_ID

# 8. Update a task
# curl -X PUT -H "Content-Type: application/json" -H "Authorization: Bearer $AUTH_TOKEN" -d '{"description": "Finish chapter 5", "priority": "medium"}' http://localhost:8080/tasks/$TASK_ID

# 9. Delete a task
# curl -X DELETE -H "Authorization: Bearer $AUTH_TOKEN" http://localhost:8080/tasks/$TASK_ID
```

## Architecture Diagram

```mermaid
graph TD
    A[Flutter Client] -- Text Input --> B(Backend Service)
    B -- Authenticate --> C(Auth Middleware)
    C -- Valid JWT --> D(API Handlers)
    D -- Extract Tasks --> E(LLM Extractor - OpenAI)
    D -- CRUD Operations --> F(Task Service)
    F -- Data Access --> G(Task Repository)
    G -- Store/Retrieve --> H(PostgreSQL Database)
    E -- Structured Tasks --> F
    F -- Store Tasks --> G
    G -- User Data --> H
    D -- User Data --> H
    C -- User Context --> D
```

## Postman/Thunder Client Collection

(To be generated as a separate JSON file)

## Error Handling

- All API errors return a JSON object `{"error": "message"}`.
- Global panic recovery middleware is active.
- Structured logging with `zerolog` is implemented.

## Testing

- Unit tests for LLM Extractor, Auth Service, and Task Service.
- Integration tests for Auth and Task API routes.

To run all tests:

```bash
go test -v ./...
```

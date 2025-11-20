You are an expert backend engineer. Build a complete backend for a Flutter to-do app that accepts audio input, converts it to text, extracts tasks, and stores them.
Use Go (Golang) with the Gin framework, Postgres as DB, S3-compatible storage for audio, and an LLM API for task extraction.
Produce production-ready clean code, a full project structure, and step-by-step setup instructions.

---

1. PROJECT GOAL

---

Create a backend service with the following responsibilities:

1. Accept audio uploads from the Flutter frontend.
2. Convert audio → text using Whisper (OpenAI API or local Whisper.cpp).
3. Extract structured tasks from the text using an LLM.
4. Store tasks and metadata in Postgres.
5. Return extracted tasks to the Flutter client.
6. Provide CRUD APIs for tasks.
7. Provide authentication (JWT).
8. Provide clean Docker deployment.

---

2. DELIVERABLES

---

Produce ALL of the following:

A. Complete backend folder structure:

```
/todo-backend
  /cmd/server
  /internal/api
  /internal/services
  /internal/repositories
  /internal/models
  /internal/config
  /internal/storage
  /internal/llm
  /internal/stt
  /migrations
  Dockerfile
  docker-compose.yml
  go.mod
  go.sum
  README.md
```

B. Fully working code for every file.
C. Docker & docker-compose files.
D. Postgres schema + migrations.
E. `.env.example` template.
F. A complete README with installation & run instructions.

---

3. TECHNOLOGY REQUIREMENTS

---

Backend Language: Go (Golang)

* Use Gin for routing
* Use GORM or sqlc
* Add CORS middleware
* Unit and integration tests

Database: Postgres

* docker-compose integration
* migrations with golang-migrate

Storage: S3-compatible (AWS or DigitalOcean Spaces)

Speech-to-Text:

* Create module `internal/stt/whisper.go`
* Support:

  * Mode A: OpenAI Whisper API
  * Mode B: local Whisper.cpp

LLM Task Extraction:

* Create module `internal/llm/extractor.go`
* Strict JSON schema enforcement
* Multi-task extraction
* Natural date parsing (tomorrow, next week, Monday morning)

---

4. API ENDPOINT REQUIREMENTS

---

AUTHENTICATION

* POST /auth/register
* POST /auth/login
* GET /auth/me
  Use JWT HS256 and bcrypt.

AUDIO UPLOAD
POST /audio

* Accept multipart/form-data
* Validate audio file
* Save to S3
* Create DB record
* Trigger STT
* Return:

```
{
  "audio_id": "a123",
  "status": "processing"
}
```

GET /audio/:id
Returns:

```
{
  "id": "a123",
  "status": "completed",
  "transcript": "...",
  "tasks": [...]
}
```

TEXT → TASK EXTRACTION
POST /tasks/from-text
Input:

```
{ "text": "Tomorrow buy milk and call Raj" }
```

Output (array of tasks):

```
[
  {
    "title": "Buy milk",
    "description": "Buy milk tomorrow",
    "due_date": "2025-11-23T08:00:00Z",
    "priority": "medium",
    "subtasks": []
  },
  {
    "title": "Call Raj",
    "description": "Call Raj tomorrow",
    "due_date": "2025-11-23T10:00:00Z",
    "priority": "low",
    "subtasks": []
  }
]
```

TASK CRUD

* GET /tasks
* POST /tasks
* PUT /tasks/:id
* DELETE /tasks/:id

---

5. MODELS AND DB SCHEMA

---

USER

* id UUID
* email
* password_hash
* created_at

TASK

* id UUID
* user_id FK
* title
* description
* due_date
* priority
* raw_text
* created_at

AUDIO_UPLOAD

* id UUID
* user_id FK
* s3_url
* transcript
* status (‘pending’, ‘processing’, ‘completed’, ‘failed’)
* created_at

---

6. TASK EXTRACTION RULES

---

Extractor must:

* Detect multiple tasks
* Handle complex date expressions
* Use Go `time` + NLP date parser
* Always return array
* Never return invalid JSON
* On failure → return empty array

LLM SYSTEM PROMPT MUST INCLUDE:

* Strict JSON schema
* No prose, JSON only
* Multi-task extraction
* Required fields:

  * title
  * description
  * due_date (ISO8601 or null)
  * priority (low|medium|high)
  * subtasks array

---

7. ERROR HANDLING

---

* Always return JSON:

```
{ "error": "message" }
```

* Panic recovery middleware
* Logging with zerolog or slog

---

8. TESTING

---

Provide:

* unit tests for extractor
* unit tests for audio upload
* integration tests for main API routes

---

9. FINAL OUTPUT REQUIREMENTS

---

The final answer MUST include:

✔ Full folder structure
✔ All Go code
✔ All migrations
✔ Dockerfile
✔ docker-compose.yml
✔ README with setup instructions
✔ Example cURL commands
✔ Mermaid/ASCII architecture diagram
✔ Postman/Thunder collection JSON

---

END OF PROMPT
Build the entire backend exactly as described above.

---


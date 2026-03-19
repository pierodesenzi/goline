# GoLine
### Minimal Task Queue API (Gin + Redis)
---
GoLine is a minimal HTTP API written in Go using Gin, backed by Redis.
It demonstrates how to structure a small but production-shaped service with:

* clean routing
* handler/service separation
* JSON validation
* Redis-backed queues

This is intentionally simple, but the structure scales.

---

## Features

* Healthcheck endpoint
* Create and enqueue tasks
* Redis-backed queue using lists
* Clean project structure (cmd + internal)
* Middleware for logging and error handling

---

## Project Structure

```
/cmd/api/main.go        # entrypoint (wiring, dependencies)
/internal/http/         # routing + middleware
/internal/tasks/        # domain logic (handler + service)
```

Separation of concerns:

* **main.go** → app wiring
* **http/** → transport layer (routes, middleware)
* **tasks/handler** → HTTP ↔ domain translation
* **tasks/service** → business logic (Redis interaction)

---

## Requirements

* Go 1.20+
* Redis running locally (default: `localhost:6379`)

---

## Setup

Clone the repo and initialize dependencies:

```bash
go mod tidy
```

Run Redis (if not already running):

```bash
redis-server
```

Start the API:

```bash
go run cmd/api/main.go
```

Server runs on:

```
http://localhost:8080
```

---

## Endpoints

### Healthcheck

```
GET /api/health
```

Response:

```json
{ "status": "ok" }
```

---

### Create Queue

```
POST /api/queue
```

Body:

```json
{
  "name": "queue1"
}
```

Response:

```json
{
  "queue": "queue1",
  "status": "created"
}
```

---

### Create Task on Queue

```
POST /api/queue/task
```

Body:

```json
{
  "queue": "queue1",
  "task": "task1"
}
```

Response:

```json
{
  "queue": "queue1",
  "status": "enqueued"
}
```
* Note: There is **no implicit "create queue" step**.

  Queues are created only using the POST /api/queue endpoint. Creating a tasks for an non existent queue will return an error.

---

## Redis Data Model

Queues are implemented using Redis **lists**.

* Key → queue name
* Value → list of tasks

Example:

```bash
RPUSH queue1 "task1"
RPUSH queue1 "task2"
```

Inspect:

```bash
LRANGE queue1 0 -1
```

---

## Design Notes

This project follows a simple layered design:

```
HTTP (Gin)
   ↓
Handler (request/response)
   ↓
Service (business logic)
   ↓
Redis (storage)
```

### Why this matters

* Handlers stay thin and testable
* Business logic is isolated from HTTP
* Easy to extend with new endpoints

---

## Next Steps

* Add worker/consumer (BLPOP / BRPOP)
* Create GET endpoint to inspect list
* Create configuration file
* Add request IDs + structured logging
* Replace `map[string]interface{}` with typed structs
* Add persistence or retry policies
* Add OpenAPI/Swagger

---

## FAQ

### Why no database?

Redis is used directly as a queue for simplicity.

### Why separate handler and service?

To avoid mixing HTTP concerns with business logic and to make the code easier to scale and test.

---

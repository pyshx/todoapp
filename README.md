# Multi-Tenant Todo API

A production-style multi-tenant task management API I built to explore clean architecture patterns, distributed systems concepts, and Go's type system.

## What This Is

This is a gRPC/HTTP API for managing tasks across multiple companies (tenants). Users can create tasks, assign them to teammates, and control who can see what—all while ensuring data isolation between companies.

## Why I Built This

I wanted to understand how to:
- **Design for multi-tenancy** without over-engineering (when is row-level isolation enough vs. schema-per-tenant?)
- **Prevent race conditions** in collaborative environments (optimistic locking seemed elegant)
- **Handle idempotency** properly (what happens when a client retries a create request?)
- **Build type-safe domain models** in Go (can the type system prevent invalid state?)
- **Structure gRPC services** with clean separation of concerns

## Interesting Problems I Solved

### 1. **Task Visibility Logic**
Not all tasks should be visible to everyone in a company. I implemented a visibility system where:
- Tasks can be `company_wide` (everyone sees them)
- Or `only_me` (only creator and assignee see them)

This required careful query filtering at the repository layer to prevent data leaks:
```go
task.company_id == user.company_id
AND (
    task.visibility == "company_wide"
    OR task.creator_id == user.id
    OR task.assignee_id == user.id
)
```

See: `pkg/task/task.go:124` for the visibility check logic.

### 2. **Optimistic Locking for Concurrent Updates**
When two users try to update the same task simultaneously, one should win cleanly. I added a `version` field that increments on each update:

```sql
UPDATE tasks
SET title = $1, version = version + 1
WHERE id = $2 AND version = $3
```

If the version doesn't match, the update fails with a conflict error. The client gets fresh data and retries. Simple, no distributed locks needed.

See: `internal/infra/postgres/task_repo.go:187`

### 3. **Idempotency Keys**
Network requests can fail and retry. Without idempotency, a client might create the same task twice. I built an in-memory store (TTL-based cleanup) that caches request results by key:

- First request: Process it, cache the result
- Retry with same key: Return cached result immediately

For production with multiple instances, I'd swap this for Redis.

See: `pkg/idempotency/` and `internal/infra/grpc/interceptors.go:154`

### 4. **Cursor-Based Pagination**
Offset pagination (`LIMIT 10 OFFSET 20`) breaks when data changes between requests—you skip or duplicate items. I used cursor pagination with `(created_at, id)` tuples:

```sql
WHERE (created_at, id) > ($cursor_time, $cursor_id)
ORDER BY created_at, id
LIMIT 20
```

Stable, consistent results even as tasks are added/deleted.

See: `internal/infra/postgres/task_repo.go:88`

### 5. **Clean Error Handling**
Instead of returning raw database errors to clients, I created typed errors (`ErrNotFound`, `ErrPermissionDenied`) that carry context (resource type, ID) and map cleanly to gRPC status codes.

See: `pkg/apperr/` and `internal/infra/grpc/errors.go`

## Architecture Decisions

### Clean Architecture with Domain-Driven Design
```
┌─────────────────────────────────────────┐
│  Transport (gRPC handlers)              │
├─────────────────────────────────────────┤
│  Use Cases (application logic)          │
├─────────────────────────────────────────┤
│  Domain (entities, business rules)      │  ← No framework dependencies
├─────────────────────────────────────────┤
│  Infrastructure (Postgres, config)      │
└─────────────────────────────────────────┘
```

**Why this structure?**
- Domain layer (`pkg/`) has zero external dependencies—it's pure business logic
- Easy to test without spinning up a database
- Could swap gRPC for GraphQL without touching domain code

### Why Connect Instead of Raw gRPC?
[Connect](https://connectrpc.com) gives me dual-protocol support: same proto definitions, works over gRPC *and* HTTP/JSON. This means I can test with `curl` instead of writing gRPC clients:

```bash
curl -X POST http://localhost:50051/todo.v1.TodoService/CreateTask \
  -H "Content-Type: application/json" \
  -d '{"title": "New task"}'
```

### Type-Safe UUIDs
Instead of `string` everywhere, I wrapped UUIDs in types:
```go
type UserID struct { uuid.UUID }
type TaskID struct { uuid.UUID }
type CompanyID struct { uuid.UUID }
```

This prevents accidentally passing a `UserID` where a `TaskID` is expected—caught at compile time.

See: `pkg/id/`

## Project Structure

```
pkg/                    # Domain layer (business logic)
  ├── task/            # Task entity, visibility rules
  ├── user/            # User entity, role-based auth
  ├── company/         # Company entity
  ├── auth/            # JWT signing/validation
  └── idempotency/     # Request deduplication

internal/
  ├── usecase/         # Application layer (one file per use case)
  ├── infra/
  │   ├── grpc/        # Transport layer (handlers, interceptors)
  │   └── postgres/    # Repository implementations
  └── di/              # Dependency injection wiring

proto/                 # Protocol buffer definitions
migrations/            # SQL schema
e2e/                   # End-to-end tests
```

## Tech Stack

- **Go 1.22+** - Love the simplicity and explicit error handling
- **gRPC/Connect** - Dual protocol support (gRPC + HTTP/JSON)
- **PostgreSQL** - Relational model fits task management well
- **Protobuf** - Type-safe API contracts
- **JWT (HS256)** - Simple auth with x-user-id fallback for dev

## Running It

```bash
# Start PostgreSQL
docker-compose up -d

# Run migrations
make migrate

# Start server
make run

# Run tests
make test
```

### Try It Out

```bash
# List tasks (as Alice)
curl -X POST http://localhost:50051/todo.v1.TodoService/ListCompanyTasks \
  -H "Content-Type: application/json" \
  -H "x-user-id: aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" \
  -d '{}'

# Create a task with idempotency
curl -X POST http://localhost:50051/todo.v1.TodoService/CreateTask \
  -H "Content-Type: application/json" \
  -H "x-user-id: aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" \
  -H "Idempotency-Key: my-unique-key-123" \
  -d '{"title": "Review PR", "visibility": "VISIBILITY_COMPANY_WIDE"}'
```

**Seed data:**
- Companies: Acme Corp (`11111111-...`), Beta Inc (`22222222-...`)
- Users: alice@acme.com (editor), bob@acme.com (viewer), charlie@beta.com (editor)

## What I'd Add Next

1. **Distributed tracing** (OpenTelemetry) - Visualize request flows across services
2. **Task dependencies** - Block task B until task A is done (DAG validation to prevent cycles)
3. **Recurring tasks** - Cron-like scheduling (interesting to model in the domain)
4. **Soft deletes with audit log** - Compliance requirements
5. **Rate limiting** - Per-user or per-company quotas
6. **Redis for idempotency** - Scale beyond single instance

## Trade-Offs I Made

### Row-Level Multi-Tenancy
I use `company_id` on every table and filter every query. This is simple and works well for <1000 tenants. For enterprise scale, I'd consider schema-per-tenant or even separate databases.

### In-Memory Idempotency Store
Fast and simple for a single instance. For horizontal scaling, this needs to be Redis with shared state.

### Optimistic Locking
Clients must handle version conflicts by retrying. This trades occasional retries for avoiding distributed locks—worth it for the simplicity.

### JWT with x-user-id Fallback
The JWT implementation is production-ready (HMAC-SHA256, configurable expiry), but I kept the `x-user-id` header fallback for easy local testing. In production, I'd remove the fallback and add token refresh, JWKS rotation, etc.

## What I Learned

- **Clean architecture isn't over-engineering** when you need testability and flexibility
- **Optimistic locking is simpler than pessimistic** for most CRUD operations
- **Cursor pagination is worth the complexity** when data consistency matters
- **Typed errors** make debugging much easier than generic error strings
- **gRPC interceptors** are powerful for cross-cutting concerns (logging, auth, metrics)
- **Domain-driven design** forces you to think about business rules upfront

## Testing

```bash
# Unit tests (fast, no database)
go test ./pkg/... ./internal/usecase/...

# Integration tests (requires Postgres)
DATABASE_URL="postgres://..." go test ./internal/infra/postgres/...

# E2E tests (requires running server)
E2E_SERVER_URL="http://localhost:50051" go test ./e2e/...
```

## API Reference

| Method | Description | Auth Required |
|--------|-------------|---------------|
| `CreateTask` | Create a new task | Editor role |
| `ListCompanyTasks` | List all visible tasks in company | Any |
| `ListMyTasks` | List tasks assigned to me | Any |
| `GetTask` | Get task by ID (if visible) | Any |
| `UpdateTask` | Update task (with version check) | Editor role |
| `DeleteTask` | Delete task | Editor role |

**Visibility Rules:**
- `VISIBILITY_ONLY_ME`: Only creator and assignee can see it
- `VISIBILITY_COMPANY_WIDE`: All users in the company can see it

**Authorization:**
- `editor` role: Can create, update, delete tasks
- `viewer` role: Can only read tasks (respecting visibility)

---

Built with curiosity about distributed systems and clean architecture patterns.

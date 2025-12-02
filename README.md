# ToDo Service

Multi-tenant ToDo management API built with Go and gRPC/Connect.

## Quick Start

```bash
# Start PostgreSQL
make db-up

# Run the server
make run

# Or run everything with Docker
make docker
```

## Authentication

The API uses JWT authentication. Include the token in the Authorization header:

```bash
Authorization: Bearer <jwt-token>
```

For development/testing, you can use the x-user-id header as a fallback:

```bash
x-user-id: <user-uuid>
```

Environment variables for JWT:
- `JWT_SECRET`: Secret key for token signing (required in production)
- `JWT_DURATION`: Token validity duration (default: 24h)

## Testing

```bash
# Run unit tests
make test

# Run integration tests (requires PostgreSQL)
DATABASE_URL="postgres://todo:todo@localhost:5433/todo?sslmode=disable" go test ./internal/infra/postgres/... -v

# Run E2E tests (requires server running)
E2E_SERVER_URL="http://localhost:50051" go test ./e2e/... -v
```

### Manual API Testing

```bash
# List tasks (using x-user-id for testing)
curl -X POST http://localhost:50051/todo.v1.TodoService/ListCompanyTasks \
  -H "Content-Type: application/json" \
  -H "x-user-id: aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" \
  -d '{}'

# Create task with idempotency key
curl -X POST http://localhost:50051/todo.v1.TodoService/CreateTask \
  -H "Content-Type: application/json" \
  -H "x-user-id: aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" \
  -H "Idempotency-Key: unique-request-id-123" \
  -d '{"title": "New task", "visibility": "VISIBILITY_COMPANY_WIDE"}'
```

## Idempotency

Mutation endpoints (CreateTask, UpdateTask, DeleteTask) support idempotency keys to prevent duplicate operations:

```bash
Idempotency-Key: <unique-request-id>
```

If the same request is sent with the same idempotency key, the server returns an `ALREADY_EXISTS` error instead of creating duplicates.

## Seed Data

**Companies:**
- Acme Corp: `11111111-1111-1111-1111-111111111111`
- Beta Inc: `22222222-2222-2222-2222-222222222222`

**Users:**
- alice@acme.com (editor): `aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa`
- bob@acme.com (viewer): `bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb`
- charlie@beta.com (editor): `cccccccc-cccc-cccc-cccc-cccccccccccc`

## API Reference

| Method | Description | Auth |
|--------|-------------|------|
| CreateTask | Create a new task | Editor |
| ListCompanyTasks | List visible tasks in company | Any |
| ListMyTasks | List tasks assigned to user | Any |
| GetTask | Get task by ID | Any (visibility) |
| UpdateTask | Update task (optimistic locking) | Editor |
| DeleteTask | Delete task | Editor |

## Architecture

```
cmd/server/          # Entry point
pkg/                 # Domain layer (entities, value objects, repo interfaces)
  ├── task/
  ├── user/
  ├── company/
  ├── auth/          # JWT authentication
  ├── idempotency/   # Request deduplication
  └── id/
internal/
  ├── usecase/       # Application layer (use cases)
  ├── infra/         # Infrastructure (PostgreSQL, gRPC)
  └── config/
proto/               # Protocol buffer definitions
migrations/          # Database schema
e2e/                 # End-to-end tests
```

## Tech Stack

- **Language:** Go 1.22+
- **Protocol:** gRPC with Connect (HTTP/JSON fallback)
- **Database:** PostgreSQL
- **Auth:** JWT (HS256)
- **Logging:** slog (structured JSON)

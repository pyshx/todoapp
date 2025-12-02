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

## Testing

```bash
# Run unit tests
make test

# Test API (requires server running)
# List tasks as alice (editor)
curl -X POST http://localhost:50051/todo.v1.TodoService/ListCompanyTasks \
  -H "Content-Type: application/json" \
  -H "x-user-id: aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" \
  -d '{}'

# Create task
curl -X POST http://localhost:50051/todo.v1.TodoService/CreateTask \
  -H "Content-Type: application/json" \
  -H "x-user-id: aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" \
  -d '{"title": "New task", "visibility": "VISIBILITY_COMPANY_WIDE"}'
```

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
  └── id/
internal/
  ├── usecase/       # Application layer (use cases)
  ├── infra/         # Infrastructure (PostgreSQL, gRPC)
  └── config/
proto/               # Protocol buffer definitions
migrations/          # Database schema
```

## Tech Stack

- **Language:** Go 1.22+
- **Protocol:** gRPC with Connect (HTTP/JSON fallback)
- **Database:** PostgreSQL
- **Logging:** slog (structured JSON)

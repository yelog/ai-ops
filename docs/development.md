# Development Guide

## Prerequisites

- Go 1.21+
- Node.js 18+
- Make

## Getting Started

### Backend

```bash
# Install dependencies
make install

# Run tests
make test

# Start development server
make run
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

## Project Structure

See [architecture.md](architecture.md) for detailed project structure.

## Testing

### Backend Tests

```bash
# Run all tests
go test ./... -v

# Run specific package
go test ./internal/storage/... -v

# Run with coverage
go test -cover ./...
```

### Frontend Tests

```bash
cd frontend
npm test
```

## Configuration

Copy `configs/config.example.yaml` to `configs/config.yaml` and modify as needed.

### Configuration Fields

| Section | Field | Description |
|---------|-------|-------------|
| server | port | Server port (default: 8080) |
| server | mode | Run mode: development/production |
| database | type | Database type (sqlite) |
| database | path | Database file path |
| auth | jwt_secret | Secret for JWT signing |
| auth | jwt_expiry_hours | JWT token expiry time |
| ai | provider | AI provider (openai) |
| ai | api_key | OpenAI API key |
| ai | model | LLM model name |
| monitor | prometheus_retention_days | Prometheus data retention |
| log | level | Log level (debug/info/warn/error) |
| log | path | Log file path |

## Database

Database is automatically initialized on first run. Schema is defined in `internal/storage/sqlite/db.go`.

### Tables

- **users**: User accounts
- **clusters**: K8S cluster metadata
- **nodes**: Cluster nodes
- **conversations**: AI chat sessions
- **messages**: Chat messages
- **knowledge_base**: Knowledge base entries
- **deployment_templates**: Deployment templates
- **deployments**: Deployment tasks
- **alert_rules**: Alert rule definitions
- **alerts**: Alert events
- **remediations**: Auto-remediation records
- **backups**: Cluster backup records
- **audit_logs**: Operation audit logs

## API Documentation

API endpoints are documented in `docs/api.md` (to be created).

### Available Endpoints

- `GET /api/v1/system/health` - Health check
- `GET /api/v1/system/version` - Server version

## Contributing

1. Create a feature branch
2. Write tests first (TDD)
3. Implement feature
4. Ensure all tests pass
5. Submit PR

## Code Style

- Go: Follow [Effective Go](https://golang.org/doc/effective_go)
- TypeScript: ESLint rules in `frontend/.eslintrc.cjs`
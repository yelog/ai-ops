# Architecture

## Overview

AI-K8S-OPS is a modular monolith application with the following layers:

1. **Presentation Layer**: React frontend (TypeScript)
2. **API Layer**: Gin-based REST API (Go)
3. **Business Logic Layer**: Domain modules (Go)
4. **Data Layer**: SQLite + File storage (Go)

## Backend Structure

```
cmd/                 - Application entry points
  server/            - API server
  agent/             - Cluster agent (future)
  cli/               - Command-line tool (future)
  
internal/            - Private application code
  api/               - API handlers and routes
    handlers/        - Request handlers
  auth/              - Authentication and authorization (future)
  cluster/           - Cluster management (future)
  deploy/            - Deployment logic (future)
  ai/                - AI interaction (future)
  monitor/           - Monitoring integration (future)
  alert/             - Alert management (future)
  remediation/       - Auto-remediation (future)
  knowledge/         - Knowledge base (future)
  audit/             - Audit logging (future)
  storage/           - Data access layer
    sqlite/          - SQLite operations
    file/            - File storage operations (future)
  grpc/              - gRPC services (future)
  k8s/               - Kubernetes client wrapper (future)
  llm/               - LLM integration (future)
  
pkg/                 - Public packages (can be imported)
  config/            - Configuration management
  logger/            - Logging utilities (future)
  crypto/            - Encryption utilities (future)
  version/           - Version information
  
configs/             - Configuration files
scripts/             - Deployment and setup scripts (future)
data/                - Data storage (database, files, logs)
tests/               - Integration tests
```

## Frontend Structure

```
frontend/src/
  components/        - Reusable UI components (future)
  pages/             - Page components (future)
  services/          - API client services (future)
  stores/            - State management (future)
  hooks/             - Custom React hooks (future)
  utils/             - Utility functions (future)
  constants/         - Constants and enums (future)
  types/             - TypeScript type definitions (future)
  i18n/              - Internationalization (future)
```

## Data Flow

```
User → Frontend → API Gateway → Business Logic → Data Layer → Storage
                    ↓
                  AI Engine → LLM API (future)
```

## Technology Stack

### Backend

- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Database**: SQLite
- **Configuration**: YAML (gopkg.in/yaml.v3)
- **Testing**: Go testing package

### Frontend

- **Framework**: React 18+
- **Language**: TypeScript
- **Build Tool**: Vite
- **UI Library**: Ant Design Pro
- **State Management**: Zustand
- **Data Fetching**: React Query
- **Charts**: ECharts

### Infrastructure

- **Container Runtime**: Docker (future)
- **Orchestration**: Kubernetes (future)
- **Monitoring**: Prometheus + Grafana (future)
- **Logging**: Loki (future)

## Security Architecture

See [docs/plans/2026-03-30-ai-k8s-ops-design.md](plans/2026-03-30-ai-k8s-ops-design.md) Security Design section.

## Deployment

See deployment documentation (to be created).

## Future Enhancements

- Agent architecture for multi-cluster management
- AI-powered diagnostics and auto-remediation
- Real-time monitoring and alerting
- Knowledge base with RAG
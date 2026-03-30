# AI-K8S-OPS

AI-driven Kubernetes operations management platform.

## Features

- **Fast Deployment**: Quick K8S cluster setup with integrated scripts
- **AI Interaction**: Natural language interface for cluster management
- **Monitoring & Repair**: Real-time monitoring with proactive fix suggestions

## Architecture

- Backend: Go + Gin + SQLite
- Frontend: React + TypeScript + Ant Design Pro
- Agent: Lightweight cluster agent for multi-cluster management

## Quick Start

### Backend

```bash
make install
make run
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

## Development

See [docs/development.md](docs/development.md) for detailed development guide.

## Architecture

See [docs/architecture.md](docs/architecture.md) for system architecture.

## API Endpoints

### System

- `GET /api/v1/system/health` - Health check
- `GET /api/v1/system/version` - Server version

## Database Schema

Database: SQLite (data/ai-k8s-ops.db)

Tables: 13 (users, clusters, nodes, conversations, messages, knowledge_base, deployment_templates, deployments, alert_rules, alerts, remediations, backups, audit_logs)

See [docs/architecture.md](docs/architecture.md) for detailed schema.

## License

MIT
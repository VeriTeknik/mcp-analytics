# MCP Analytics Service

A high-performance analytics and search service for MCP (Model Context Protocol) servers, powering the plugged.in app's discovery and insights features.

## Overview

The Analytics Service provides:
- 🔍 **Advanced Search** - Full-text search with filtering, faceting, and sorting
- 📊 **Real-time Analytics** - Track installs, usage, ratings, and engagement
- 🚀 **Discovery Features** - Featured servers, trending, top-rated, and personalized recommendations
- 📈 **Metrics & Insights** - Comprehensive analytics for server performance and user behavior
- ⚡ **High Performance** - Built with Go, Elasticsearch, and Redis for speed and scale

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  plugged.in App │────▶│ Analytics API   │────▶│ Elasticsearch   │
└─────────────────┘     └────────┬────────┘     └─────────────────┘
                                 │                        
                                 ├────────▶ PostgreSQL (User Data)
                                 ├────────▶ MongoDB (Analytics)
                                 └────────▶ Redis (Cache)
```

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.23+ (for local development)
- 8GB RAM minimum
- 20GB disk space

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/pluggedin/mcp-analytics.git
   cd mcp-analytics
   ```

2. **Copy environment configuration**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start all services**
   ```bash
   make dev
   ```

4. **Check service health**
   ```bash
   make health-check
   ```

5. **View logs**
   ```bash
   make logs-analytics
   ```

### Service URLs
- **Analytics API**: http://localhost:8081
- **Kibana**: http://localhost:5601
- **Adminer (PostgreSQL)**: http://localhost:8082
- **Redis Commander**: http://localhost:8083

## API Documentation

See [PLUGGEDIN_INTEGRATION.md](PLUGGEDIN_INTEGRATION.md) for comprehensive API documentation and integration guide.

### Key Endpoints

#### Search
```bash
GET /v1/search?q=database&package_type=npm&sort=popularity
```

#### Discovery
```bash
GET /v1/featured
GET /v1/trending?period=week
GET /v1/top-rated
```

#### Analytics
```bash
GET /v1/servers/{id}/analytics
POST /v1/installs
POST /v1/ratings
POST /v1/usage
```

## Development

### Project Structure
```
mcp-analytics/
├── cmd/analytics/          # Application entry point
├── internal/
│   ├── api/               # HTTP handlers and routing
│   ├── analytics/         # Analytics business logic
│   ├── cache/            # Redis caching layer
│   ├── config/           # Configuration management
│   ├── database/         # Database connections
│   ├── model/            # Data models
│   ├── search/           # Elasticsearch integration
│   └── service/          # Core business services
├── scripts/              # Utility scripts
├── docs/                 # Documentation
└── deployments/          # Deployment configurations
```

### Common Tasks

```bash
# Run tests
make test

# Run with coverage
make test-coverage

# Format code
make fmt

# Run linter
make lint

# Build production image
make build

# Connect to databases
make psql          # PostgreSQL CLI
make mongo-shell   # MongoDB shell
make redis-cli     # Redis CLI
```

### Database Migrations

Migrations run automatically on startup. To run manually:
```bash
make migrate
```

### Seed Test Data

```bash
make seed
```

## Testing

### Unit Tests
```bash
go test ./...
```

### Integration Tests
```bash
go test -tags=integration ./...
```

### Load Testing
```bash
make load-test
```

## Deployment

### Production Build
```bash
make prod-build
```

### Docker Deployment
```bash
docker compose -f docker-compose.prod.yml up -d
```

### Kubernetes Deployment
```bash
kubectl apply -f deployments/k8s/
```

## Monitoring

### Health Checks
- **Health**: `/health` - Basic health check
- **Ready**: `/ready` - Checks all dependencies

### Metrics
- Prometheus metrics available at `/metrics`
- Custom business metrics for tracking
- Performance monitoring with OpenTelemetry

### Logging
- Structured JSON logging with zerolog
- Log aggregation ready
- Configurable log levels

## Configuration

See `.env.example` for all configuration options. Key settings:

- **Database URLs**: Configure connections to PostgreSQL, MongoDB, Elasticsearch, Redis
- **Cache TTLs**: Customize cache durations for different data types
- **Rate Limits**: Set API rate limits
- **Feature Flags**: Enable/disable features

## Performance

### Optimizations
- Response caching with Redis
- Database query optimization
- Elasticsearch query tuning
- Connection pooling
- Bulk operations

### Benchmarks
- Search: <50ms p99 latency
- API endpoints: <100ms p99 latency
- 10k+ requests/second capacity

## Security

- Internal API authentication
- Rate limiting per endpoint
- Input validation and sanitization
- SQL injection prevention
- XSS protection

## Troubleshooting

### Common Issues

1. **Elasticsearch fails to start**
   - Increase Docker memory limit
   - Check `vm.max_map_count` setting

2. **Connection refused errors**
   - Ensure all services are running: `docker compose ps`
   - Check service health: `make health-check`

3. **Slow search performance**
   - Check Elasticsearch indices: `make elastic-indices`
   - Review query complexity
   - Increase cache TTL

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run `make fmt lint test`
6. Submit a pull request

## License

Copyright © 2024 plugged.in. All rights reserved.

## Support

- Documentation: [PLUGGEDIN_INTEGRATION.md](PLUGGEDIN_INTEGRATION.md)
- Issues: [GitHub Issues](https://github.com/pluggedin/mcp-analytics/issues)
- Email: analytics@plugged.in
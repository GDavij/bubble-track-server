# Configuration

Environment-based configuration for Kubernetes deployment.

## Defaults (for K8s service discovery)

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_HOST` | postgres.database | PostgreSQL host (K8s service) |
| `POSTGRES_PORT` | 5432 | PostgreSQL port |
| `POSTGRES_USER` | bubble | Database user |
| `POSTGRES_DB` | bubble | Database name |
| `QDRANT_HOST` | qdrant.database | Qdrant host |
| `QDRANT_PORT` | 6334 | Qdrant gRPC port |
| `REDIS_HOST` | redis-stack.stack | Redis host |
| `REDIS_PORT` | 6379 | Redis port |
| `GENAI_API_KEY` | (required) | Gemini API key |
| `SERVER_PORT` | 8080 | HTTP server port |
| `SERVER_ENV` | production | production |

## Loading Config

```go
cfg := config.Load()
// Returns Config struct with all settings
```

## Config Structure

```go
type Config struct {
    Postgres  PostgresConfig
    Qdrant   QdrantConfig
    Redis    RedisConfig
    GenAI    GenAIConfig
    Server   ServerConfig
}

type PostgresConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    DB       string
}

type ServerConfig struct {
    Port string
    Env  string
}
```

## Kubernetes Deployment

```yaml
apiVersion: v1
kind: Service
metadata:
  name: bubble-track-server
spec:
  ports:
    - port: 8080
      targetPort: 8080
---
apiVersion: apps/v1
kind: Deployment
spec:
  containers:
    - name: server
      image: bubble-track-server:latest
      ports:
        - containerPort: 8080
      env:
        - name: POSTGRES_HOST
          value: "postgres.database"
        - name: QDRANT_HOST
          value: "qdrant.database"
        - name: REDIS_HOST
          value: "redis-stack.stack"
        - name: GENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: bubble-secrets
              key: genai-api-key
```
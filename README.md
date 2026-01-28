# libs-go~~~~

Shared Go libraries for the ControlPlane platform. This module provides foundational utilities, abstractions, and integrations used across multiple services. These libraries are provided in the hope that they may be as useful to you as they have been to our engineering team

**Requires Go 1.24.0+**

## Packages

### Configuration & Logging

| Package | Description |
|---------|-------------|
| **config** | Environment variable parsing with schema-driven configuration using `cpln:` struct tags. Supports type mapping, sensitive field masking, and configuration summarization. |
| **logging** | Structured logging built on Uber's Zap library. Provides context-aware logging with trace ID propagation and configurable log levels. |
| **errors** | Error code interface, error handling utilities, HTTP status code integration, and domain-specific error definitions. |

### Data Access & Storage

| Package | Description |
|---------|-------------|
| **database** | GORM-based database abstraction with connection pooling, read-write/read-only connection support, and PostgreSQL-specific repository patterns. |
| **postgresql** | PostgreSQL-specific utilities including containment query builders for complex WHERE clauses. |
| **checkpoints** | Checkpoint management for job tracking with query and reset operations. Uses GORM models with composite primary keys. |
| **cache** | Multi-backend caching abstraction supporting in-memory and Redis backends. Generic interface with TTL support and binary transcoding. |

### HTTP & Web

| Package | Description |
|---------|-------------|
| **web** | HTTP request/response utilities with JSON serialization, bearer token extraction, JWT validation, and typed generic request functions. |
| **data-service** | HTTP client for data-service API communication with CRUD operations, context-aware requests, and bearer token authentication. |
| **metering** | HTTP client for metering/metrics service with query composition, time-series decomposition, and bucket management. |

### Message Queue & Pub/Sub

| Package | Description |
|---------|-------------|
| **nats** | NATS message queue client with queue subscriptions, multiple workers, Prometheus metrics integration, and TLS/mTLS support. |
| **pubsub** | Abstract pub/sub interfaces with generic `Publisher[T]`, `Consumer[T]`, and `Client[T]` types. Backend-agnostic design works with NATS, Cloud Pub/Sub, etc. |
| **relay/replication** | PostgreSQL WAL replication for change data capture with insert/update/delete actions. Supports PostgreSQL and SQS destinations. |

### Concurrency & Job Management

| Package | Description |
|---------|-------------|
| **threading** | Background job execution framework with safe goroutine wrappers, work queues, and parallel execution utilities. |
| **leader-election** | Kubernetes-based leader election with role determination, service endpoint monitoring, and pod version-based ordering. |
| **process** | Graceful shutdown management with `Term()` and `Wait()` primitives for clean service termination. |
| **negotiate** | Distributed sync job framework with leader election, cron scheduling, work queue processing, and retry policies. |

### Type & Object Handling

| Package | Description |
|---------|-------------|
| **dynamic-objects** | Recursive object field extraction and manipulation with pivot/flatten operations, dot-notation field paths, and fingerprinting for change detection. |
| **scanner** | Reflection-based struct scanning with mapper registry for custom type conversions. |
| **types** | Type utility functions including pointer dereferencing and interface checking. |
| **metadata** | Struct tag parsing for `cpln:` tags with support for default values, mappers, env vars, and sensitive field markers. |

### Functional Programming & Collections

| Package | Description |
|---------|-------------|
| **pipeline** | Generic functional utilities including Map/Filter operations, Enumerable type for method chaining, and collection operations (Flatten, Difference, IndexOf). |
| **batches** | Slice batching utilities with `NewBatchIterator[S, T]()` for chunking large collections. |
| **maps** | Map merging utilities. |

### Time & String Utilities

| Package | Description |
|---------|-------------|
| **time-utils** | Time comparison utilities (UTC normalized), inclusive/exclusive range checking, time segmentation, and step definitions (hour, day, week, month). |
| **strings** | String case manipulation utilities. |
| **math** | Numeric utilities including clamp functions and decimal place rounding. |

### Kubernetes & Infrastructure

| Package | Description |
|---------|-------------|
| **crd** | Kubernetes CRD generation from Go structs with OpenAPI v3 schema derivation via reflection. |
| **images** | Image path translation for ControlPlane internal images. Converts `/org/{org}/image/{image}` references to full registry URLs. |

### Domain-Specific

| Package | Description |
|---------|-------------|
| **events** | Data service event definitions with types for created, updated, exists, and deleted events. |
| **terminal** | Terminal animation utilities. |
| **common** | Shared constants including trace ID keys, HTTP header constants (`X-Cpln-Workload-Link`, `X-Cpln-Identity-Link`), and context key types. |

## Usage

Import packages using the module path:

```go
import (
    "github.com/controlplane-com/libs-go/pkg/logging"
    "github.com/controlplane-com/libs-go/pkg/database"
    "github.com/controlplane-com/libs-go/pkg/cache"
)
```

## Testing

```bash
make test
```

## Dependencies

Key external dependencies:
- **Kubernetes**: k8s.io/client-go, k8s.io/apimachinery
- **Cloud Providers**: cloud.google.com/go (Firestore, Pub/Sub, Secret Manager)
- **Databases**: gorm.io/gorm, gorm.io/driver/postgres
- **Messaging**: github.com/nats-io/nats.go
- **Caching**: github.com/redis/go-redis, github.com/patrickmn/go-cache
- **Logging**: go.uber.org/zap
- **Observability**: github.com/prometheus/client_golang

# evalops/proto

Canonical protobuf definitions for EvalOps cross-service contracts.

## Packages

| Package | Description | Consumers |
|---------|-------------|-----------|
| `identity/v1` | Token introspection, organizations, members, sessions | gate, chat, llm-gateway, service-runtime |
| `meter/v1` | Usage recording and cost attribution | llm-gateway, chat |
| `audit/v1` | Audit event recording and querying | llm-gateway, chat, gate |
| `memory/v1` | Semantic memory storage and recall | chat, ensemble |

## Usage

### Go

```go
import (
    identityv1 "github.com/evalops/proto/gen/go/identity/v1"
    meterv1 "github.com/evalops/proto/gen/go/meter/v1"
)
```

### TypeScript

```typescript
import { RecallRequest } from "@evalops/proto/memory/v1/memory_pb";
```

## Development

```bash
# Lint proto files
make lint

# Check for breaking changes against main
make breaking

# Regenerate all code
make generate
```

## Adding a New Package

1. Create `proto/<service>/v1/<service>.proto`
2. Set `option go_package = "github.com/evalops/proto/gen/go/<service>/v1;<service>v1";`
3. Run `make lint && make generate`
4. Commit the proto file and generated code together

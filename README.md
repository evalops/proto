# Proto

`proto` is the canonical home for cross-service protobuf definitions in the
EvalOps ecosystem.

It holds `.proto` files for the shared contracts that multiple services depend
on, generates type-safe Go and TypeScript packages from those definitions, and
enforces schema compatibility through CI. Services import the generated
packages instead of maintaining their own hand-written struct copies.

## Goals

- Provide a single source of truth for cross-service API types.
- Generate Go, TypeScript, and Python packages from one set of definitions.
- Catch breaking contract changes at compile time, not in production.
- Keep wire encoding efficient with protobuf binary format.
- Support Connect-RPC service definitions for typed cross-service calls.

## Non-Goals

- This repo does not hold service-internal types that never cross a service boundary.
- This repo does not define the external REST/OpenAPI surface of any service.
- This repo does not own service implementation or business logic.
- This repo does not replace service-local proto definitions for internal RPC
  (e.g. gate's proxy protocol protos or chat's conversation streaming protos).

## Packages

| Package | Description | Consumers |
|---------|-------------|-----------|
| `identity/v1` | Token introspection, organizations, members, API keys, sessions | gate, chat, llm-gateway, service-runtime, identity |
| `meter/v1` | Usage recording, querying, cost attribution | llm-gateway, chat, platform |
| `audit/v1` | Audit event recording, actors, resources, querying, export | llm-gateway, chat, gate, platform |
| `memory/v1` | Semantic memory storage, recall, embeddings, consolidation | chat, ensemble, platform |
| `approvals/v1` | Approval workflows, policies, decisions, habits, escalation | approvals, ensemble, maestro, chat, objectives |
| `connectors/v1` | Integration connections, health, source-of-truth, capabilities | connectors, pipeline, parker, entities, objectives |
| `entities/v1` | Canonical entity resolution, search, links, correlation graph | entities, pipeline, parker, connectors, ensemble |
| `governance/v1` | Safety evaluation, PII detection, retention, legal holds | governance, approvals, objectives, ensemble, chat |
| `notifications/v1` | Notification delivery, preferences, channels, status | notifications, approvals, objectives, pipeline, parker |
| `objectives/v1` | Objective lifecycle, wake scheduling, provenance, mutations | objectives, approvals, governance, pipeline, parker |
| `skills/v1` | Shared skill registry, versions, search, import/export | skills, chat, ensemble, maestro, pipeline |
| `events/v1` | Shared NATS event envelope and change journal payloads | service-runtime, pipeline, parker |
| `tap/v1` | Normalized tap webhook payloads and field-level diffs | ensemble-tap, pipeline |
| `prompts/v1` | Prompt versioning, deployment tracking, eval linkage, resolution | prompts, llm-gateway, fermata, maestro, ensemble |

## Architecture Diagrams

### Contract Ownership

```mermaid
flowchart TD
    Proto[evalops/proto]
    Identity[identity]
    Meter[meter]
    Audit[audit]
    Memory[memory]
    Approvals[approvals]
    Connectors[connectors]
    Entities[entities]
    Governance[governance]
    Notifications[notifications]
    Objectives[objectives]
    Skills[skills]
    Gateway[llm-gateway]
    Chat[chat]
    Gate[gate]
    Ensemble[ensemble]
    SRT[service-runtime]

    Proto -->|identity/v1| Identity
    Proto -->|identity/v1| Gateway
    Proto -->|identity/v1| Chat
    Proto -->|identity/v1| Gate
    Proto -->|identity/v1| SRT
    Proto -->|meter/v1| Meter
    Proto -->|meter/v1| Gateway
    Proto -->|meter/v1| Chat
    Proto -->|audit/v1| Audit
    Proto -->|audit/v1| Gateway
    Proto -->|audit/v1| Chat
    Proto -->|audit/v1| Gate
    Proto -->|memory/v1| Memory
    Proto -->|memory/v1| Chat
    Proto -->|memory/v1| Ensemble
    Proto -->|approvals/v1| Approvals
    Proto -->|approvals/v1| Chat
    Proto -->|approvals/v1| Ensemble
    Proto -->|approvals/v1| Objectives
    Proto -->|connectors/v1| Connectors
    Proto -->|connectors/v1| Pipeline
    Proto -->|connectors/v1| Parker
    Proto -->|connectors/v1| Entities
    Proto -->|entities/v1| Entities
    Proto -->|entities/v1| Pipeline
    Proto -->|entities/v1| Parker
    Proto -->|entities/v1| Ensemble
    Proto -->|governance/v1| Governance
    Proto -->|governance/v1| Approvals
    Proto -->|governance/v1| Objectives
    Proto -->|governance/v1| Chat
    Proto -->|notifications/v1| Notifications
    Proto -->|notifications/v1| Approvals
    Proto -->|notifications/v1| Objectives
    Proto -->|objectives/v1| Objectives
    Proto -->|objectives/v1| Approvals
    Proto -->|objectives/v1| Governance
    Proto -->|skills/v1| Skills
    Proto -->|skills/v1| Chat
    Proto -->|skills/v1| Ensemble
    Proto -->|skills/v1| Maestro[maestro]
    Proto -->|events/v1| SRT
    Proto -->|events/v1| Parker[Parker]
    Proto -->|events/v1| Pipeline[Pipeline]
    Proto -->|tap/v1| EnsembleTap[ensemble-tap]
    Proto -->|tap/v1| Pipeline
    Proto -->|prompts/v1| Prompts[prompts]
    Proto -->|prompts/v1| Gateway
    Proto -->|prompts/v1| Fermata[fermata]
    Proto -->|prompts/v1| Maestro[maestro]
    Proto -->|prompts/v1| Ensemble
```

### What This Replaces

```mermaid
flowchart LR
    subgraph Before
        GW1[llm-gateway\nhand-written IntrospectionResult\n6 fields]
        SRT1[service-runtime\nhand-written IntrospectionResult\n11 fields]
        Chat1[chat\nhand-written issueServiceTokenResponse]
    end

    subgraph After
        Proto2[evalops/proto\nidentity/v1/tokens.proto\nIntrospectResponse\n15 fields]
        GW2[llm-gateway\nimports identityv1]
        SRT2[service-runtime\nimports identityv1]
        Chat2[chat\nimports identityv1]
        Proto2 --> GW2
        Proto2 --> SRT2
        Proto2 --> Chat2
    end
```

## Usage

### Go

```go
import (
    identityv1 "github.com/evalops/proto/gen/go/identity/v1"
    meterv1 "github.com/evalops/proto/gen/go/meter/v1"
    auditv1 "github.com/evalops/proto/gen/go/audit/v1"
    memoryv1 "github.com/evalops/proto/gen/go/memory/v1"
    promptsv1 "github.com/evalops/proto/gen/go/prompts/v1"
)

// Use generated types directly
req := &meterv1.RecordUsageRequest{
    OrganizationId: orgID,
    Model:          "claude-opus-4.6",
    PromptTokens:   1250,
    CostMicros:     42300,
}

// Resolve the active prompt version at inference time
resolveReq := &promptsv1.ResolveRequest{
    Name:    "sdr-outreach",
    Surface: "ensemble",
    Env:     "production",
}
```

### Go with Connect-RPC

```go
import (
    meterv1connect "github.com/evalops/proto/gen/go/meter/v1/meterv1connect"
    "connectrpc.com/connect"
)

client := meterv1connect.NewMeterServiceClient(
    http.DefaultClient,
    "https://meter.internal",
)
resp, err := client.RecordUsage(ctx, connect.NewRequest(req))
```

### TypeScript

```typescript
import { RecallRequest } from "@evalops/proto/memory/v1/memory_pb";
import { createClient } from "@connectrpc/connect-web";
import { MemoryService } from "@evalops/proto/memory/v1/memory_connect";
```

## Event Envelope Guidance

`events/v1.CloudEvent` is the canonical typed envelope for the shared internal
bus.

- Set `subject` to the actual bus subject when the transport has one. This
  keeps NATS routing context visible without forcing consumers to infer it from
  `type`.
- Keep `data` as a typed message packed in `google.protobuf.Any`; do not fall
  back to service-local ad hoc payload blobs.
- Prefer `data_content_type=application/protobuf` for the canonical envelope.
  Some services still emit structured CloudEvents JSON with
  `specversion`/`datacontenttype` and JSON `data` while they migrate. Consumers
  should tolerate both shapes until the bus converges, but new typed contracts
  should start from `events/v1.CloudEvent`.

## Development

Prerequisites:

- [buf](https://buf.build/docs/installation) v1.50+
- Go 1.22+

```bash
# Lint proto files
make lint

# Check for breaking changes against main
make breaking

# Regenerate Go and TypeScript packages
make generate

# Run contract tests for generated packages
make test

# Clean and regenerate from scratch
make clean generate
```

Fixture-backed contract examples live alongside the owning proto package under
`proto/<service>/v1/testdata/`. Keep those JSON examples aligned with the
canonical protojson field names because `make test` unmarshals them through the
generated types.

High-risk boundary fixtures belong there too. The current catalog includes
canonical `events/v1.CloudEvent` examples for `pipeline.changes.activity.create`
with `outcome=replied` and `parker.changes.work_relationship.update` with
`status=terminated`, so downstream consumers can pin the semantics they depend
on instead of only the wire shape.

## Adding a New Proto

1. Create `proto/<service>/v1/<file>.proto`.
2. Set the Go package option:
   ```protobuf
   option go_package = "github.com/evalops/proto/gen/go/<service>/v1;<service>v1";
   ```
3. Run `make lint` to validate.
4. Run `make generate` to produce Go and TypeScript packages.
5. Commit the `.proto` file and generated code together.
6. After merge, consumers can import the new package immediately.

## Adding a Breaking Change

`buf breaking` runs on every PR against `main`. If your change breaks wire
compatibility (renaming a field, changing a field number, removing a field),
the PR will fail CI.

To make a safe change:

- Add new fields with new field numbers. Old consumers ignore unknown fields.
- Deprecate fields with `[deprecated = true]` instead of removing them.
- If a field must be removed, reserve its number: `reserved 5;`

## Repository Layout

```text
buf.yaml                buf module configuration
buf.gen.yaml            codegen plugin configuration
proto/                  source .proto files
  identity/v1/          token introspection, orgs, members, sessions
  meter/v1/             usage recording and cost attribution
  audit/v1/             audit event recording and querying
  memory/v1/            semantic memory storage and recall
    testdata/           canonical protojson fixtures for contract tests
  approvals/v1/         approval workflows, policies, decisions, habits
  connectors/v1/        integration lifecycle, health, source-of-truth
  entities/v1/          canonical entity resolution and search
  governance/v1/        safety evaluation, PII detection, retention
  notifications/v1/     multi-channel delivery, preferences, status
  objectives/v1/        objective lifecycle, provenance, wake scheduling
  skills/v1/            shared skill registry and version history
  events/v1/            CloudEvent envelope and change journal payloads
    testdata/           typed event fixtures for cross-service contracts
  tap/v1/               normalized provider event payloads
  prompts/v1/           prompt versioning, deployment, eval linkage
gen/                    generated code (committed)
  go/                   Go protobuf + Connect-RPC packages
  ts/                   TypeScript protobuf-es + Connect-ES packages
.github/workflows/      CI: lint, breaking, generate, build
```

## CI

The CI pipeline runs on every push to `main` and every PR:

1. `buf lint` — validates proto files against the STANDARD rule set.
2. `buf breaking` — checks for wire-incompatible changes against `main` (PRs only).
3. `buf generate` — regenerates code and fails if the committed code is stale.
4. `go test` — verifies the generated Go packages and contract tests pass.

## Related Issues

- evalops/proto#1 — bootstrap tracking issue
- evalops/platform#9110 — cross-service protobuf coordination
- evalops/identity#56 — adopt identity proto types
- evalops/meter#30 — adopt meter proto types
- evalops/audit#23 — adopt audit proto types
- evalops/memory#32 — adopt memory proto types
- evalops/service-runtime#24 — shared protobuf event schemas

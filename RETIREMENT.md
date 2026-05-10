# Standalone Proto Retirement

`evalops/proto` is no longer the active source of truth for EvalOps protobuf
contracts. Active contract work belongs in `evalops/platform`.

## Source Of Truth

| Concern | Location |
| --- | --- |
| Protobuf schemas | `evalops/platform` `proto/` |
| Generated Go, TypeScript, Python, OpenAPI, and JSON Schema outputs | `evalops/platform` `gen/` |
| Drift and consolidation tracking | `evalops/platform` `docs/repositories/consolidation.json` |
| Wave 2 tracker | `evalops/platform#1768` |

## Allowed Changes Here

- Critical package compatibility fixes while downstream consumers finish moving
  to Platform-published artifacts.
- README, issue-routing, and repository metadata updates that make the retired
  state clearer.
- Emergency release repair, with a matching Platform issue or PR link.

## Not Allowed Here

- New protobuf packages or fields.
- Generated SDK refreshes that do not originate from Platform.
- CI, release, or fanout changes that make this repository look like the active
  contract upstream again.

## Contributor Flow

1. Open contract changes against `evalops/platform`.
2. Run the Platform generation and consolidation checks there.
3. Link the Platform PR or issue from any temporary compatibility work in this
   repository.
4. Prefer archiving this repository once the package compatibility window is
   closed.

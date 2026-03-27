@AGENTS.md

## Conventions
- `EventBus[T]` is generic — type parameter is the event type, matches `sse.Publisher[T]` from axon
- Import alias required: `axonnats "github.com/benaskins/axon-nats"` (conflicts with nats.go package name)
- Options pattern: `WithSubject(subject)` to configure NATS subject (default: "events")

## Constraints
- Depends on axon (for `sse.Publisher[T]`) and `nats-io/nats.go` only
- Do not add dependencies on other axon-* modules
- NATS must not leak into core axon — this repo is the boundary
- Do not add HTTP handlers — this provides pub/sub adapters only

## Testing
- `go test ./...` — tests skip gracefully when NATS is unavailable
- `go vet ./...` — must be clean
- Tests require a running NATS server; set `NATS_URL` or default `nats://127.0.0.1:4222`

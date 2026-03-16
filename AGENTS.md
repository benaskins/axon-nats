# axon-nats

## Overview

axon-nats provides NATS adapters for axon services. Package name: `nats`.

Import: `github.com/benaskins/axon-nats`

**Import alias:** Since this package and `github.com/nats-io/nats.go` both use package name `nats`, consumers should import with an alias: `axonnats "github.com/benaskins/axon-nats"`.

## Build & Test

```bash
go test ./...           # Tests skip gracefully when NATS is unavailable
go vet ./...
```

Tests require a running NATS server. Set `NATS_URL` or defaults to `nats://127.0.0.1:4222`.

## Contents

- **EventBus[T]** — NATS-backed pub/sub implementing `sse.Publisher[T]` from axon. Enables horizontal scaling of SSE services by fanning out events across instances via a NATS cluster.
- **WithSubject(subject)** — Option to set the NATS subject (default: "events").

## Dependencies

- `github.com/benaskins/axon` — for `sse.Publisher[T]` interface
- `github.com/nats-io/nats.go` — NATS client

# axon-nats

> Primitives · Part of the [lamina](https://github.com/benaskins/lamina-mono) workspace

NATS adapters for axon services. `EventBus[T]` implements `sse.Publisher[T]` from axon, enabling horizontal scaling of SSE services by fanning out events across all instances connected to a NATS cluster.

## Getting started

```bash
go get github.com/benaskins/axon-nats
```

```go
conn, _ := nats.Connect("nats://127.0.0.1:4222")
bus := axonnats.NewEventBus[MyEvent](conn, axonnats.WithSubject("chat.events"))

bus.Publish(MyEvent{Text: "hello"})

ch := bus.Subscribe("client-1")
ev := <-ch

bus.Unsubscribe("client-1")
bus.Close()
```

## Key types

- **`EventBus[T]`** — NATS-backed pub/sub implementing `sse.Publisher[T]`. Each subscriber gets a unique NATS subscription for fan-out delivery.
- **`NewEventBus[T](conn, opts...)`** — Constructor taking a `*nats.Conn` and functional options.
- **`WithSubject(subject)`** — Option to set the NATS subject (default: `"events"`).

## License

MIT

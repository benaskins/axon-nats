package nats

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/nats-io/nats.go"
)

// EventBus is a distributed pub/sub backed by NATS, enabling horizontal
// scaling of SSE services. Events published on any instance are delivered to
// subscribers on all instances connected to the same NATS cluster.
//
// EventBus implements push.Publisher[T].
type EventBus[T any] struct {
	conn    *nats.Conn
	subject string

	mu   sync.Mutex
	subs map[string]*client[T]
}

type client[T any] struct {
	ch   chan T
	nsub *nats.Subscription
}

// Option configures an EventBus.
type Option func(*config)

type config struct {
	subject string
}

// WithSubject sets the NATS subject used for publishing and subscribing.
// Defaults to "events".
func WithSubject(subject string) Option {
	return func(c *config) {
		c.subject = subject
	}
}

// NewEventBus creates an EventBus connected to the given NATS connection.
// Each subscriber gets a unique NATS subscription on the configured subject,
// ensuring fan-out delivery across all instances.
func NewEventBus[T any](conn *nats.Conn, opts ...Option) *EventBus[T] {
	cfg := config{subject: "events"}
	for _, o := range opts {
		o(&cfg)
	}
	return &EventBus[T]{
		conn:    conn,
		subject: cfg.subject,
		subs:    make(map[string]*client[T]),
	}
}

func (b *EventBus[T]) Subscribe(clientID string) <-chan T {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan T, 16)

	nsub, err := b.conn.Subscribe(b.subject, func(msg *nats.Msg) {
		var ev T
		if err := json.Unmarshal(msg.Data, &ev); err != nil {
			slog.Warn("nats event bus: failed to unmarshal event",
				"client", clientID, "error", err)
			return
		}
		select {
		case ch <- ev:
		default:
			slog.Warn("nats event bus: dropping event for slow subscriber",
				"client", clientID)
		}
	})
	if err != nil {
		slog.Error("nats event bus: failed to subscribe",
			"client", clientID, "error", err)
		close(ch)
		return ch
	}

	b.subs[clientID] = &client[T]{ch: ch, nsub: nsub}
	return ch
}

func (b *EventBus[T]) Unsubscribe(clientID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	c, ok := b.subs[clientID]
	if !ok {
		return
	}

	_ = c.nsub.Unsubscribe()
	close(c.ch)
	delete(b.subs, clientID)
}

func (b *EventBus[T]) Publish(ev T) {
	data, err := json.Marshal(ev)
	if err != nil {
		slog.Error("nats event bus: failed to marshal event", "error", err)
		return
	}

	if err := b.conn.Publish(b.subject, data); err != nil {
		slog.Error("nats event bus: failed to publish", "error", err)
	}
}

// Close drains all subscriptions and closes local channels.
func (b *EventBus[T]) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for id, c := range b.subs {
		_ = c.nsub.Unsubscribe()
		close(c.ch)
		delete(b.subs, id)
	}
}

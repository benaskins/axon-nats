package nats

import (
	"os"
	"testing"
	"time"

	"github.com/benaskins/axon/sse"
	"github.com/nats-io/nats.go"
)

type testEvent struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// Compile-time check: EventBus satisfies sse.Publisher.
var _ sse.Publisher[testEvent] = (*EventBus[testEvent])(nil)

func natsConn(t *testing.T) *nats.Conn {
	t.Helper()
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}
	nc, err := nats.Connect(url, nats.Timeout(2*time.Second))
	if err != nil {
		t.Skipf("skipping: NATS not available at %s: %v", url, err)
	}
	t.Cleanup(nc.Close)
	return nc
}

func TestEventBus_SubscribeReceivesEvents(t *testing.T) {
	nc := natsConn(t)
	bus := NewEventBus[testEvent](nc, WithSubject("test.subscribe"))
	t.Cleanup(bus.Close)

	ch := bus.Subscribe("client1")
	defer bus.Unsubscribe("client1")

	nc.Flush()

	bus.Publish(testEvent{Type: "image", ID: "img-1"})
	nc.Flush()

	select {
	case ev := <-ch:
		if ev.Type != "image" {
			t.Errorf("expected type image, got %s", ev.Type)
		}
		if ev.ID != "img-1" {
			t.Errorf("expected ID img-1, got %s", ev.ID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestEventBus_UnsubscribeStopsEvents(t *testing.T) {
	nc := natsConn(t)
	bus := NewEventBus[testEvent](nc, WithSubject("test.unsub"))
	t.Cleanup(bus.Close)

	ch := bus.Subscribe("client1")
	bus.Unsubscribe("client1")

	ev, ok := <-ch
	if ok {
		t.Fatal("expected channel to be closed after unsubscribe")
	}
	if ev.Type != "" {
		t.Errorf("expected zero value from closed channel, got %s", ev.Type)
	}
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	nc := natsConn(t)
	bus := NewEventBus[testEvent](nc, WithSubject("test.multi"))
	t.Cleanup(bus.Close)

	ch1 := bus.Subscribe("client1")
	ch2 := bus.Subscribe("client2")
	defer bus.Unsubscribe("client1")
	defer bus.Unsubscribe("client2")

	nc.Flush()

	bus.Publish(testEvent{Type: "image", ID: "img-1"})
	nc.Flush()

	for _, ch := range []<-chan testEvent{ch1, ch2} {
		select {
		case ev := <-ch:
			if ev.ID != "img-1" {
				t.Errorf("expected img-1, got %s", ev.ID)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timed out")
		}
	}
}

func TestEventBus_CrossInstance(t *testing.T) {
	nc := natsConn(t)

	bus1 := NewEventBus[testEvent](nc, WithSubject("test.cross"))
	bus2 := NewEventBus[testEvent](nc, WithSubject("test.cross"))
	t.Cleanup(bus1.Close)
	t.Cleanup(bus2.Close)

	ch := bus2.Subscribe("client-on-instance-2")
	defer bus2.Unsubscribe("client-on-instance-2")

	nc.Flush()

	bus1.Publish(testEvent{Type: "update", ID: "u-1"})
	nc.Flush()

	select {
	case ev := <-ch:
		if ev.Type != "update" || ev.ID != "u-1" {
			t.Errorf("unexpected event: %+v", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for cross-instance event")
	}
}

func TestEventBus_PublishNonBlocking(t *testing.T) {
	nc := natsConn(t)
	bus := NewEventBus[testEvent](nc, WithSubject("test.nonblock"))
	t.Cleanup(bus.Close)

	_ = bus.Subscribe("slow-client")
	defer bus.Unsubscribe("slow-client")

	nc.Flush()

	done := make(chan struct{})
	go func() {
		for i := 0; i < 20; i++ {
			bus.Publish(testEvent{Type: "image", ID: "img"})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("publish blocked on slow subscriber")
	}
}

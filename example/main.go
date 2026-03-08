package main

import (
	"fmt"
	"log"

	axonnats "github.com/benaskins/axon-nats"
	"github.com/nats-io/nats.go"
)

type ChatEvent struct {
	Room    string `json:"room"`
	Message string `json:"message"`
}

func main() {
	conn, err := nats.Connect("nats://127.0.0.1:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	bus := axonnats.NewEventBus[ChatEvent](conn, axonnats.WithSubject("chat.events"))
	defer bus.Close()

	ch := bus.Subscribe("demo-client")

	bus.Publish(ChatEvent{Room: "general", Message: "hello from axon-nats"})

	ev := <-ch
	fmt.Printf("received: room=%s message=%s\n", ev.Room, ev.Message)

	bus.Unsubscribe("demo-client")
}

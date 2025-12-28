// Package eventbus provides an interface for event bus.
package eventbus

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

type Event string

const (
	EventTagMutation      Event = "tag.mutation"
	EventLocationMutation Event = "location.mutation"
	EventItemMutation     Event = "item.mutation"
)

type GroupMutationEvent struct {
	GID uuid.UUID
}

type eventData struct {
	event Event
	data  any
}

type EventBus struct {
	started bool
	ch      chan eventData

	mu          sync.RWMutex
	subscribers map[Event][]func(any)
}

func New() *EventBus {
	return &EventBus{
		ch: make(chan eventData, 100),
		subscribers: map[Event][]func(any){
			EventTagMutation:      {},
			EventLocationMutation: {},
			EventItemMutation:     {},
		},
	}
}

func (e *EventBus) Run(ctx context.Context) error {
	if e.started {
		panic("event bus already started")
	}

	e.started = true

	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-e.ch:
			e.mu.RLock()
			arr, ok := e.subscribers[event.event]
			e.mu.RUnlock()

			if !ok {
				continue
			}

			for _, fn := range arr {
				fn(event.data)
			}
		}
	}
}

func (e *EventBus) Publish(event Event, data any) {
	e.ch <- eventData{
		event: event,
		data:  data,
	}
}

func (e *EventBus) Subscribe(event Event, fn func(any)) {
	e.mu.Lock()
	defer e.mu.Unlock()

	arr, ok := e.subscribers[event]
	if !ok {
		panic("event not found")
	}

	e.subscribers[event] = append(arr, fn)
}

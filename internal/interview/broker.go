package interview

import "sync"

type broker struct {
	mu     sync.Mutex
	nextID int
	subs   map[int]chan Event
	closed bool
}

func newBroker() *broker {
	return &broker{
		subs: make(map[int]chan Event),
	}
}

func (b *broker) subscribe() (<-chan Event, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		ch := make(chan Event)
		close(ch)
		return ch, func() {}
	}

	id := b.nextID
	b.nextID++

	ch := make(chan Event, 64)
	b.subs[id] = ch

	return ch, func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if existing, ok := b.subs[id]; ok {
			delete(b.subs, id)
			close(existing)
		}
	}
}

func (b *broker) publish(event Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	for id, ch := range b.subs {
		select {
		case ch <- event:
		default:
			delete(b.subs, id)
			close(ch)
		}
	}
}

func (b *broker) close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}
	b.closed = true

	for id, ch := range b.subs {
		delete(b.subs, id)
		close(ch)
	}
}

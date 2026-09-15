package notify

import (
	"encoding/json"
	"sync"
	"testing"
	"time"
)

func TestInProcessFanout_DeliversToSubscribers(t *testing.T) {
	f := NewInProcessFanout()
	ch := f.Subscribe("tenant-a")
	defer f.Unsubscribe("tenant-a", ch)

	f.Publish("tenant-a", Envelope{Payload: json.RawMessage(`{"id":1}`)})

	select {
	case env := <-ch:
		if string(env.Payload) != `{"id":1}` {
			t.Fatalf("wrong payload: %s", env.Payload)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("subscriber never received event")
	}
}

func TestInProcessFanout_IsolatesTenants(t *testing.T) {
	f := NewInProcessFanout()
	a := f.Subscribe("tenant-a")
	b := f.Subscribe("tenant-b")
	defer f.Unsubscribe("tenant-a", a)
	defer f.Unsubscribe("tenant-b", b)

	f.Publish("tenant-a", Envelope{Payload: json.RawMessage(`{"x":1}`)})

	select {
	case <-b:
		t.Fatal("tenant-b received tenant-a event")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestInProcessFanout_UnsubscribeStopsDelivery(t *testing.T) {
	f := NewInProcessFanout()
	ch := f.Subscribe("t")
	f.Unsubscribe("t", ch)

	// After unsubscribe channel is closed; further publish must not panic.
	f.Publish("t", Envelope{Payload: json.RawMessage(`{}`)})
}

func TestInProcessFanout_DropsWhenBufferFull(t *testing.T) {
	f := NewInProcessFanout()
	ch := f.Subscribe("t")
	defer f.Unsubscribe("t", ch)

	// Fill the buffer without draining.
	for range 200 {
		f.Publish("t", Envelope{Payload: json.RawMessage(`{}`)})
	}
	// Test just asserts no goroutine explodes; delivery is best-effort.
	_ = ch
}

func TestInProcessFanout_ConcurrentPublishSafe(t *testing.T) {
	f := NewInProcessFanout()
	ch := f.Subscribe("t")
	defer f.Unsubscribe("t", ch)

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f.Publish("t", Envelope{Payload: json.RawMessage(`{}`)})
		}()
	}
	wg.Wait()
}

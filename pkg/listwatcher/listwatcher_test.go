package listwatcher

import (
	"testing"
	"time"

	"github.com/streadway/amqp"
)

func TestWatchQueue_NoBlock(t *testing.T) {
	// TODO
	lw, err := NewListWatcher(DefaultListwatcherConfig())
	if err != nil {
		t.Fatal(err)
	}

	handler := func(msg amqp.Delivery) {

		// t.Log(string(msg.ContentType))
		t.Log(string(msg.Body))
	}

	cancel, err := lw.WatchQueue_NoBlock("apiServer", handler)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("start to sleep 10 seconds")
	time.Sleep(time.Second * 20)
	t.Log("start to cancel")
	cancel()
}

func TestWatchQueue_Block(t *testing.T) {
	// TODO
	lw, err := NewListWatcher(DefaultListwatcherConfig())
	if err != nil {
		t.Fatal(err)
	}

	handler := func(msg amqp.Delivery) {

		// t.Log(string(msg.ContentType))
		t.Log(string(msg.Body))
	}

	stop := make(chan struct{})
	go lw.WatchQueue_Block("apiServer", handler, stop)
	t.Log("start to sleep 10 seconds")
	time.Sleep(time.Second * 10)
	t.Log("start to stop")
	close(stop)
}

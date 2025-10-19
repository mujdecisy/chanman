package chanman_test

import (
	"reflect"
	"testing"

	"github.com/mujdecisy/chanman"
)

func TestInitChanAndDestroyChan(t *testing.T) {
	allowedTypes := []reflect.Type{reflect.TypeOf("")}
	name := "testchan1"
	// Should succeed
	err := chanman.InitChan(name, 1, allowedTypes)
	if err != nil {
		t.Fatalf("InitChan failed: %v", err)
	}
	// Should fail (duplicate)
	err = chanman.InitChan(name, 1, allowedTypes)
	if err == nil {
		t.Fatalf("InitChan should fail for duplicate channel name")
	}
	// Should succeed
	err = chanman.DestroyChan(name)
	if err != nil {
		t.Fatalf("DestroyChan failed: %v", err)
	}
	// Should fail (already destroyed)
	err = chanman.DestroyChan(name)
	if err == nil {
		t.Fatalf("DestroyChan should fail for non-existent channel")
	}
}

func TestPubSub(t *testing.T) {
	allowedTypes := []reflect.Type{reflect.TypeOf(42)}

	name := "testchan2"
	err := chanman.InitChan(name, 2, allowedTypes)
	if err != nil {
		t.Fatalf("InitChan failed: %v", err)
	}

	err = chanman.Sub(name, func(msg chanman.ChanMsg) bool {
		_, ok := msg.Data.(int)
		if !ok {
			t.Errorf("Expected int, got %T", msg.Data)
		}
		return false
	})
	if err != nil {
		t.Fatalf("Sub failed: %v", err)
	}

	// Should succeed
	err = chanman.Pub(name, 123)
	if err != nil {
		t.Fatalf("Pub failed: %v", err)
	}

	// Should fail
	err = chanman.Pub(name, "456")
	if err == nil {
		t.Errorf("Pub should fail when not int")
	}

	err = chanman.DestroyChan(name)
	if err != nil {
		t.Fatalf("DestroyChan failed: %v", err)
	}
}

func TestSubTwice(t *testing.T) {
	allowedTypes := []reflect.Type{reflect.TypeOf(1.0)}
	name := "testchan3"
	chanman.InitChan(name, 1, allowedTypes)
	err := chanman.Sub(name, func(msg chanman.ChanMsg) bool { return false })
	if err != nil {
		t.Fatalf("First Sub failed: %v", err)
	}
	err = chanman.Sub(name, func(msg chanman.ChanMsg) bool { return false })
	if err != nil {
		t.Errorf("Second Sub failed: %v", err)
	}
	chanman.DestroyChan(name)
}

func TestPubCallerInfo(t *testing.T) {
	name := "testchan4"
	allowedTypes := []reflect.Type{reflect.TypeOf("")}

	chanman.SetVerbose(true)
	defer chanman.SetVerbose(false)

	err := chanman.InitChan(name, 1, allowedTypes)
	if err != nil {
		t.Fatalf("InitChan failed: %v", err)
	}

	err = chanman.Sub(name, func(msg chanman.ChanMsg) bool { return false })
	if err != nil {
		t.Fatalf("Sub failed: %v", err)
	}

	err = pubWrapper(name)
	if err != nil {
		t.Fatalf("Pub failed: %v", err)
	}

	err = chanman.DestroyChan(name)
	if err != nil {
		t.Fatalf("DestroyChan failed: %v", err)
	}
}

func pubWrapper(channelName string) error {
	return chanman.Pub(channelName, "test message")
}

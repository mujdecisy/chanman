package chanman

import (
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
)

func InitChan(name string, bufferSize int, allowedMessageTypes []reflect.Type) error {
	_, err := getChan(name)
	if err == nil {
		return fmt.Errorf("channel with name %s already exists", name)
	}

	newChan := &chanDef{
		name:            name,
		bufferSize:      bufferSize,
		allowedMsgTypes: allowedMessageTypes,
		listenerCount:   0,
		channel:         make(chan ChanMsg, bufferSize),
		killChanList:    []*chan int{},
	}

	addChan(newChan)
	logf("INF", "%s initialized\n\tbuffersize: %d", name, bufferSize)

	return nil
}

func DestroyChan(name string) error {
	ch, err := getChan(name)
	if err != nil {
		return fmt.Errorf("channel with name %s not found: %w", name, err)
	}

	if ch.listenerCount > 0 {
		for _, kc := range ch.killChanList {
			*kc <- 1
		}
	}

	time.Sleep(200 * time.Millisecond)
	removeChan(name)

	return nil
}

func Sub(channelName string, listenerFunction func(msg ChanMsg) bool) error {
	chdef, err := getChan(channelName)
	if err != nil {
		logf("ERR", "%s sub failed\n\t%v", channelName, err)
		return fmt.Errorf("failed to subscribe to channel %s: %w", channelName, err)
	}

	increaseListenerCount(channelName)

	if chdef.listenerCount > 1 {
		logf("WRN", "%s already has listeners\n\tlistener count: %d", channelName, chdef.listenerCount)
	}

	killChan := make(chan int)
	chdef.killChanList = append(chdef.killChanList, &killChan)

	go func() {
		logf("INF", "%s sub started", channelName)
		for {
			select {
			case msg := <-chdef.channel:
				logf("INF", "%s recieved msg <%s>", channelName, msg.Uuid)
				stopListening := listenerFunction(msg)

				if stopListening {
					decreaseListenerCount(channelName)
					logf("INF", "%s lost a listener\n\tlistener count: %d", channelName, chdef.listenerCount)
				}
			case <-killChan:
				logf("INF", "%s is closing", channelName)
				return
			}

		}
	}()

	return nil
}

func Pub(channelName string, msg any) error {
	chdef, err := getChan(channelName)
	if err != nil {
		return fmt.Errorf("failed to publish to channel %s: %w", channelName, err)
	}

	if chdef.listenerCount == 0 {
		return fmt.Errorf("no subscribers found for channel %s, message will be discarded", channelName)
	}

	if len(chdef.channel) >= chdef.bufferSize {
		return fmt.Errorf("channel %s is full, message will be discarded", channelName)
	}

	msgType := reflect.TypeOf(msg)
	pubAllowed := false
	for _, allowedType := range chdef.allowedMsgTypes {
		if msgType == allowedType {
			pubAllowed = true
			break
		}
	}

	if !pubAllowed {
		return fmt.Errorf("message type %s not allowed for channel %s", msgType, channelName)
	}

	msgUuid := uuid.New().String()[:8]
	chdef.channel <- ChanMsg{Uuid: msgUuid, Name: channelName, Data: msg}

	logf("INF", "%s published msg<%s>", channelName, msgUuid)

	return nil
}

func SetVerbose(verbose bool) {
	chanManServiceInstance.mutex.Lock()
	defer chanManServiceInstance.mutex.Unlock()

	chanManServiceInstance.verbose = verbose
	if verbose {
		logf("INF", "verbose mode enabled")
	} else {
		logf("INF", "verbose mode disabled")
	}
}

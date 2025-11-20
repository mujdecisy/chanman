package chanman

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"
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
		msgCounter:      0,
		killChanList:    []*chan int{},
		mutex:           sync.RWMutex{},
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
				if msg.Id != "" {
					logf("INF", "%s recieved msg#%d <%s>", channelName, msg.Number, msg.Id)
				} else {
					logf("INF", "%s recieved msg#%d", channelName, msg.Number)
				}

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
	pc, _, _, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()

	chanMsg, err := pubWithId(channelName, msg, "")
	if err != nil {
		return fmt.Errorf("failed to publish to channel %s: %w", channelName, err)
	}

	if chanMsg.Id != "" {
		logf("INF", "%s published msg#%d <%s> by [%s]", channelName, chanMsg.Number, chanMsg.Id, funcName)
	} else {
		logf("INF", "%s published msg#%d by [%s]", channelName, chanMsg.Number, funcName)
	}
	return nil
}

func PubWithId(channelName string, msg any, id string) error {
	pc, _, _, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()

	chanMsg, err := pubWithId(channelName, msg, id)
	if err != nil {
		return fmt.Errorf("failed to publish to channel %s: %w", channelName, err)
	}

	logf("INF", "%s published msg#%d <%s> by [%s]", channelName, chanMsg.Number, chanMsg.Id, funcName)
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

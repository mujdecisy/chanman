package chanman

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
)

var chanDefs = []chanDef{}

type chanDef struct {
	name            string
	bufferSize      int
	allowedMsgTypes []reflect.Type
	listenerCount   int
	channel         chan ChanMsg
}

type ChanMsg struct {
	Uuid string
	Name string
	Data any
}

func InitChan(name string, bufferSize int, allowedMessageTypes []reflect.Type) error {
	for _, def := range chanDefs {
		if def.name == name {
			return fmt.Errorf("channel with name %s already exists", name)
		}
	}
	chanDefs = append(chanDefs, chanDef{
		name:            name,
		bufferSize:      bufferSize,
		allowedMsgTypes: allowedMessageTypes,
		listenerCount:   0,
		channel:         make(chan ChanMsg, bufferSize),
	})

	return nil
}

func Sub(channelName string, listenerFunction func(msg ChanMsg) bool) error {
	chdef, err := getChan(channelName)
	if err != nil {
		return fmt.Errorf("failed to subscribe to channel %s: %w", channelName, err)
	}

	if chdef.listenerCount > 0 {
		fmt.Printf("[WRN] %d listener/s already exists for channel %s, adding another listener", chdef.listenerCount, channelName)
	}

	chdef.listenerCount++

	go func() {
		for msg := range chdef.channel {
			stopListening := listenerFunction(msg)

			if stopListening {
				chdef.listenerCount--
				if chdef.listenerCount == 0 {
					fmt.Printf("[WRN] No more listeners for channel %s, closing channel\n", channelName)
					if err := removeChan(channelName); err != nil {
						fmt.Printf("[ERR] Failed to remove channel %s: %v\n", channelName, err)
					}
				}
			}
		}
	}()

	return nil
}

func Pub(channelName string, msg any) error {
	chdef, err := getChan(channelName)
	if err != nil {
		return fmt.Errorf("failed to subscribe to channel %s: %w", channelName, err)
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

	return nil
}

func getChan(name string) (*chanDef, error) {
	for _, def := range chanDefs {
		if def.name == name {
			return &def, nil
		}
	}
	return nil, fmt.Errorf("channel with name %s not found", name)
}

func removeChan(name string) error {
	for i, def := range chanDefs {
		if def.name == name {
			close(def.channel)
			chanDefs = append(chanDefs[:i], chanDefs[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("channel with name %s not found", name)
}

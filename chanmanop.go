package chanman

import (
	"fmt"
	"reflect"
)

func logf(level, format string, args ...any) {
	if !chanManServiceInstance.verbose {
		return
	}
	switch level {
	case "INF":
		fmt.Printf("chmn [INF] "+format+"\n", args...)
	case "WRN":
		fmt.Printf("chmn [WRN] "+format+"\n", args...)
	case "ERR":
		fmt.Printf("chmn [ERR] "+format+"\n", args...)
	default:
		fmt.Printf("chmn [UNK] "+format+"\n", args...)
	}
}

func getChan(name string) (*chanDef, error) {
	chanManServiceInstance.mutex.RLock()
	defer chanManServiceInstance.mutex.RUnlock()

	cds := &chanManServiceInstance.chanDefs

	for i := range *cds {
		if (*cds)[i].name == name {
			return (*cds)[i], nil
		}
	}

	return nil, fmt.Errorf("channel with name %s not found", name)
}

func removeChan(name string) error {
	chanManServiceInstance.mutex.Lock()
	defer chanManServiceInstance.mutex.Unlock()

	cds := &chanManServiceInstance.chanDefs

	for i, def := range *cds {
		if def.name == name {
			close(def.channel)
			*cds = append((*cds)[:i], (*cds)[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("channel with name %s not found", name)
}

func addChan(newChDef *chanDef) error {
	chanManServiceInstance.mutex.Lock()
	defer chanManServiceInstance.mutex.Unlock()

	cds := &chanManServiceInstance.chanDefs

	for _, def := range *cds {
		if def.name == newChDef.name {
			return fmt.Errorf("channel with name %s already exists", newChDef.name)
		}
	}

	*cds = append(*cds, newChDef)
	logf("INF", "%s added with buffer size %d", newChDef.name, newChDef.bufferSize)
	return nil
}

func increaseListenerCount(name string) error {
	chanManServiceInstance.mutex.Lock()
	defer chanManServiceInstance.mutex.Unlock()

	cds := &chanManServiceInstance.chanDefs

	for i := range *cds {
		if (*cds)[i].name == name {
			(*cds)[i].listenerCount++
			return nil
		}
	}
	return fmt.Errorf("channel with name %s not found", name)
}

func decreaseListenerCount(name string) error {
	chanManServiceInstance.mutex.Lock()
	defer chanManServiceInstance.mutex.Unlock()

	cds := &chanManServiceInstance.chanDefs

	for i := range *cds {
		if (*cds)[i].name == name {
			if (*cds)[i].listenerCount > 0 {
				(*cds)[i].listenerCount--
			}
			return nil
		}
	}
	return fmt.Errorf("channel with name %s not found", name)
}

func getAndIncMsgNumber(name string) (int64, error) {
	cds := &chanManServiceInstance.chanDefs
	for i := range *cds {
		if (*cds)[i].name == name {
			(*cds)[i].mutex.Lock()
			defer (*cds)[i].mutex.Unlock()
			number := (*cds)[i].msgCounter
			(*cds)[i].msgCounter++
			return number, nil
		}
	}
	return -1, fmt.Errorf("channel with name %s not found", name)
}

func pubWithTag(channelName string, msg any, tag string) (ChanMsg, error) {
	chdef, err := getChan(channelName)
	if err != nil {
		return ChanMsg{}, fmt.Errorf("failed to publish to channel %s: %w", channelName, err)
	}

	if chdef.listenerCount == 0 {
		return ChanMsg{}, fmt.Errorf("no subscribers found for channel %s, message will be discarded", channelName)
	}

	if len(chdef.channel) >= chdef.bufferSize {
		return ChanMsg{}, fmt.Errorf("channel %s is full, message will be discarded", channelName)
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
		return ChanMsg{}, fmt.Errorf("message type %s not allowed for channel %s", msgType, channelName)
	}

	msgNumber, err := getAndIncMsgNumber(channelName)
	if err != nil {
		return ChanMsg{}, fmt.Errorf("failed to get message counter for channel %s: %w", channelName, err)
	}

	chanMsg := ChanMsg{Tag: tag, Number: msgNumber, Name: channelName, Data: msg}
	chdef.channel <- chanMsg

	return chanMsg, nil
}

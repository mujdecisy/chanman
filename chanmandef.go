package chanman

import (
	"reflect"
	"sync"
)

type ChanMsg struct {
	Uuid string
	Name string
	Data any
}

type chanDef struct {
	name            string
	bufferSize      int
	allowedMsgTypes []reflect.Type
	listenerCount   int
	channel         chan ChanMsg
	killChanList    []*chan int
}

type chanManService struct {
	chanDefs []*chanDef
	mutex    sync.RWMutex
	verbose  bool
}

var chanManServiceInstance = &chanManService{
	chanDefs: make([]*chanDef, 0),
	mutex:    sync.RWMutex{},
	verbose:  true,
}

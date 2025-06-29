package chanman

import "fmt"

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

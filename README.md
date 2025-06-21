# CHANMAN

An umbrella project for managing multiple channels in your golang application.

# Installation

```bash
go get github.com/mujdecisy/chanman
```

# Usage

```go
package main

import (
    "fmt"
    "reflect"
    "time"

    "github.com/mujdecisy/chanman"
)

func main() {
    // Define allowed message types
    allowedTypes := []reflect.Type{reflect.TypeOf("")}

    // Initialize a channel named "mychan" with buffer size 2
    err := chanman.InitChan("mychan", 2, allowedTypes)
    if err != nil {
        panic(err)
    }

    // Subscribe to the channel
    err = chanman.Sub("mychan", func(msg chanman.ChanMsg) bool {
        fmt.Printf("Received: %v (uuid: %s)\n", msg.Data, msg.Uuid)
        // Return false to keep listening, true to stop
        return false
    })
    if err != nil {
        panic(err)
    }

    // Publish a message
    err = chanman.Pub("mychan", "hello world!")
    if err != nil {
        panic(err)
    }

    // Give goroutine time to process
    time.Sleep(100 * time.Millisecond)
}
```

# CHANMAN

A lightweight channel manager for Go applications that provides named channels with type safety, listener timeouts, and message tracking.

## Features

- 🔒 **Type-safe channels**: Define allowed message types per channel
- 🏷️ **Named channels**: Access channels by string names across your application
- ⏱️ **Listener timeouts**: Prevent slow listeners from blocking message processing
- 📊 **Message tracking**: Sequential message numbering and optional tagging
- 🔍 **Debug logging**: Built-in verbose mode for troubleshooting
- 🧵 **Thread-safe**: Concurrent access with mutex protection
- 📦 **Multiple subscribers**: Support for multiple listeners per channel

## Installation

```bash
go get github.com/mujdecisy/chanman
```

## Quick Start

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

    // Subscribe to the channel with a 300ms timeout
    err = chanman.Sub("mychan", 300*time.Millisecond, func(msg chanman.ChanMsg) bool {
        fmt.Printf("Received msg#%d: %v\n", msg.Number, msg.Data)
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

    // Destroy the channel when done
    err = chanman.DestroyChan("mychan")
    if err != nil {
        panic(err)
    }
}
```

## API Reference

### Channel Management

#### InitChan
```go
func InitChan(name string, bufferSize int, allowedMessageTypes []reflect.Type) error
```
Initialize a new named channel with a specified buffer size and allowed message types.

**Parameters:**
- `name`: Unique identifier for the channel
- `bufferSize`: Number of messages the channel can buffer
- `allowedMessageTypes`: Slice of reflect.Type specifying which types can be published

**Returns:** Error if channel with the same name already exists

#### DestroyChan
```go
func DestroyChan(name string) error
```
Destroy a channel and terminate all its listeners.

**Parameters:**
- `name`: Name of the channel to destroy

**Returns:** Error if channel doesn't exist

### Publishing Messages

#### Pub
```go
func Pub(channelName string, msg any) error
```
Publish a message to a named channel. Automatically tracks the calling function for debugging.

**Parameters:**
- `channelName`: Name of the target channel
- `msg`: Message data (must match one of the allowed types)

**Returns:** Error if:
- Channel doesn't exist
- Message type is not allowed
- Channel is full
- No subscribers are listening

#### PubWithTag
```go
func PubWithTag(channelName string, msg any, tag string) error
```
Publish a tagged message to a channel. Tags help identify and filter specific messages.

**Parameters:**
- `channelName`: Name of the target channel
- `msg`: Message data
- `tag`: Custom tag for message identification

**Returns:** Same error conditions as `Pub`

### Subscribing to Messages

#### Sub
```go
func Sub(channelName string, timeout time.Duration, listenerFunction func(msg ChanMsg) bool) error
```
Subscribe to a channel with a listener function. Multiple subscribers are supported.

**Parameters:**
- `channelName`: Name of the channel to subscribe to
- `timeout`: Maximum time to wait for listener function to complete
- `listenerFunction`: Callback function that processes messages
  - Returns `true` to unsubscribe, `false` to keep listening
  
**Returns:** Error if channel doesn't exist

**Note:** If the listener function exceeds the timeout, a warning is logged, but the channel continues processing new messages.

### Message Structure

```go
type ChanMsg struct {
    Tag    string  // Optional custom tag
    Number int64   // Sequential message number (starts at 0)
    Name   string  // Channel name
    Data   any     // Message payload
}
```

### Utility Functions

#### SetVerbose
```go
func SetVerbose(verbose bool)
```
Enable or disable debug logging. When enabled, logs channel operations, message flow, and warnings.

## Usage Examples

### Multiple Subscribers

```go
allowedTypes := []reflect.Type{reflect.TypeOf(0)}
chanman.InitChan("multi", 10, allowedTypes)

// First subscriber
chanman.Sub("multi", 300*time.Millisecond, func(msg chanman.ChanMsg) bool {
    fmt.Printf("Subscriber 1: %d\n", msg.Data)
    return false
})

// Second subscriber
chanman.Sub("multi", 300*time.Millisecond, func(msg chanman.ChanMsg) bool {
    fmt.Printf("Subscriber 2: %d\n", msg.Data)
    return false
})

chanman.Pub("multi", 42)
```

### Tagged Messages

```go
allowedTypes := []reflect.Type{reflect.TypeOf("")}
chanman.InitChan("tagged", 5, allowedTypes)

chanman.Sub("tagged", 300*time.Millisecond, func(msg chanman.ChanMsg) bool {
    if msg.Tag == "important" {
        fmt.Printf("Important message: %s\n", msg.Data)
    }
    return false
})

chanman.PubWithTag("tagged", "urgent update", "important")
```

### Message Numbering

```go
allowedTypes := []reflect.Type{reflect.TypeOf("")}
chanman.InitChan("numbered", 10, allowedTypes)

chanman.Sub("numbered", 300*time.Millisecond, func(msg chanman.ChanMsg) bool {
    // Messages are numbered sequentially: 0, 1, 2, ...
    fmt.Printf("Message #%d: %s\n", msg.Number, msg.Data)
    return false
})

chanman.Pub("numbered", "first")  // Number: 0
chanman.Pub("numbered", "second") // Number: 1
chanman.Pub("numbered", "third")  // Number: 2
```

### Listener Timeout Handling

```go
allowedTypes := []reflect.Type{reflect.TypeOf("")}
chanman.InitChan("timeout", 10, allowedTypes)

chanman.SetVerbose(true) // Enable logging to see timeout warnings

// Listener with slow processing
chanman.Sub("timeout", 100*time.Millisecond, func(msg chanman.ChanMsg) bool {
    time.Sleep(300 * time.Millisecond) // Exceeds timeout
    fmt.Printf("Processed: %s\n", msg.Data)
    return false
})

chanman.Pub("timeout", "message 1") // Will timeout but continue processing
chanman.Pub("timeout", "message 2") // Channel remains responsive
```

## Error Handling

All functions return descriptive errors for common issues:

- **Duplicate channel name**: Channel already exists during `InitChan`
- **Channel not found**: Invalid channel name in any operation
- **Type mismatch**: Publishing a message with disallowed type
- **No subscribers**: Publishing to a channel with no listeners
- **Buffer full**: Publishing when channel buffer is at capacity

## Best Practices

1. **Define specific types**: Use precise type definitions for type safety
   ```go
   type MyMessage struct { ID int; Content string }
   allowedTypes := []reflect.Type{reflect.TypeOf(MyMessage{})}
   ```

2. **Set appropriate timeouts**: Balance responsiveness with processing needs
   ```go
   fastTimeout := 50 * time.Millisecond   // For quick operations
   slowTimeout := 5 * time.Second         // For I/O operations
   ```

3. **Enable verbose mode during development**: Helps debug message flow
   ```go
   chanman.SetVerbose(true)
   ```

4. **Always destroy channels**: Clean up resources when done
   ```go
   defer chanman.DestroyChan("mychan")
   ```

5. **Size buffers appropriately**: Prevent message drops during bursts
   ```go
   chanman.InitChan("bursty", 100, types) // Large buffer for high throughput
   ```

## License

See [LICENSE](LICENSE) file for details.

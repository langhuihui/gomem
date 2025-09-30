# GoMem

GoMem is a high-performance memory allocator library for Go, extracted from the Monibuca project.

## Features

- **Multiple Allocation Strategies**: Support for both single-tree and two-tree (AVL) allocation algorithms
- **Buddy Allocator**: Optional buddy system for efficient memory pooling
- **Recyclable Memory**: Memory recycling support with automatic cleanup
- **Scalable Allocator**: Dynamically growing memory allocator
- **Memory Reader**: Efficient multi-buffer reader with zero-copy operations

## Build Tags

The library supports several build tags to customize behavior:

- `twotree`: Use two-tree (AVL) implementation instead of single treap
- `enable_buddy`: Enable buddy allocator for memory pooling
- `disable_rm`: Disable recyclable memory features for reduced overhead

## Installation

```bash
go get github.com/langhuihui/gomem
```

## Usage

### Basic Memory Allocation

```go
package main

import "github.com/langhuihui/gomem"

func main() {
    // Create a scalable memory allocator
    allocator := gomem.NewScalableMemoryAllocator(1024)
    
    // Allocate memory
    buf := allocator.Malloc(256)
    
    // Use the buffer...
    copy(buf, []byte("Hello, World!"))
    
    // Free the memory
    allocator.Free(buf)
}
```

### Recyclable Memory

```go
// Create recyclable memory for batch operations
allocator := gomem.NewScalableMemoryAllocator(1024)
rm := gomem.NewRecyclableMemory(allocator)

// Allocate multiple buffers
buf1 := rm.NextN(128)
buf2 := rm.NextN(256)

// Use the buffers...
copy(buf1, []byte("Buffer 1"))
copy(buf2, []byte("Buffer 2"))

// Recycle all memory at once
rm.Recycle()
```

### Memory Buffer Operations

```go
// Create a memory buffer
mem := gomem.NewMemory([]byte{1, 2, 3, 4, 5})

// Add more data
mem.PushOne([]byte{6, 7, 8})

// Get total size and buffer count
fmt.Printf("Size: %d, Buffers: %d\n", mem.Size, mem.Count())

// Convert to bytes
data := mem.ToBytes()
```

### Memory Reader

```go
// Create a memory reader
reader := gomem.NewReadableBuffersFromBytes([]byte{1, 2, 3}, []byte{4, 5, 6})

// Read data
buf := make([]byte, 6)
n, err := reader.Read(buf)
// buf now contains [1, 2, 3, 4, 5, 6]
```

## Performance Considerations

- Use `enable_buddy` build tag for better memory pooling in high-throughput scenarios
- Use `disable_rm` build tag to reduce overhead when recyclable memory is not needed
- Use `twotree` build tag for more balanced allocation performance

## License

MIT

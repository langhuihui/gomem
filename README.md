# GoMem

<div align="center">
  <img src="logo.png" alt="GoMem Logo" width="200"/>
</div>

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/langhuihui/gomem?style=flat-square)](https://goreportcard.com/report/github.com/langhuihui/gomem)

> **Language**: [English](README.md) | [中文](README_CN.md)

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

### Partial Memory Deallocation

```go
package main

import "github.com/langhuihui/gomem"

func main() {
    // Create a scalable memory allocator
    allocator := gomem.NewScalableMemoryAllocator(1024)
    
    // Allocate a large block of memory
    buf := allocator.Malloc(1024)
    
    // Use different parts of the memory
    part1 := buf[0:256]    // First 256 bytes
    part2 := buf[256:512]  // Middle 256 bytes  
    part3 := buf[512:1024] // Last 512 bytes
    
    // Fill with data
    copy(part1, []byte("Part 1 data"))
    copy(part2, []byte("Part 2 data"))
    copy(part3, []byte("Part 3 data"))
    
    // Partial deallocation - can free parts of memory
    allocator.Free(part1)  // Free first 256 bytes
    allocator.Free(part2)  // Free middle 256 bytes
    
    // Continue using remaining memory
    copy(part3, []byte("Updated part 3"))
    
    // Finally free remaining memory
    allocator.Free(part3)
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

## Concurrency Safety

⚠️ **Important**: Malloc and Free operations must be called from the same goroutine to avoid race conditions. For more elegant usage, consider using [gotask](https://github.com/langhuihui/gotask), where you can allocate memory in the `Start` method and free it in the `Dispose` method.

```go
// ❌ Wrong: Different goroutines
go func() {
    buf := allocator.Malloc(256)
    // ... use buffer
}()

go func() {
    allocator.Free(buf) // Race condition!
}()

// ✅ Correct: Same goroutine
buf := allocator.Malloc(256)
// ... use buffer
allocator.Free(buf)

// ✅ Elegant: Using gotask
type MyTask struct {
    allocator *gomem.ScalableMemoryAllocator
    buffer []byte
}

func (t *MyTask) Start() {
    t.allocator = gomem.NewScalableMemoryAllocator(1024)
    t.buffer = t.allocator.Malloc(256)
}

func (t *MyTask) Dispose() {
    t.allocator.Free(t.buffer)
}
```

## Performance Considerations

- Use `enable_buddy` build tag for better memory pooling in high-throughput scenarios
- **RecyclableMemory enabled is 53% faster** than disabled version and uses less memory
- Use `disable_rm` build tag only when you don't need memory management features (reduces complexity but sacrifices performance)
- **Single-tree allocator is significantly faster** than two-tree allocator (77-86% faster for allocation operations)
- Use `twotree` build tag only if you need faster find operations (100% faster than single-tree)

## Benchmark Results

The following benchmark results were obtained on Apple M2 Pro (ARM64) with Go 1.23.0:

### Single-Tree vs Two-Tree Allocator Performance Comparison

| Operation Type | Single-Tree (ns/op) | Two-Tree (ns/op) | Performance Difference | Winner |
|----------------|-------------------|------------------|----------------------|--------|
| **Basic Allocation** | 12.33 | 22.71 | **84% faster** | Single-Tree |
| **Small Allocation (64B)** | 12.32 | 22.60 | **84% faster** | Single-Tree |
| **Large Allocation (8KB)** | 12.14 | 22.61 | **86% faster** | Single-Tree |
| **Sequential Allocation** | 1961 | 3467 | **77% faster** | Single-Tree |
| **Random Allocation** | 12.47 | 23.02 | **85% faster** | Single-Tree |
| **Find Operation** | 3.03 | 1.51 | **100% faster** | Two-Tree |
| **GetFreeSize** | 3.94 | 4.27 | **8% faster** | Single-Tree |

**Key Findings:**
- Single-tree allocator is **77-86% faster** for memory allocation operations
- Two-tree allocator is **100% faster** for find operations only
- Single-tree allocator is recommended for most use cases due to superior allocation performance

### RecyclableMemory Performance Comparison (RM Enabled vs Disabled)

| Operation Type | RM Enabled (ns/op) | RM Disabled (ns/op) | Performance Difference | Memory Usage |
|----------------|-------------------|---------------------|----------------------|--------------|
| **Basic Operations** | 335.2 | 511.9 | **53% faster** | Enabled: 1536B/2 allocs, Disabled: 1788B/2 allocs |
| **Multiple Allocations** | - | 1035.1 | - | Disabled: 3875B/10 allocs |
| **Clone Operations** | - | 53.7 | - | Disabled: 240B/1 alloc |

**Key Findings:**
- RecyclableMemory enabled is **53% faster** for basic operations
- RM enabled uses less memory (1536B vs 1788B for basic operations)
- RM enabled provides true memory management with recycling capabilities
- RM disabled uses simple `make([]byte, size)` without memory pooling

### Memory Allocator Performance (Single-Tree)

| Benchmark | Operations/sec | Time/op | Memory/op | Allocs/op |
|-----------|----------------|---------|-----------|-----------|
| Allocate | 96,758,520 | 15.08 ns | 0 B | 0 |
| AllocateSmall | 98,864,434 | 12.49 ns | 0 B | 0 |
| AllocateLarge | 100,000,000 | 12.65 ns | 0 B | 0 |
| SequentialAlloc | 1,321,965 | 942.2 ns | 0 B | 0 |
| RandomAlloc | 96,241,566 | 12.79 ns | 0 B | 0 |
| GetFreeSize | 303,367,089 | 3.934 ns | 0 B | 0 |

### Memory Operations Performance

| Benchmark | Operations/sec | Time/op | Memory/op | Allocs/op |
|-----------|----------------|---------|-----------|-----------|
| PushOne | 31,982,593 | 35.05 ns | 143 B | 0 |
| Push | 17,666,751 | 70.40 ns | 259 B | 0 |
| ToBytes | 119,496 | 11,806 ns | 106,496 B | 1 |
| CopyTo | 417,379 | 2,905 ns | 0 B | 0 |
| Append | 979,598 | 1,859 ns | 7,319 B | 0 |
| Count | 1,000,000,000 | 0.3209 ns | 0 B | 0 |
| Range | 32,809,593 | 36.08 ns | 0 B | 0 |

### Memory Reader Performance

| Benchmark | Operations/sec | Time/op | Memory/op | Allocs/op |
|-----------|----------------|---------|-----------|-----------|
| Read | 10,355,643 | 112.4 ns | 112 B | 2 |
| ReadByte | 536,228 | 2,235 ns | 56 B | 2 |
| ReadBytes | 2,556,602 | 608.7 ns | 1,080 B | 18 |
| ReadBE | 408,663 | 3,587 ns | 56 B | 2 |
| Skip | 8,762,934 | 125.8 ns | 56 B | 2 |
| Range | 15,608,808 | 70.99 ns | 80 B | 2 |
| RangeN | 20,101,638 | 79.09 ns | 80 B | 2 |
| LEB128Unmarshal | 356,560 | 3,052 ns | 56 B | 2 |

### Buddy Allocator Performance

| Benchmark | Operations/sec | Time/op | Memory/op | Allocs/op |
|-----------|----------------|---------|-----------|-----------|
| Alloc | 4,017,826 | 388.2 ns | 0 B | 0 |
| AllocSmall | 3,092,535 | 410.7 ns | 0 B | 0 |
| AllocLarge | 3,723,950 | 276.4 ns | 0 B | 0 |
| SequentialAlloc | 62,786 | 17,997 ns | 0 B | 0 |
| RandomAlloc | 3,249,220 | 357.8 ns | 0 B | 0 |
| Pool | 27,800 | 56,846 ns | 196,139 B | 0 |
| NonPowerOf2 | 3,167,425 | 317.8 ns | 0 B | 0 |

### ScalableMemoryAllocator Performance

| Benchmark | Operations/sec | Time/op | Memory/op | Allocs/op |
|-----------|----------------|---------|-----------|-----------|
| **Basic Operations** |
| Malloc | 92,943,320 | 13.22 ns | 0 B | 0 |
| MallocSmall (64B) | 73,196,394 | 16.62 ns | 0 B | 0 |
| MallocLarge (8KB) | 10,000 | 127,506 ns | 4,191,139 B | 5 |
| **Memory Borrowing** |
| Borrow | 221,620,256 | 5.425 ns | 0 B | 0 |
| BorrowSmall (64B) | 90,733,239 | 13.38 ns | 0 B | 0 |
| BorrowLarge (8KB) | 80,812,390 | 12.58 ns | 0 B | 0 |
| **Allocation Patterns** |
| SequentialAlloc | 789,878 | 1,541 ns | 0 B | 0 |
| RandomAlloc | 32,514 | 38,625 ns | 1,197,044 B | 1 |
| RandomBorrow | 144,988,590 | 8.261 ns | 0 B | 0 |
| MixedPattern | 131,418,630 | 9.210 ns | 0 B | 0 |
| **Advanced Operations** |
| GetStats | 1,000,000,000 | 0.3013 ns | 0 B | 0 |
| FreeRest | 52,918,608 | 23.25 ns | 0 B | 0 |
| Scaling | 10,000 | 107,642 ns | 3,351,399 B | 4 |
| Concurrent | 2,332,717 | 519.4 ns | 0 B | 0 |
| MemoryPressure | 10,000 | 145,329 ns | 4,193,342 B | 7 |

### RecyclableMemory Performance

| Benchmark | Operations/sec | Time/op | Memory/op | Allocs/op |
|-----------|----------------|---------|-----------|-----------|
| NextN | 31,148,637 | 32.11 ns | 0 B | 0 |
| BatchRecycle | 3,902,038 | 312.4 ns | 0 B | 0 |
| WithRecycleIndexes | 3,706,173 | 331.5 ns | 0 B | 0 |

### Performance Summary

- **Single-Tree Allocator**: Extremely fast allocation/deallocation with ~12ns per operation and zero memory allocations
- **Two-Tree Allocator**: Slower allocation (~23ns per operation) but faster find operations (~1.5ns vs ~3ns)
- **ScalableMemoryAllocator**: High-performance scalable allocator with dynamic growth
  - **Malloc operations**: ~13-17ns per operation with zero memory allocations
  - **Borrow operations**: Extremely fast ~5-13ns per operation (borrowing is 2-3x faster than malloc)
  - **Memory efficiency**: Zero garbage collection pressure for small/medium allocations
  - **Scaling capability**: Automatically grows to accommodate larger allocations
- **RecyclableMemory**: Efficient batch memory management
  - **NextN operations**: ~32ns per operation with zero memory allocations
  - **Batch recycling**: ~312ns for recycling 10 buffers at once
  - **Memory efficiency**: 53% faster than disabled version with better memory efficiency
- **Memory Operations**: Efficient buffer management with minimal overhead
- **Memory Reader**: High-performance reading with zero-copy operations
- **Buddy Allocator**: Fast power-of-2 allocation with pool support for reduced GC pressure

**Key Performance Insights**:
- **Borrow is fastest**: Borrow operations (5-13ns) are 2-3x faster than malloc operations (13-17ns)
- **Zero GC pressure**: Most operations produce zero memory allocations
- **Excellent scaling**: ScalableMemoryAllocator handles dynamic growth efficiently
- **Batch efficiency**: RecyclableMemory provides efficient batch operations

**Recommendations**: 
- Use **ScalableMemoryAllocator** for applications requiring dynamic memory growth
- Prefer **Borrow** over **Malloc** when possible for maximum performance
- Use **RecyclableMemory** for batch operations requiring multiple allocations
- Use single-tree allocator (default) for most applications due to superior allocation performance
- Keep RecyclableMemory enabled (default) for better performance and memory efficiency
- Only use two-tree allocator if find operations are critical and frequent
- Only use `disable_rm` tag when you don't need memory management features

## License

MIT

---

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

If you have any questions or need help, please open an issue on GitHub.

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=langhuihui/gomem&type=Date)](https://star-history.com/#langhuihui/gomem&Date)

---

<div align="center">
  <sub>Built with ❤️ by the GoMem team</sub>
</div>

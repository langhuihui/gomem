package gomem

import (
	"fmt"
	"testing"
)

func ExampleNewScalableMemoryAllocator() {
	// Create a scalable memory allocator
	allocator := NewScalableMemoryAllocator(1024)

	// Allocate memory
	buf1 := allocator.Malloc(256)
	buf2 := allocator.Malloc(512)

	fmt.Printf("Allocated %d bytes\n", len(buf1))
	fmt.Printf("Allocated %d bytes\n", len(buf2))

	// Use the buffers
	copy(buf1, []byte("Hello, World!"))
	copy(buf2, []byte("This is a test buffer"))

	// Free the memory
	allocator.Free(buf1)
	allocator.Free(buf2)

	// Output:
	// Allocated 256 bytes
	// Allocated 512 bytes
}

func ExampleNewRecyclableMemory() {
	// Create a scalable memory allocator
	allocator := NewScalableMemoryAllocator(1024)

	// Create recyclable memory
	rm := NewRecyclableMemory(allocator)

	// Allocate some memory
	buf1 := rm.NextN(128)
	buf2 := rm.NextN(256)

	fmt.Printf("Allocated %d bytes\n", len(buf1))
	fmt.Printf("Allocated %d bytes\n", len(buf2))

	// Use the buffers
	copy(buf1, []byte("Buffer 1"))
	copy(buf2, []byte("Buffer 2"))

	// Recycle all memory at once
	rm.Recycle()

	// Output:
	// Allocated 128 bytes
	// Allocated 256 bytes
}

func ExampleMemory() {
	// Create a new memory buffer
	mem := NewMemory([]byte{1, 2, 3, 4, 5})

	// Add more data
	mem.PushOne([]byte{6, 7, 8})

	fmt.Printf("Total size: %d\n", mem.Size)
	fmt.Printf("Number of buffers: %d\n", mem.Count())

	// Convert to bytes
	data := mem.ToBytes()
	fmt.Printf("Data: %v\n", data)

	// Output:
	// Total size: 8
	// Number of buffers: 2
	// Data: [1 2 3 4 5 6 7 8]
}

func TestExamples(t *testing.T) {
	ExampleNewScalableMemoryAllocator()
	ExampleNewRecyclableMemory()
	ExampleMemory()
}

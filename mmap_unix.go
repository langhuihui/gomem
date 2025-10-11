//go:build (linux || darwin) && enable_mmap

package gomem

import (
	"fmt"
	"syscall"
	"unsafe"
)

func createMemoryAllocator(size int) *MemoryAllocator {
	// Allocate anonymous memory using mmap
	// PROT_READ | PROT_WRITE: readable and writable
	// MAP_ANON | MAP_PRIVATE: anonymous private mapping
	memory, err := syscall.Mmap(
		-1, // use -1 for anonymous mapping
		0,  // offset
		size,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_ANON|syscall.MAP_PRIVATE,
	)
	if err != nil {
		panic(fmt.Sprintf("mmap failed: %v", err))
	}

	start := int64(uintptr(unsafe.Pointer(&memory[0])))
	ret := &MemoryAllocator{
		allocator: NewAllocator(size),
		Size:      size,
		memory:    memory,
		start:     start,
		recycle: func() {
			// Release the mmap allocated memory
			if err := syscall.Munmap(memory); err != nil {
				panic(fmt.Sprintf("munmap failed: %v", err))
			}
		},
	}
	ret.allocator.Init(size)
	return ret
}

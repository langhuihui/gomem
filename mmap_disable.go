//go:build !enable_mmap
package gomem

import "unsafe"

func createMemoryAllocator(size int) *MemoryAllocator {
	memory := make([]byte, size)
	ret := &MemoryAllocator{
		allocator: NewAllocator(size),
		Size:      size,
		memory:    memory,
		start:     int64(uintptr(unsafe.Pointer(&memory[0]))),
	}
	ret.allocator.Init(size)
	return ret
}
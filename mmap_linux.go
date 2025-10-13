//go:build linux && enable_mmap

package gomem

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	// MADV_HUGEPAGE advises the kernel to use transparent huge pages for this memory region
	MADV_HUGEPAGE = 14
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

	// Enable Transparent Huge Pages (THP) for better performance
	// This advises the kernel to use huge pages (typically 2MB on x86_64) instead of 4KB pages
	// which can significantly reduce TLB misses and improve memory access performance
	err = madvise(memory, MADV_HUGEPAGE)
	if err != nil {
		// Don't panic if madvise fails, as THP might not be available
		// Just log the error and continue with regular pages
		// In production, you might want to log this properly
		_ = err // Silent failure - THP is a performance optimization, not required
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

// madvise provides hints to the kernel about memory usage patterns
func madvise(b []byte, advice int) error {
	if len(b) == 0 {
		return nil
	}
	_, _, errno := syscall.Syscall(
		syscall.SYS_MADVISE,
		uintptr(unsafe.Pointer(&b[0])),
		uintptr(len(b)),
		uintptr(advice),
	)
	if errno != 0 {
		return errno
	}
	return nil
}

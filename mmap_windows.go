//go:build windows && amd64 && enable_mmap

package gomem

import (
	"fmt"
	"syscall"
	"unsafe"
)

func createMemoryAllocator(size int) *MemoryAllocator {
	// Create anonymous memory mapping using CreateFileMapping (similar to Unix MAP_ANON)
	// INVALID_HANDLE_VALUE (-1) indicates anonymous mapping, not file-backed
	low, high := uint32(size), uint32(size>>32)
	fmap, err := syscall.CreateFileMapping(
		syscall.InvalidHandle,  // hFile: INVALID_HANDLE_VALUE for anonymous mapping
		nil,                    // lpFileMappingAttributes: default security attributes
		syscall.PAGE_READWRITE, // flProtect: readable and writable
		high,                   // dwMaximumSizeHigh: high 32 bits of size
		low,                    // dwMaximumSizeLow: low 32 bits of size
		nil,                    // lpName: anonymous mapping
	)
	if err != nil {
		panic(fmt.Sprintf("CreateFileMapping failed: %v", err))
	}

	// Map the file mapping object into the process address space
	ptr, err := syscall.MapViewOfFile(
		fmap,                   // hFileMappingObject
		syscall.FILE_MAP_WRITE, // dwDesiredAccess: readable and writable
		0,                      // dwFileOffsetHigh
		0,                      // dwFileOffsetLow
		uintptr(size),          // dwNumberOfBytesToMap
	)
	if err != nil {
		syscall.CloseHandle(fmap)
		panic(fmt.Sprintf("MapViewOfFile failed: %v", err))
	}

	// Convert pointer to []byte slice
	memory := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), size)
	start := int64(ptr)

	ret := &MemoryAllocator{
		allocator: NewAllocator(size),
		Size:      size,
		memory:    memory,
		start:     start,
		recycle: func() {
			// Unmap the view
			if err := syscall.UnmapViewOfFile(ptr); err != nil {
				panic(fmt.Sprintf("UnmapViewOfFile failed: %v", err))
			}
			// Close the file mapping handle
			syscall.CloseHandle(fmap)
		},
	}
	ret.allocator.Init(size)
	return ret
}

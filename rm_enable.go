//go:build !disable_rm

/*
Package gomem provides efficient memory management and buffer operations for Go applications.
This file contains the RecyclableMemory and memory allocator implementations, which provide
advanced memory management capabilities with automatic recycling and scalable allocation.

RecyclableMemory extends the basic Memory functionality with automatic memory recycling,
allowing for efficient reuse of allocated memory blocks. The memory allocators provide
different strategies for memory allocation, from simple fixed-size allocators to scalable
multi-tier allocators that can grow dynamically based on demand.

Key features:
- RecyclableMemory with automatic memory recycling
- ScalableMemoryAllocator for dynamic memory pool management
- MemoryAllocator for fixed-size memory pools
- Support for borrowing memory (temporary allocation)
- Automatic cleanup and memory reuse
- Integration with io.Reader for efficient data reading
- Memory statistics and monitoring capabilities

Example usage:

	// Create a scalable memory allocator
	allocator := NewScalableMemoryAllocator(1024 * 1024) // 1MB initial size

	// Create recyclable memory
	rm := NewRecyclableMemory(allocator)
	rm.InitRecycleIndexes(10) // Reserve space for 10 allocations

	// Allocate memory
	data := rm.NextN(1024) // Allocate 1KB
	copy(data, []byte("Hello, World!"))

	// Allocate more memory
	moreData := rm.NextN(512) // Allocate 512 bytes

	// Use the memory...

	// Recycle all allocated memory back to the allocator
	rm.Recycle()

	// The memory is now available for reuse
*/
package gomem

import (
	"fmt"
	"io"
	"slices"
	"unsafe"
)

type RecyclableMemory struct {
	allocator *ScalableMemoryAllocator
	Memory
	recycleIndexes []int
}

func NewRecyclableMemory(allocator *ScalableMemoryAllocator) RecyclableMemory {
	return RecyclableMemory{allocator: allocator}
}

// InitRecycleIndexes init the recycle indexes, if not init, Recycle will free all the memory
func (r *RecyclableMemory) InitRecycleIndexes(max int) {
	if r.recycleIndexes == nil {
		r.recycleIndexes = make([]int, 0, max)
	}
}

func (r *RecyclableMemory) GetAllocator() *ScalableMemoryAllocator {
	return r.allocator
}

// NextN allocate size bytes and add the index to the recycle indexes
func (r *RecyclableMemory) NextN(size int) (memory []byte) {
	memory = r.allocator.Malloc(size)
	if r.recycleIndexes != nil {
		r.recycleIndexes = append(r.recycleIndexes, r.Count())
	}
	r.PushOne(memory)
	return
}

// AddRecycleBytes add the bytes will be recycled
func (r *RecyclableMemory) AddRecycleBytes(b []byte) {
	if r.recycleIndexes != nil {
		r.recycleIndexes = append(r.recycleIndexes, r.Count())
	}
	r.PushOne(b)
}

func (r *RecyclableMemory) SetAllocator(allocator *ScalableMemoryAllocator) {
	r.allocator = allocator
}

func (r *RecyclableMemory) Recycle() {
	if r.recycleIndexes != nil {
		for _, index := range r.recycleIndexes {
			r.allocator.Free(r.Buffers[index])
		}
		r.recycleIndexes = r.recycleIndexes[:0]
	} else {
		for _, buf := range r.Buffers {
			r.allocator.Free(buf)
		}
	}
	r.Reset()
}

type MemoryAllocator struct {
	allocator *Allocator
	start     int64
	memory    []byte
	Size      int
	recycle   func()
}

func (ma *MemoryAllocator) Recycle() {
	ma.allocator.Recycle()
	if ma.recycle != nil {
		ma.recycle()
	}
}

func (ma *MemoryAllocator) Find(size int) (memory []byte) {
	if offset := ma.allocator.Find(size); offset != -1 {
		memory = ma.memory[offset : offset+size]
	}
	return
}

func (ma *MemoryAllocator) Malloc(size int) (memory []byte) {
	if offset := ma.allocator.Allocate(size); offset != -1 {
		memory = ma.memory[offset : offset+size]
	}
	return
}

func (ma *MemoryAllocator) free(start, size int) (ret bool) {
	if start < 0 || start+size > ma.Size {
		return
	}
	ma.allocator.Free(start, size)
	return true
}

// GetBlocks return the blocks of the allocator
func (ma *MemoryAllocator) GetBlocks() (blocks []*Block) {
	return ma.allocator.GetBlocks()
}

type ScalableMemoryAllocator struct {
	children    []*MemoryAllocator
	totalMalloc int64
	totalFree   int64
	size        int
	childSize   int
}

func NewScalableMemoryAllocator(size int) (ret *ScalableMemoryAllocator) {
	return &ScalableMemoryAllocator{children: []*MemoryAllocator{GetMemoryAllocator(size)}, size: size, childSize: size}
}

func (sma *ScalableMemoryAllocator) checkSize() {
	var totalFree int
	for _, child := range sma.children {
		totalFree += child.allocator.GetFreeSize()
	}
	if inUse := sma.totalMalloc - sma.totalFree; totalFree != sma.size-int(inUse) {
		panic("CheckSize")
	} else {
		if inUse > 3000000 {
			fmt.Println(uintptr(unsafe.Pointer(sma)), inUse)
		}
	}
}

func (sma *ScalableMemoryAllocator) addMallocCount(size int) {
	sma.totalMalloc += int64(size)
}

func (sma *ScalableMemoryAllocator) addFreeCount(size int) {
	sma.totalFree += int64(size)
}

func (sma *ScalableMemoryAllocator) GetTotalMalloc() int64 {
	return sma.totalMalloc
}

func (sma *ScalableMemoryAllocator) GetTotalFree() int64 {
	return sma.totalFree
}

func (sma *ScalableMemoryAllocator) GetChildren() []*MemoryAllocator {
	return sma.children
}

func (sma *ScalableMemoryAllocator) Recycle() {
	for _, child := range sma.children {
		child.Recycle()
	}
	sma.children = nil
}

// Borrow = Malloc + Free = Find, must use the memory at once
func (sma *ScalableMemoryAllocator) Borrow(size int) (memory []byte) {
	if sma == nil || size > MaxBlockSize {
		return
	}
	defer sma.addMallocCount(size)
	var child *MemoryAllocator
	for _, child = range sma.children {
		if memory = child.Find(size); memory != nil {
			return
		}
	}
	for sma.childSize < MaxBlockSize {
		sma.childSize = sma.childSize << 1
		if sma.childSize >= size {
			break
		}
	}
	child = GetMemoryAllocator(sma.childSize)
	sma.size += child.Size
	memory = child.Find(size)
	sma.children = append(sma.children, child)
	return
}

func (sma *ScalableMemoryAllocator) Malloc(size int) (memory []byte) {
	if sma == nil || size > MaxBlockSize {
		return make([]byte, size)
	}
	defer sma.addMallocCount(size)
	var child *MemoryAllocator
	for _, child = range sma.children {
		if memory = child.Malloc(size); memory != nil {
			return
		}
	}
	for sma.childSize < MaxBlockSize {
		sma.childSize = sma.childSize << 1
		if sma.childSize >= size {
			break
		}
	}
	child = GetMemoryAllocator(sma.childSize)
	sma.size += child.Size
	memory = child.Malloc(size)
	sma.children = append(sma.children, child)
	return
}

func (sma *ScalableMemoryAllocator) GetAllocator() *ScalableMemoryAllocator {
	return sma
}

func (sma *ScalableMemoryAllocator) Read(reader io.Reader, n int) (mem []byte, err error) {
	mem = sma.Malloc(n)
	meml := n
	if n, err = reader.Read(mem); err == nil {
		if n < meml {
			sma.Free(mem[n:])
			mem = mem[:n]
		}
	} else {
		sma.Free(mem)
	}
	return
}

// FreeRest keep the first keep bytes and free the rest
func (sma *ScalableMemoryAllocator) FreeRest(mem *[]byte, keep int) {
	if m := *mem; keep < len(m) {
		sma.Free(m[keep:])
		*mem = m[:keep]
	}
}

func (sma *ScalableMemoryAllocator) Free(mem []byte) bool {
	if sma == nil {
		return false
	}
	ptr := int64(uintptr(unsafe.Pointer(&mem[0])))
	size := len(mem)
	for i, child := range sma.children {
		if start := int(ptr - child.start); start >= 0 && start < child.Size && child.free(start, size) {
			sma.addFreeCount(size)
			if len(sma.children) > 1 && child.allocator.sizeTree.End-child.allocator.sizeTree.Start == child.Size {
				child.Recycle()
				sma.children = slices.Delete(sma.children, i, i+1)
				sma.size -= child.Size
			}
			return true
		}
	}
	return false
}

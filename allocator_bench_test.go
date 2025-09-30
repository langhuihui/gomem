package gomem

import (
	"testing"
)

// BenchmarkAllocatorAllocate benchmarks memory allocation
func BenchmarkAllocatorAllocate(b *testing.B) {
	allocator := NewAllocator(1024 * 64) // 64KB pool

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		offset := allocator.Allocate(1024)
		if offset == -1 {
			b.Fatal("Failed to allocate memory")
		}
		allocator.Free(offset, 1024)
	}
}

// BenchmarkAllocatorAllocateSmall benchmarks small memory allocation
func BenchmarkAllocatorAllocateSmall(b *testing.B) {
	allocator := NewAllocator(1024 * 64) // 64KB pool

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		offset := allocator.Allocate(64)
		if offset == -1 {
			b.Fatal("Failed to allocate memory")
		}
		allocator.Free(offset, 64)
	}
}

// BenchmarkAllocatorAllocateLarge benchmarks large memory allocation
func BenchmarkAllocatorAllocateLarge(b *testing.B) {
	allocator := NewAllocator(1024 * 64) // 64KB pool

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		offset := allocator.Allocate(8192)
		if offset == -1 {
			b.Fatal("Failed to allocate memory")
		}
		allocator.Free(offset, 8192)
	}
}

// BenchmarkAllocatorSequentialAlloc benchmarks sequential allocation and deallocation
func BenchmarkAllocatorSequentialAlloc(b *testing.B) {
	allocator := NewAllocator(1024 * 64) // 64KB pool
	allocations := make([]int, 50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Allocate 50 blocks
		for j := 0; j < 50; j++ {
			offset := allocator.Allocate(1024)
			if offset == -1 {
				b.Fatal("Failed to allocate memory")
			}
			allocations[j] = offset
		}

		// Free all blocks
		for j := 0; j < 50; j++ {
			allocator.Free(allocations[j], 1024)
		}
	}
}

// BenchmarkAllocatorRandomAlloc benchmarks random allocation pattern
func BenchmarkAllocatorRandomAlloc(b *testing.B) {
	allocator := NewAllocator(1024 * 64) // 64KB pool
	sizes := []int{64, 128, 256, 512, 1024, 2048, 4096}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		size := sizes[i%len(sizes)]
		offset := allocator.Allocate(size)
		if offset == -1 {
			b.Fatal("Failed to allocate memory")
		}
		allocator.Free(offset, size)
	}
}

// BenchmarkAllocatorGetFreeSize benchmarks getting free size
func BenchmarkAllocatorGetFreeSize(b *testing.B) {
	allocator := NewAllocator(1024 * 64) // 64KB pool

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = allocator.GetFreeSize()
	}
}

package gomem

import (
	"testing"
)

// BenchmarkTwoTreeAllocator benchmarks the two-tree allocator
func BenchmarkTwoTreeAllocator(b *testing.B) {
	allocator := NewAllocator(1024 * 1024) // 1MB pool

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		offset := allocator.Allocate(1024)
		if offset == -1 {
			b.Fatal("Failed to allocate memory")
		}
		allocator.Free(offset, 1024)
	}
}

// BenchmarkSingleTreeAllocator benchmarks the single-tree allocator
func BenchmarkSingleTreeAllocator(b *testing.B) {
	allocator := NewAllocator(1024 * 1024) // 1MB pool

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		offset := allocator.Allocate(1024)
		if offset == -1 {
			b.Fatal("Failed to allocate memory")
		}
		allocator.Free(offset, 1024)
	}
}

// BenchmarkTwoTreeSmallAlloc benchmarks small allocations with two-tree
func BenchmarkTwoTreeSmallAlloc(b *testing.B) {
	allocator := NewAllocator(1024 * 1024) // 1MB pool

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		offset := allocator.Allocate(64)
		if offset == -1 {
			b.Fatal("Failed to allocate memory")
		}
		allocator.Free(offset, 64)
	}
}

// BenchmarkSingleTreeSmallAlloc benchmarks small allocations with single-tree
func BenchmarkSingleTreeSmallAlloc(b *testing.B) {
	allocator := NewAllocator(1024 * 1024) // 1MB pool

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		offset := allocator.Allocate(64)
		if offset == -1 {
			b.Fatal("Failed to allocate memory")
		}
		allocator.Free(offset, 64)
	}
}

// BenchmarkTwoTreeLargeAlloc benchmarks large allocations with two-tree
func BenchmarkTwoTreeLargeAlloc(b *testing.B) {
	allocator := NewAllocator(1024 * 1024) // 1MB pool

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		offset := allocator.Allocate(8192)
		if offset == -1 {
			b.Fatal("Failed to allocate memory")
		}
		allocator.Free(offset, 8192)
	}
}

// BenchmarkSingleTreeLargeAlloc benchmarks large allocations with single-tree
func BenchmarkSingleTreeLargeAlloc(b *testing.B) {
	allocator := NewAllocator(1024 * 1024) // 1MB pool

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		offset := allocator.Allocate(8192)
		if offset == -1 {
			b.Fatal("Failed to allocate memory")
		}
		allocator.Free(offset, 8192)
	}
}

// BenchmarkTwoTreeSequential benchmarks sequential allocation pattern with two-tree
func BenchmarkTwoTreeSequential(b *testing.B) {
	allocator := NewAllocator(1024 * 1024) // 1MB pool
	allocations := make([]int, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Allocate 100 blocks
		for j := 0; j < 100; j++ {
			offset := allocator.Allocate(1024)
			if offset == -1 {
				b.Fatal("Failed to allocate memory")
			}
			allocations[j] = offset
		}

		// Free all blocks
		for j := 0; j < 100; j++ {
			allocator.Free(allocations[j], 1024)
		}
	}
}

// BenchmarkSingleTreeSequential benchmarks sequential allocation pattern with single-tree
func BenchmarkSingleTreeSequential(b *testing.B) {
	allocator := NewAllocator(1024 * 1024) // 1MB pool
	allocations := make([]int, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Allocate 100 blocks
		for j := 0; j < 100; j++ {
			offset := allocator.Allocate(1024)
			if offset == -1 {
				b.Fatal("Failed to allocate memory")
			}
			allocations[j] = offset
		}

		// Free all blocks
		for j := 0; j < 100; j++ {
			allocator.Free(allocations[j], 1024)
		}
	}
}

// BenchmarkTwoTreeRandom benchmarks random allocation pattern with two-tree
func BenchmarkTwoTreeRandom(b *testing.B) {
	allocator := NewAllocator(1024 * 1024) // 1MB pool
	sizes := []int{64, 128, 256, 512, 1024, 2048, 4096, 8192}

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

// BenchmarkSingleTreeRandom benchmarks random allocation pattern with single-tree
func BenchmarkSingleTreeRandom(b *testing.B) {
	allocator := NewAllocator(1024 * 1024) // 1MB pool
	sizes := []int{64, 128, 256, 512, 1024, 2048, 4096, 8192}

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

// BenchmarkTwoTreeFind benchmarks find operation with two-tree
func BenchmarkTwoTreeFind(b *testing.B) {
	allocator := NewAllocator(1024 * 1024) // 1MB pool

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = allocator.Find(1024)
	}
}

// BenchmarkSingleTreeFind benchmarks find operation with single-tree
func BenchmarkSingleTreeFind(b *testing.B) {
	allocator := NewAllocator(1024 * 1024) // 1MB pool

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = allocator.Find(1024)
	}
}

// BenchmarkTwoTreeGetFreeSize benchmarks GetFreeSize with two-tree
func BenchmarkTwoTreeGetFreeSize(b *testing.B) {
	allocator := NewAllocator(1024 * 1024) // 1MB pool

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = allocator.GetFreeSize()
	}
}

// BenchmarkSingleTreeGetFreeSize benchmarks GetFreeSize with single-tree
func BenchmarkSingleTreeGetFreeSize(b *testing.B) {
	allocator := NewAllocator(1024 * 1024) // 1MB pool

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = allocator.GetFreeSize()
	}
}

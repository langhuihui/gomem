package gomem

import (
	"testing"
)

// BenchmarkScalableMemoryAllocatorMalloc benchmarks basic malloc operations
func BenchmarkScalableMemoryAllocatorMalloc(b *testing.B) {
	allocator := NewScalableMemoryAllocator(1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mem := allocator.Malloc(1024)
		if mem == nil {
			b.Fatal("Failed to allocate memory")
		}
		allocator.Free(mem)
	}
}

// BenchmarkScalableMemoryAllocatorMallocSmall benchmarks small memory allocation
func BenchmarkScalableMemoryAllocatorMallocSmall(b *testing.B) {
	allocator := NewScalableMemoryAllocator(1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mem := allocator.Malloc(64)
		if mem == nil {
			b.Fatal("Failed to allocate memory")
		}
		allocator.Free(mem)
	}
}

// BenchmarkScalableMemoryAllocatorMallocLarge benchmarks large memory allocation
func BenchmarkScalableMemoryAllocatorMallocLarge(b *testing.B) {
	allocator := NewScalableMemoryAllocator(1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mem := allocator.Malloc(8192)
		if mem == nil {
			b.Fatal("Failed to allocate memory")
		}
		allocator.Free(mem)
	}
}

// BenchmarkScalableMemoryAllocatorBorrow benchmarks borrow operations
func BenchmarkScalableMemoryAllocatorBorrow(b *testing.B) {
	allocator := NewScalableMemoryAllocator(1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mem := allocator.Borrow(1024)
		if mem == nil {
			b.Fatal("Failed to borrow memory")
		}
		// Borrowed memory is automatically freed when not used
	}
}

// BenchmarkScalableMemoryAllocatorBorrowSmall benchmarks small memory borrowing
func BenchmarkScalableMemoryAllocatorBorrowSmall(b *testing.B) {
	allocator := NewScalableMemoryAllocator(1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mem := allocator.Borrow(64)
		if mem == nil {
			b.Fatal("Failed to borrow memory")
		}
	}
}

// BenchmarkScalableMemoryAllocatorBorrowLarge benchmarks large memory borrowing
func BenchmarkScalableMemoryAllocatorBorrowLarge(b *testing.B) {
	allocator := NewScalableMemoryAllocator(1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mem := allocator.Borrow(8192)
		if mem == nil {
			b.Fatal("Failed to borrow memory")
		}
	}
}

// BenchmarkScalableMemoryAllocatorSequentialAlloc benchmarks sequential allocation pattern
func BenchmarkScalableMemoryAllocatorSequentialAlloc(b *testing.B) {
	allocator := NewScalableMemoryAllocator(1024)
	allocations := make([][]byte, 50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Allocate 50 blocks
		for j := 0; j < 50; j++ {
			mem := allocator.Malloc(1024)
			if mem == nil {
				b.Fatal("Failed to allocate memory")
			}
			allocations[j] = mem
		}

		// Free all blocks
		for j := 0; j < 50; j++ {
			allocator.Free(allocations[j])
		}
	}
}

// BenchmarkScalableMemoryAllocatorRandomAlloc benchmarks random allocation pattern
func BenchmarkScalableMemoryAllocatorRandomAlloc(b *testing.B) {
	allocator := NewScalableMemoryAllocator(1024)
	sizes := []int{64, 128, 256, 512, 1024, 2048, 4096}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		size := sizes[i%len(sizes)]
		mem := allocator.Malloc(size)
		if mem == nil {
			b.Fatal("Failed to allocate memory")
		}
		allocator.Free(mem)
	}
}

// BenchmarkScalableMemoryAllocatorRandomBorrow benchmarks random borrow pattern
func BenchmarkScalableMemoryAllocatorRandomBorrow(b *testing.B) {
	allocator := NewScalableMemoryAllocator(1024)
	sizes := []int{64, 128, 256, 512, 1024, 2048, 4096}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		size := sizes[i%len(sizes)]
		mem := allocator.Borrow(size)
		if mem == nil {
			b.Fatal("Failed to borrow memory")
		}
	}
}

// BenchmarkScalableMemoryAllocatorMixedPattern benchmarks mixed malloc/borrow pattern
func BenchmarkScalableMemoryAllocatorMixedPattern(b *testing.B) {
	allocator := NewScalableMemoryAllocator(1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Mix of malloc and borrow operations
		if i%2 == 0 {
			mem := allocator.Malloc(1024)
			if mem == nil {
				b.Fatal("Failed to allocate memory")
			}
			allocator.Free(mem)
		} else {
			mem := allocator.Borrow(1024)
			if mem == nil {
				b.Fatal("Failed to borrow memory")
			}
		}
	}
}

// BenchmarkScalableMemoryAllocatorGetStats benchmarks getting statistics
func BenchmarkScalableMemoryAllocatorGetStats(b *testing.B) {
	allocator := NewScalableMemoryAllocator(1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = allocator.GetTotalMalloc()
		_ = allocator.GetTotalFree()
		_ = len(allocator.GetChildren())
	}
}

// BenchmarkRecyclableMemoryNextN benchmarks RecyclableMemory NextN operation
func BenchmarkRecyclableMemoryNextN(b *testing.B) {
	allocator := NewScalableMemoryAllocator(1024)
	rm := NewRecyclableMemory(allocator)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rm.NextN(1024)
		if i%100 == 0 {
			rm.Recycle()
		}
	}
}

// BenchmarkRecyclableMemoryBatchRecycle benchmarks batch recycling
func BenchmarkRecyclableMemoryBatchRecycle(b *testing.B) {
	allocator := NewScalableMemoryAllocator(1024)
	rm := NewRecyclableMemory(allocator)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Allocate multiple buffers
		for j := 0; j < 10; j++ {
			rm.NextN(1024)
		}
		// Recycle all at once
		rm.Recycle()
	}
}

// BenchmarkRecyclableMemoryWithRecycleIndexes benchmarks with recycle indexes
func BenchmarkRecyclableMemoryWithRecycleIndexes(b *testing.B) {
	allocator := NewScalableMemoryAllocator(1024)
	rm := NewRecyclableMemory(allocator)
	rm.InitRecycleIndexes(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Allocate multiple buffers
		for j := 0; j < 10; j++ {
			rm.NextN(1024)
		}
		// Recycle all at once
		rm.Recycle()
	}
}

// BenchmarkScalableMemoryAllocatorFreeRest benchmarks FreeRest operation
func BenchmarkScalableMemoryAllocatorFreeRest(b *testing.B) {
	allocator := NewScalableMemoryAllocator(1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mem := allocator.Malloc(1024)
		if mem == nil {
			b.Fatal("Failed to allocate memory")
		}
		// Keep only first 256 bytes, free the rest
		allocator.FreeRest(&mem, 256)
		allocator.Free(mem)
	}
}

// BenchmarkScalableMemoryAllocatorScaling benchmarks allocator scaling behavior
func BenchmarkScalableMemoryAllocatorScaling(b *testing.B) {
	allocator := NewScalableMemoryAllocator(1024)
	sizes := []int{1024, 2048, 4096, 8192, 16384}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		size := sizes[i%len(sizes)]
		mem := allocator.Malloc(size)
		if mem == nil {
			b.Fatal("Failed to allocate memory")
		}
		allocator.Free(mem)
	}
}

// BenchmarkScalableMemoryAllocatorConcurrent benchmarks concurrent-like allocation pattern
func BenchmarkScalableMemoryAllocatorConcurrent(b *testing.B) {
	allocator := NewScalableMemoryAllocator(1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate concurrent allocation pattern
		buffers := make([][]byte, 20)

		// Allocate multiple buffers of different sizes
		for j := 0; j < 20; j++ {
			size := 64 * (j + 1) // Different sizes
			mem := allocator.Malloc(size)
			if mem == nil {
				b.Fatal("Failed to allocate memory")
			}
			buffers[j] = mem
		}

		// Free all buffers
		for j := 0; j < 20; j++ {
			allocator.Free(buffers[j])
		}
	}
}

// BenchmarkScalableMemoryAllocatorMemoryPressure benchmarks under memory pressure
func BenchmarkScalableMemoryAllocatorMemoryPressure(b *testing.B) {
	allocator := NewScalableMemoryAllocator(1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create memory pressure by allocating large blocks
		buffers := make([][]byte, 0, 100)

		for j := 0; j < 100; j++ {
			size := 1024 + j*64 // Increasing sizes
			mem := allocator.Malloc(size)
			if mem != nil {
				buffers = append(buffers, mem)
			}
			// Free some buffers to create fragmentation
			if j%3 == 0 && len(buffers) > 0 {
				allocator.Free(buffers[0])
				buffers = buffers[1:]
			}
		}

		// Free remaining buffers
		for _, buf := range buffers {
			allocator.Free(buf)
		}
	}
}

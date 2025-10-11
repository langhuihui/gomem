package gomem

import (
	"testing"
	"unsafe"
)

// BenchmarkMemoryAllocatorMalloc benchmarks memory allocation
func BenchmarkMemoryAllocatorMalloc(b *testing.B) {
	ma := createMemoryAllocator(1024 * 1024) // 1MB pool
	defer ma.Recycle()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mem := ma.Malloc(1024)
		if mem == nil {
			b.Fatal("Failed to allocate memory")
		}
		ma.free(int(int64(uintptr(unsafe.Pointer(&mem[0])))-ma.start), 1024)
	}
}

// BenchmarkMemoryAllocatorMallocSmall benchmarks small memory allocation
func BenchmarkMemoryAllocatorMallocSmall(b *testing.B) {
	ma := createMemoryAllocator(1024 * 1024) // 1MB pool
	defer ma.Recycle()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mem := ma.Malloc(64)
		if mem == nil {
			b.Fatal("Failed to allocate memory")
		}
		ma.free(int(int64(uintptr(unsafe.Pointer(&mem[0])))-ma.start), 64)
	}
}

// BenchmarkMemoryAllocatorMallocLarge benchmarks large memory allocation
func BenchmarkMemoryAllocatorMallocLarge(b *testing.B) {
	ma := createMemoryAllocator(1024 * 1024) // 1MB pool
	defer ma.Recycle()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mem := ma.Malloc(8192)
		if mem == nil {
			b.Fatal("Failed to allocate memory")
		}
		ma.free(int(int64(uintptr(unsafe.Pointer(&mem[0])))-ma.start), 8192)
	}
}

// BenchmarkMemoryAllocatorSequential benchmarks sequential allocation pattern
func BenchmarkMemoryAllocatorSequential(b *testing.B) {
	ma := createMemoryAllocator(1024 * 1024) // 1MB pool
	defer ma.Recycle()
	allocations := make([][]byte, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Allocate 100 blocks
		for j := 0; j < 100; j++ {
			mem := ma.Malloc(1024)
			if mem == nil {
				b.Fatal("Failed to allocate memory")
			}
			allocations[j] = mem
		}

		// Free all blocks
		for j := 0; j < 100; j++ {
			ma.free(int(int64(uintptr(unsafe.Pointer(&allocations[j][0])))-ma.start), 1024)
		}
	}
}

// BenchmarkMemoryAllocatorRandom benchmarks random allocation pattern
func BenchmarkMemoryAllocatorRandom(b *testing.B) {
	ma := createMemoryAllocator(1024 * 1024) // 1MB pool
	defer ma.Recycle()
	sizes := []int{64, 128, 256, 512, 1024, 2048, 4096}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		size := sizes[i%len(sizes)]
		mem := ma.Malloc(size)
		if mem == nil {
			b.Fatal("Failed to allocate memory")
		}
		ma.free(int(int64(uintptr(unsafe.Pointer(&mem[0])))-ma.start), size)
	}
}

// BenchmarkMemoryAllocatorFind benchmarks find operation
func BenchmarkMemoryAllocatorFind(b *testing.B) {
	ma := createMemoryAllocator(1024 * 1024) // 1MB pool
	defer ma.Recycle()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ma.Find(1024)
	}
}

// BenchmarkMemoryAllocatorCreation benchmarks memory allocator creation
func BenchmarkMemoryAllocatorCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ma := createMemoryAllocator(1024 * 1024) // 1MB pool
		ma.Recycle()
	}
}

// BenchmarkMemoryAllocatorCreationSmall benchmarks small memory allocator creation
func BenchmarkMemoryAllocatorCreationSmall(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ma := createMemoryAllocator(16384) // 16KB pool
		ma.Recycle()
	}
}

// BenchmarkMemoryAllocatorCreationLarge benchmarks large memory allocator creation
func BenchmarkMemoryAllocatorCreationLarge(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ma := createMemoryAllocator(16 * 1024 * 1024) // 16MB pool
		ma.Recycle()
	}
}

// BenchmarkMemoryAllocatorWriteAccess benchmarks memory write access
func BenchmarkMemoryAllocatorWriteAccess(b *testing.B) {
	ma := createMemoryAllocator(1024 * 1024) // 1MB pool
	defer ma.Recycle()
	mem := ma.Malloc(1024)
	if mem == nil {
		b.Fatal("Failed to allocate memory")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(mem); j++ {
			mem[j] = byte(j % 256)
		}
	}
}

// BenchmarkMemoryAllocatorReadAccess benchmarks memory read access
func BenchmarkMemoryAllocatorReadAccess(b *testing.B) {
	ma := createMemoryAllocator(1024 * 1024) // 1MB pool
	defer ma.Recycle()
	mem := ma.Malloc(1024)
	if mem == nil {
		b.Fatal("Failed to allocate memory")
	}
	// Pre-fill with data
	for j := 0; j < len(mem); j++ {
		mem[j] = byte(j % 256)
	}

	b.ResetTimer()
	var sum int
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(mem); j++ {
			sum += int(mem[j])
		}
	}
	_ = sum
}

// BenchmarkMemoryAllocatorCopyAccess benchmarks memory copy access
func BenchmarkMemoryAllocatorCopyAccess(b *testing.B) {
	ma := createMemoryAllocator(1024 * 1024) // 1MB pool
	defer ma.Recycle()
	mem := ma.Malloc(1024)
	if mem == nil {
		b.Fatal("Failed to allocate memory")
	}
	buf := make([]byte, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(buf, mem)
	}
}

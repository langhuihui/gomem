package gomem

import (
	"testing"
)

// BenchmarkBuddyAlloc benchmarks buddy allocator allocation
func BenchmarkBuddyAlloc(b *testing.B) {
	buddy := NewBuddy()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			offset, err := buddy.Alloc(1024)
			if err != nil {
				b.Fatal("Failed to allocate memory:", err)
			}
			err = buddy.Free(offset)
			if err != nil {
				b.Fatal("Failed to free memory:", err)
			}
		}
	})
}

// BenchmarkBuddyAllocSmall benchmarks small buddy allocation
func BenchmarkBuddyAllocSmall(b *testing.B) {
	buddy := NewBuddy()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			offset, err := buddy.Alloc(64)
			if err != nil {
				b.Fatal("Failed to allocate memory:", err)
			}
			err = buddy.Free(offset)
			if err != nil {
				b.Fatal("Failed to free memory:", err)
			}
		}
	})
}

// BenchmarkBuddyAllocLarge benchmarks large buddy allocation
func BenchmarkBuddyAllocLarge(b *testing.B) {
	buddy := NewBuddy()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			offset, err := buddy.Alloc(8192)
			if err != nil {
				b.Fatal("Failed to allocate memory:", err)
			}
			err = buddy.Free(offset)
			if err != nil {
				b.Fatal("Failed to free memory:", err)
			}
		}
	})
}

// BenchmarkBuddySequentialAlloc benchmarks sequential allocation and deallocation
func BenchmarkBuddySequentialAlloc(b *testing.B) {
	buddy := NewBuddy()
	allocations := make([]int, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Allocate 100 blocks
		for j := 0; j < 100; j++ {
			offset, err := buddy.Alloc(1024)
			if err != nil {
				b.Fatal("Failed to allocate memory:", err)
			}
			allocations[j] = offset
		}

		// Free all blocks
		for j := 0; j < 100; j++ {
			err := buddy.Free(allocations[j])
			if err != nil {
				b.Fatal("Failed to free memory:", err)
			}
		}
	}
}

// BenchmarkBuddyRandomAlloc benchmarks random allocation pattern
func BenchmarkBuddyRandomAlloc(b *testing.B) {
	buddy := NewBuddy()
	sizes := []int{64, 128, 256, 512, 1024, 2048, 4096}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			size := sizes[b.N%len(sizes)]
			offset, err := buddy.Alloc(size)
			if err != nil {
				b.Fatal("Failed to allocate memory:", err)
			}
			err = buddy.Free(offset)
			if err != nil {
				b.Fatal("Failed to free memory:", err)
			}
		}
	})
}

// BenchmarkBuddyPool benchmarks buddy pool operations
func BenchmarkBuddyPool(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buddy := GetBuddy()
			offset, err := buddy.Alloc(1024)
			if err != nil {
				b.Fatal("Failed to allocate memory:", err)
			}
			err = buddy.Free(offset)
			if err != nil {
				b.Fatal("Failed to free memory:", err)
			}
			PutBuddy(buddy)
		}
	})
}

// BenchmarkBuddyNonPowerOf2 benchmarks non-power-of-2 allocation
func BenchmarkBuddyNonPowerOf2(b *testing.B) {
	buddy := NewBuddy()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			offset, err := buddy.Alloc(1000) // Not a power of 2
			if err != nil {
				b.Fatal("Failed to allocate memory:", err)
			}
			err = buddy.Free(offset)
			if err != nil {
				b.Fatal("Failed to free memory:", err)
			}
		}
	})
}

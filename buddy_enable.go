//go:build enable_buddy

package gomem

func createMemoryAllocatorFromBuddy(size int, buddy *Buddy, offset int) *MemoryAllocator {
	ret := &MemoryAllocator{
		allocator: NewAllocator(size),
		Size:      size,
		memory:    buddy.memoryPool[offset : offset+size],
		start:     buddy.poolStart + int64(offset),
		recycle: func() {
			buddy.Free(offset >> MinPowerOf2)
		},
	}
	ret.allocator.Init(size)
	return ret
}

func GetMemoryAllocator(size int) (ret *MemoryAllocator) {
	if size < BuddySize {
		requiredSize := size >> MinPowerOf2
		// Loop to get an available buddy from the pool
		for {
			buddy := GetBuddy()
			defer PutBuddy(buddy)
			offset, err := buddy.Alloc(requiredSize)
			if err == nil {
				// Allocation successful, use this buddy
				return createMemoryAllocatorFromBuddy(size, buddy, offset<<MinPowerOf2)
			}
		}
	}
	// No buddy available or size too large, use system memory
	return createMemoryAllocator(size)
}

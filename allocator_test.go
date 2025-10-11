package gomem

import (
	"slices"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestAllocator(t *testing.T) {
	allocator := NewAllocator(1000)

	// Allocate memory
	block1 := allocator.Allocate(100)
	if block1 != 0 {
		t.Error("Failed to allocate memory")
	}

	// Allocate memory
	block2 := allocator.Allocate(200)
	if block2 != 100 {
		t.Error("Failed to allocate memory")
	}

	// Free memory
	allocator.Free(0, 299)
	if allocator.GetFreeSize() != 999 {
		t.Error("Failed to free memory")
	}
	allocator.Free(299, 1)

	// Reallocate memory
	block3 := allocator.Allocate(50)
	if block3 != 0 {
		t.Error("Failed to allocate memory")
	}

	// Free memory
	allocator.Free(0, 50)

	// Allocate memory larger than available space
	block4 := allocator.Allocate(1000)
	if block4 != 0 {
		t.Error("Should not allocate memory larger than available space")
	}
}

func FuzzAllocator(f *testing.F) {
	f.Add(100, false)
	allocator := NewAllocator(65535)
	var used [][2]int
	var totalMalloc, totalFree int = 0, 0
	f.Fuzz(func(t *testing.T, size int, alloc bool) {
		free := !alloc
		if size <= 0 {
			return
		}
		t.Logf("totalFree:%d,size:%d, free:%v", totalFree, size, free)
		defer func() {
			t.Logf("totalMalloc:%d, totalFree:%d, freeSize:%d", totalMalloc, totalFree, allocator.GetFreeSize())
			if totalMalloc-totalFree != allocator.Size-allocator.GetFreeSize() {
				t.Logf("totalUsed:%d, used:%d", totalMalloc-totalFree, allocator.Size-allocator.GetFreeSize())
				t.FailNow()
			}
		}()
		if free {
			if len(used) == 0 {
				return
			}
			for _, u := range used {
				if u[1] > size {
					totalFree += size
					t.Logf("totalFree1:%d, free:%v", totalFree, size)
					allocator.Free(u[0], size)
					u[1] -= size
					u[0] += size
					return
				}
			}
			allocator.Free(used[0][0], used[0][1])
			totalFree += used[0][1]
			t.Logf("totalFree2:%d, free:%v", totalFree, used[0][1])
			used = slices.Delete(used, 0, 1)
			return
		}
		offset := allocator.Allocate(size)
		if offset == -1 {
			return
		}
		used = append(used, [2]int{offset, size})
		totalMalloc += size
		t.Logf("totalMalloc:%d, free:%v", totalMalloc, size)
	})
}

const testData = `
- malloc: true
  offset: 0
  size: 16384
- malloc: false
  offset: 139
  size: 16245
- malloc: false
  offset: 0
  size: 50
- malloc: false
  offset: 50
  size: 31
- malloc: false
  offset: 81
  size: 9
- malloc: false
  offset: 90
  size: 26
- malloc: false
  offset: 116
  size: 21
- malloc: false
  offset: 137
  size: 2
- malloc: true
  offset: 0
  size: 16384
- malloc: false
  offset: 277
  size: 16107
- malloc: true
  offset: 0
  size: 16384
- malloc: false
  offset: 432
  size: 16229
- malloc: false
  offset: 0
  size: 277
- malloc: false
  offset: 277
  size: 58
- malloc: false
  offset: 335
  size: 60
- malloc: false
  offset: 395
  size: 9
- malloc: false
  offset: 404
  size: 26
- malloc: true
  offset: 0
  size: 16384
- malloc: false
  offset: 557
  size: 16259
- malloc: false
  offset: 430
  size: 2
`

var history []History

func init() {
	yaml.Unmarshal([]byte(testData), &history)
}

func TestAllocatorUseData(t *testing.T) {
	allocator := NewAllocator(65535)
	for _, h := range history {
		if h.Malloc {
			allocator.Allocate(h.Size)
		} else {
			allocator.Free(h.Offset, h.Size)
		}
	}
}

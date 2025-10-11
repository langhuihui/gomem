//go:build !enable_buddy && !disable_rm

package gomem

import (
	"sync"
)

var pool0, pool1, pool2 sync.Pool

func init() {
	pool0.New = func() any {
		ret := createMemoryAllocator(defaultBufSize)
		ret.recycle = func() {
			pool0.Put(ret)
		}
		return ret
	}
	pool1.New = func() any {
		ret := createMemoryAllocator(1 << MinPowerOf2)
		ret.recycle = func() {
			pool1.Put(ret)
		}
		return ret
	}
	pool2.New = func() any {
		ret := createMemoryAllocator(1 << (MinPowerOf2 + 2))
		ret.recycle = func() {
			pool2.Put(ret)
		}
		return ret
	}
}

func GetMemoryAllocator(size int) (ret *MemoryAllocator) {
	switch size {
	case defaultBufSize:
		ret = pool0.Get().(*MemoryAllocator)
		ret.allocator.Init(size)
	case 1 << MinPowerOf2:
		ret = pool1.Get().(*MemoryAllocator)
		ret.allocator.Init(size)
	case 1 << (MinPowerOf2 + 2):
		ret = pool2.Get().(*MemoryAllocator)
		ret.allocator.Init(size)
	default:
		ret = createMemoryAllocator(size)
	}
	return
}

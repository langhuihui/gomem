//go:build disable_rm

package gomem

func GetMemoryAllocator(size int) (ret *MemoryAllocator) {
	return &MemoryAllocator{Size: size}
}

package gomem

import (
	"errors"
	"sync"
	"unsafe"
)

type Buddy struct {
	size       int
	longests   [BuddySize>>(MinPowerOf2-1) - 1]int
	memoryPool [BuddySize]byte
	poolStart  int64
	lock       sync.Mutex // protects concurrent access to longests array
}

var (
	InValidParameterErr = errors.New("buddy: invalid parameter")
	NotFoundErr         = errors.New("buddy: can't find block")
	buddyPool           = sync.Pool{
		New: func() interface{} {
			return NewBuddy()
		},
	}
)

// GetBuddy gets a Buddy instance from the pool
func GetBuddy() *Buddy {
	buddy := buddyPool.Get().(*Buddy)
	return buddy
}

// PutBuddy puts a Buddy instance back to the pool
func PutBuddy(b *Buddy) {
	buddyPool.Put(b)
}

// NewBuddy creates a buddy instance.
// If the parameter isn't valid, return the nil and error as well
func NewBuddy() *Buddy {
	size := BuddySize >> MinPowerOf2
	ret := &Buddy{
		size: size,
	}
	for nodeSize, i := 2*size, 0; i < len(ret.longests); i++ {
		if isPowerOf2(i + 1) {
			nodeSize /= 2
		}
		ret.longests[i] = nodeSize
	}
	ret.poolStart = int64(uintptr(unsafe.Pointer(&ret.memoryPool[0])))

	return ret
}

// Alloc find a unused block according to the size
// return the offset of the block(regard 0 as the beginning)
// and parameter error if any
func (b *Buddy) Alloc(size int) (offset int, err error) {
	if size <= 0 {
		err = InValidParameterErr
		return
	}
	if !isPowerOf2(size) {
		size = fixSize(size)
	}
	b.lock.Lock()
	defer b.lock.Unlock()
	if size > b.longests[0] {
		err = NotFoundErr
		return
	}
	index := 0
	for nodeSize := b.size; nodeSize != size; nodeSize /= 2 {
		if left := leftChild(index); b.longests[left] >= size {
			index = left
		} else {
			index = rightChild(index)
		}
	}
	b.longests[index] = 0 // mark zero as used
	offset = (index+1)*size - b.size
	// update the parent node's size
	for index != 0 {
		index = parent(index)
		b.longests[index] = max(b.longests[leftChild(index)], b.longests[rightChild(index)])
	}
	return
}

// Free find a block according to the offset and mark it as unused
// return error if not found or parameter invalid
func (b *Buddy) Free(offset int) error {
	if offset < 0 || offset >= b.size {
		return InValidParameterErr
	}
	b.lock.Lock()
	defer b.lock.Unlock()
	nodeSize := 1
	index := offset + b.size - 1
	for ; b.longests[index] != 0; index = parent(index) {
		nodeSize *= 2
		if index == 0 {
			return NotFoundErr
		}
	}
	b.longests[index] = nodeSize
	// update parent node's size
	for index != 0 {
		index = parent(index)
		nodeSize *= 2

		leftSize := b.longests[leftChild(index)]
		rightSize := b.longests[rightChild(index)]
		if leftSize+rightSize == nodeSize {
			b.longests[index] = nodeSize
		} else {
			b.longests[index] = max(leftSize, rightSize)
		}
	}
	return nil
}

// helpers
func isPowerOf2(size int) bool {
	return size&(size-1) == 0
}

func fixSize(size int) int {
	size |= size >> 1
	size |= size >> 2
	size |= size >> 4
	size |= size >> 8
	size |= size >> 16
	return size + 1
}

func leftChild(index int) int {
	return 2*index + 1
}

func rightChild(index int) int {
	return 2*index + 2
}

func parent(index int) int {
	return (index+1)/2 - 1
}

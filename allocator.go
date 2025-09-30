//go:build !twotree

package gomem

type (
	Block struct {
		Start, End          int
		parent, left, right *Block
	}
	History struct {
		Malloc bool
		Offset int
		Size   int
	}
	Allocator struct {
		pool     *Block
		sizeTree *Block // Single treap instead of sizeTree/offsetTree
		Size     int
		// history  []History
	}
)

// Update NewAllocator
func NewAllocator(size int) (result *Allocator) {
	result = &Allocator{
		sizeTree: &Block{Start: 0, End: size},
		Size:     size,
	}
	return
}

func (p *Block) rotateDone(x, y *Block, a *Allocator) {
	x.parent = p
	if p == nil {
		a.sizeTree = x
	} else if p.left == y {
		p.left = x
	} else {
		p.right = x
	}
}

// Add rotation helpers similar to semaRoot
func (x *Block) rotateLeft(a *Allocator) {
	p, y, b := x.parent, x.right, x.right.left
	y.left, x.parent, x.right = x, y, b
	if b != nil {
		b.parent = x
	}
	p.rotateDone(y, x, a)
}

func (y *Block) rotateRight(a *Allocator) {
	p, x, b := y.parent, y.left, y.left.right
	x.right, y.parent, y.left = y, x, b
	if b != nil {
		b.parent = y
	}
	p.rotateDone(x, y, a)
}

func (b *Block) insert(block *Block, allocator *Allocator) *Block {
	if b == nil {
		return block
	}

	if block.End == block.Start {
		panic("empty block")
	}

	// Insert as BST using Start value
	if block.Start < b.Start {
		b.left = b.left.insert(block, allocator)
		if b.left != nil {
			b.left.parent = b
		}
	} else {
		b.right = b.right.insert(block, allocator)
		if b.right != nil {
			b.right.parent = b
		}
	}

	// Heapify based on block size (End-Start)
	blockSize := block.End - block.Start
	nodeSize := b.End - b.Start

	if blockSize < nodeSize {
		// Need to rotate up if current node has smaller size
		if block == b.left {
			b.rotateRight(allocator)
			return block
		} else if block == b.right {
			b.rotateLeft(allocator)
			return block
		}
	}
	return b
}

func (b *Block) find(size int) (block *Block) {
	if b == nil {
		return nil
	}

	// First check if current block can be used
	if blockSize := b.End - b.Start; blockSize >= size {
		// If exact match, return this block
		if blockSize == size {
			return b
		}
		// Keep searching left for potentially better fit
		if left := b.left.find(size); left != nil {
			return left
		}
		// If no better fit found, use this block
		return b
	}
	// If current block too small, only check right side
	return b.right.find(size)
}

func (b *Block) Walk(fn func(*Block)) {
	if b == nil {
		return
	}
	b.left.Walk(fn)
	fn(b)
	b.right.Walk(fn)
}

func (a *Allocator) putBlock(block *Block) {
	block.right = nil
	block.left = nil
	block.parent = a.pool
	a.pool = block
}

func (a *Allocator) Allocate(size int) (offset int) {
	// a.history = append(a.history, History{Malloc: true, Size: size})
	block := a.sizeTree.find(size)
	if block == nil {
		return -1
	}
	offset = block.Start
	a.deleteBlock(block)
	if blockSize := block.End - block.Start; blockSize == size {
		// Remove entire block
		a.putBlock(block)
	} else {
		block.Start += size
		a.insert(block)
	}
	return offset
}

func (a *Allocator) deleteBlock(block *Block) {
	// Rotate block down to leaf
	for block.left != nil || block.right != nil {
		if block.right == nil || (block.left != nil && (block.left.End-block.left.Start) > (block.right.End-block.right.Start)) {
			block.rotateRight(a)
		} else {
			block.rotateLeft(a)
		}
	}

	// Remove leaf
	if p := block.parent; p != nil {
		if p.left == block {
			p.left = nil
		} else {
			p.right = nil
		}
	} else {
		a.sizeTree = nil
	}
}

func (a *Allocator) insert(block *Block) {
	// a.sizeTree.Walk(func(b *Block) {
	// 	if block.Start >= b.Start && block.Start < b.End {
	// 		out, _ := yaml.Marshal(a.history)
	// 		fmt.Println(string(out))
	// 	}
	// })
	a.sizeTree = a.sizeTree.insert(block, a)
	// if a.sizeTree.parent != nil {
	// 	panic("sizeTree parent is not nil")
	// }
}

func (a *Allocator) Free(offset, size int) {
	// a.history = append(a.history, History{Malloc: false, Offset: offset, Size: size})
	// Try to merge with adjacent blocks
	// Find adjacent blocks
	switch left, right := a.findLeftAdjacent(offset), a.findRightAdjacent(offset+size); true {
	case left != nil && right != nil:
		a.deleteBlock(right)
		a.deleteBlock(left)
		left.End = right.End
		a.insert(left)
		a.putBlock(right)
	case left == nil && right == nil:
		block := a.getBlock(offset, offset+size)
		a.insert(block)
	case left != nil:
		a.deleteBlock(left)
		left.End = offset + size
		a.insert(left)
	case right != nil:
		a.deleteBlock(right)
		right.Start = offset
		a.insert(right)
	}
}

func (a *Allocator) findLeftAdjacent(offset int) (curr *Block) {
	curr = a.sizeTree
	for curr != nil {
		if curr.End == offset {
			return
		}
		if curr.End < offset {
			curr = curr.right
		} else {
			curr = curr.left
		}
	}
	return
}

func (a *Allocator) findRightAdjacent(offset int) (curr *Block) {
	curr = a.sizeTree
	for curr != nil {
		if curr.Start == offset {
			return
		}
		if curr.Start > offset {
			curr = curr.left
		} else {
			curr = curr.right
		}
	}
	return
}

func (a *Allocator) getBlock(start, end int) *Block {
	if a.pool == nil {
		return &Block{Start: start, End: end}
	} else {
		block := a.pool
		a.pool = block.parent
		block.parent = nil
		block.Start, block.End = start, end
		return block
	}
}

func (a *Allocator) GetFreeSize() (size int) {
	a.sizeTree.Walk(func(b *Block) {
		size += b.End - b.Start
	})
	return
}

func (a *Allocator) Recycle() {
	a.sizeTree.Walk(a.putBlock)
	a.sizeTree = nil
	a.pool = nil
}

func (a *Allocator) Init(size int) {
	a.sizeTree = a.getBlock(0, size)
	a.Size = size
}

func (a *Allocator) Find(size int) (offset int) {
	block := a.sizeTree.find(size)
	if block == nil {
		return -1
	}
	return block.Start
}

func (a *Allocator) GetBlocks() (blocks []*Block) {
	a.sizeTree.Walk(func(b *Block) {
		blocks = append(blocks, b)
	})
	return
}

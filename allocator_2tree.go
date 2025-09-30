//go:build twotree

package gomem

const TreeIndexSize = 0
const TreeIndexOffset = 1

type (
	Tree struct {
		left, right *Block
		height      int
	}
	Block struct {
		Start, End int
		trees      [2]Tree
	}
	History struct {
		Malloc bool
		Offset int
		Size   int
	}
	Allocator struct {
		pool       *Block
		sizeTree   *Block
		offsetTree *Block
		Size       int
		//history    []History
	}
)

func (t *Tree) deleteLeft(b *Block, treeIndex int) {
	t.left = t.left.delete(b, treeIndex)
}

func (t *Tree) deleteRight(b *Block, treeIndex int) {
	t.right = t.right.delete(b, treeIndex)
}

func NewAllocator(size int) (result *Allocator) {
	root := &Block{Start: 0, End: size}
	result = &Allocator{
		sizeTree:   root,
		offsetTree: root,
		Size:       size,
	}
	return
}

func compareBySize(a, b *Block) bool {
	//if a.Start == b.Start {
	//	panic("duplicate block")
	//}
	if sizea, sizeb := a.End-a.Start, b.End-b.Start; sizea != sizeb {
		return sizea < sizeb
	}
	return a.Start < b.Start
}

func compareByOffset(a, b *Block) bool {
	//if a.Start == b.Start {
	//	panic("duplicate block")
	//}
	return a.Start < b.Start
}

var compares = [...]func(a, b *Block) bool{compareBySize, compareByOffset}
var emptyTrees = [2]Tree{}

func (b *Block) insert(block *Block, treeIndex int) *Block {
	if b == nil {
		return block
	}
	if tree := &b.trees[treeIndex]; compares[treeIndex](block, b) {
		tree.left = tree.left.insert(block, treeIndex)
	} else {
		tree.right = tree.right.insert(block, treeIndex)
	}
	b.updateHeight(treeIndex)
	return b.balance(treeIndex)
}

func (b *Block) getLeftHeight(treeIndex int) int {
	return b.trees[treeIndex].left.getHeight(treeIndex)
}

func (b *Block) getRightHeight(treeIndex int) int {
	return b.trees[treeIndex].right.getHeight(treeIndex)
}

func (b *Block) getHeight(treeIndex int) int {
	if b == nil {
		return 0
	}
	return b.trees[treeIndex].height
}

func (b *Block) updateHeight(treeIndex int) {
	b.trees[treeIndex].height = 1 + max(b.getLeftHeight(treeIndex), b.getRightHeight(treeIndex))
}

func (b *Block) balance(treeIndex int) *Block {
	if b == nil {
		return nil
	}
	if tree := &b.trees[treeIndex]; b.getLeftHeight(treeIndex)-b.getRightHeight(treeIndex) > 1 {
		if tree.left.getRightHeight(treeIndex) > tree.left.getLeftHeight(treeIndex) {
			tree.left = tree.left.rotateLeft(treeIndex)
		}
		return b.rotateRight(treeIndex)
	} else if b.getRightHeight(treeIndex)-b.getLeftHeight(treeIndex) > 1 {
		if tree.right.getLeftHeight(treeIndex) > tree.right.getRightHeight(treeIndex) {
			tree.right = tree.right.rotateRight(treeIndex)
		}
		return b.rotateLeft(treeIndex)
	}
	return b
}

func (b *Block) rotateLeft(treeIndex int) *Block {
	newRoot := b.trees[treeIndex].right
	b.trees[treeIndex].right = newRoot.trees[treeIndex].left
	newRoot.trees[treeIndex].left = b
	b.updateHeight(treeIndex)
	newRoot.updateHeight(treeIndex)
	return newRoot
}

func (b *Block) rotateRight(treeIndex int) *Block {
	newRoot := b.trees[treeIndex].left
	b.trees[treeIndex].left = newRoot.trees[treeIndex].right
	newRoot.trees[treeIndex].right = b
	b.updateHeight(treeIndex)
	newRoot.updateHeight(treeIndex)
	return newRoot
}

func (b *Block) findMin(treeIndex int) *Block {
	if left := b.trees[treeIndex].left; left == nil {
		return b
	} else {
		return left.findMin(treeIndex)
	}
}

func (b *Block) delete(block *Block, treeIndex int) *Block {
	if b == nil {
		return nil
	}
	defer func() {
		block.trees[treeIndex] = emptyTrees[treeIndex]
	}()
	if compareFunc, tree := compares[treeIndex], &b.trees[treeIndex]; b == block {
		if tree.left == nil {
			return tree.right
		} else if tree.right == nil {
			return tree.left
		}
		minBlock := tree.right.findMin(treeIndex)
		tree.deleteRight(minBlock, treeIndex)
		minTree := &minBlock.trees[treeIndex]
		minTree.left = tree.left
		minTree.right = tree.right
		minTree.height = tree.height
		return minBlock
	} else if compareFunc(block, b) {
		tree.deleteLeft(block, treeIndex)
	} else {
		tree.deleteRight(block, treeIndex)
	}
	b.updateHeight(treeIndex)
	return b.balance(treeIndex)
}

func (a *Allocator) Init(size int) {
	a.Size = size
	root := a.getBlock(0, size)
	a.sizeTree = root
	a.offsetTree = root
}

func (a *Allocator) Find(size int) (offset int) {
	block := a.findAvailableBlock(size)
	if block == nil {
		return -1
	}
	offset = block.Start
	return
}

func (a *Allocator) Allocate(size int) (offset int) {
	//a.history = append(a.history, History{Malloc: true, Size: size})
	block := a.findAvailableBlock(size)
	if block == nil {
		return -1
	}
	offset = block.Start
	a.deleteSizeTree(block)
	a.deleteOffsetTree(block)
	if newStart := offset + size; newStart < block.End {
		block.Start = newStart
		a.insertSizeTree(block)
		a.insertOffsetTree(block)
	} else {
		a.putBlock(block)
	}
	return
}

func (a *Allocator) findAvailableBlock(size int) (lastAvailableBlock *Block) {
	block := a.sizeTree
	for block != nil {
		if bSize := block.End - block.Start; bSize == size {
			return block
		} else if tree := &block.trees[TreeIndexSize]; size < bSize {
			lastAvailableBlock = block
			block = tree.left
		} else {
			block = tree.right
		}
	}
	return
}

func (a *Allocator) getBlock(start, end int) *Block {
	if a.pool == nil {
		return &Block{Start: start, End: end}
	} else {
		block := a.pool
		a.pool = block.trees[TreeIndexSize].left
		block.trees = emptyTrees
		block.Start, block.End = start, end
		return block
	}
}

func (a *Allocator) putBlock(b *Block) {
	b.trees = emptyTrees
	b.trees[TreeIndexSize].left = a.pool
	a.pool = b
}

func (a *Allocator) Free(offset, size int) {
	//a.history = append(a.history, History{Malloc: false, Offset: offset, Size: size})
	switch leftAdjacent, rightAdjacent := a.offsetTree.findLeftAdjacentBlock(offset), a.offsetTree.findRightAdjacentBlock(offset+size); true {
	case leftAdjacent != nil && rightAdjacent != nil:
		a.deleteOffsetTree(rightAdjacent)
		a.deleteSizeTree(rightAdjacent)
		a.deleteSizeTree(leftAdjacent)
		leftAdjacent.End = rightAdjacent.End
		a.insertSizeTree(leftAdjacent)
		a.putBlock(rightAdjacent)
	case leftAdjacent == nil && rightAdjacent == nil:
		block := a.getBlock(offset, offset+size)
		a.insertSizeTree(block)
		a.insertOffsetTree(block)
	case leftAdjacent != nil:
		a.deleteSizeTree(leftAdjacent)
		leftAdjacent.End = offset + size
		a.insertSizeTree(leftAdjacent)
	case rightAdjacent != nil:
		a.deleteOffsetTree(rightAdjacent)
		a.deleteSizeTree(rightAdjacent)
		rightAdjacent.Start = offset
		a.insertSizeTree(rightAdjacent)
		a.insertOffsetTree(rightAdjacent)
	}
}

func (a *Allocator) GetBlocks() (blocks []*Block) {
	a.offsetTree.Walk(func(block *Block) {
		blocks = append(blocks, block)
	}, 1)
	return
}

func (a *Allocator) GetFreeSize() (ret int) {
	a.offsetTree.Walk(func(block *Block) {
		ret += block.End - block.Start
	}, 1)
	return
}

func (a *Allocator) insertSizeTree(block *Block) {
	//if block.End == block.Start {
	//	panic("empty block")
	//}
	//a.sizeTree.Walk(func(b *Block) {
	//	if block.Start >= b.Start && block.Start < b.End {
	//		out, _ := yaml.Marshal(a.history)
	//		fmt.Println(string(out))
	//	}
	//}, 0)
	a.sizeTree = a.sizeTree.insert(block, TreeIndexSize)
}

func (a *Allocator) insertOffsetTree(block *Block) {
	//if block.End == block.Start {
	//	panic("empty block")
	//}
	//a.offsetTree.Walk(func(b *Block) {
	//	if block.Start >= b.Start && block.Start < b.End {
	//		out, _ := yaml.Marshal(a.history)
	//		fmt.Println(string(out))
	//	}
	//}, 1)
	a.offsetTree = a.offsetTree.insert(block, TreeIndexOffset)
}

func (a *Allocator) deleteSizeTree(block *Block) {
	a.sizeTree = a.sizeTree.delete(block, TreeIndexSize)
}

func (a *Allocator) deleteOffsetTree(block *Block) {
	a.offsetTree = a.offsetTree.delete(block, TreeIndexOffset)
}

func (b *Block) findLeftAdjacentBlock(offset int) *Block {
	for b != nil {
		if b.End == offset {
			return b
		}
		if tree := &b.trees[TreeIndexOffset]; b.End > offset {
			b = tree.left
		} else {
			b = tree.right
		}
	}
	return nil
}

func (b *Block) findRightAdjacentBlock(offset int) *Block {
	for b != nil {
		if b.Start == offset {
			return b
		}
		if tree := &b.trees[TreeIndexOffset]; b.Start < offset {
			b = tree.right
		} else {
			b = tree.left
		}
	}
	return nil
}

func (a *Allocator) Recycle() {
	a.sizeTree.Walk(func(block *Block) {
		a.putBlock(block)
	}, 0)
	a.sizeTree = nil
	a.offsetTree = nil
}

func (b *Block) Walk(fn func(*Block), index int) {
	if b == nil {
		return
	}
	b.trees[index].left.Walk(fn, index)
	fn(b)
	b.trees[index].right.Walk(fn, index)
}

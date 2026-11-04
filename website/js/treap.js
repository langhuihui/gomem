/*
 * A faithful JavaScript port of gomem/allocator.go (the default single-treap allocator).
 * Every mutation emits a "step" containing a full snapshot of the tree so that the
 * visualizer can replay the algorithm frame by frame.
 *
 *   - BST order : Block.Start  (in-order traversal == address order)
 *   - Heap order: Block size (End - Start), maintained by rotations on insert
 */
(function (global) {
  'use strict';

  class Block {
    constructor(id, start, end) {
      this.id = id;
      this.start = start;
      this.end = end;
      this.parent = null;
      this.left = null;
      this.right = null;
    }
    get size() {
      return this.end - this.start;
    }
  }

  const fmt = (b) => '[' + b.start + ', ' + b.end + ')';

  class TreapAllocator {
    constructor(size) {
      this.total = size;
      this.seq = 0;
      this.pool = null;
      this.root = new Block(++this.seq, 0, size);
      this.floating = new Set();
      this.steps = null;
      this.allocsView = [];
      this.curMemHL = null;
    }

    /* ---------------- recording ---------------- */

    begin() {
      this.steps = [];
    }

    end() {
      const s = this.steps || [];
      this.steps = null;
      return s;
    }

    emit(code, mark, desc, hl, opt) {
      if (!this.steps) return;
      const step = {
        code: code,
        mark: mark,
        desc: desc,
        hl: hl || {},
        kind: 'info',
        memHL: this.curMemHL,
        snap: this.snapshot(),
        allocs: this.allocsView.map((a) => Object.assign({}, a)),
      };
      if (opt) Object.assign(step, opt);
      this.steps.push(step);
    }

    snapshot() {
      const nodes = {};
      const all = {};
      const walk = (b, depth) => {
        if (!b || depth > 64) return;
        const n = {
          id: b.id,
          start: b.start,
          end: b.end,
          left: b.left ? b.left.id : null,
          right: b.right ? b.right.id : null,
        };
        nodes[b.id] = n;
        all[b.id] = n;
        walk(b.left, depth + 1);
        walk(b.right, depth + 1);
      };
      walk(this.root, 0);
      const floating = [];
      this.floating.forEach((b) => {
        if (nodes[b.id]) return;
        const f = { id: b.id, start: b.start, end: b.end };
        floating.push(f);
        all[b.id] = f;
      });
      let pool = 0;
      for (let p = this.pool; p && pool < 4096; p = p.parent) pool++;
      return {
        root: this.root ? this.root.id : null,
        nodes: nodes,
        floating: floating,
        all: all,
        pool: pool,
        total: this.total,
      };
    }

    pathHL(path, cur, cls) {
      const hl = {};
      path.forEach((id) => (hl[id] = 'path'));
      if (cur != null) hl[cur] = cls;
      return hl;
    }

    /* ---------------- rotations (rotateDone / rotateLeft / rotateRight) ---------------- */

    rotateDone(p, x, y) {
      x.parent = p;
      if (p == null) this.root = x;
      else if (p.left === y) p.left = x;
      else p.right = x;
    }

    rotateLeft(x) {
      const p = x.parent;
      const y = x.right;
      const b = y.left;
      y.left = x;
      x.parent = y;
      x.right = b;
      if (b) b.parent = x;
      this.rotateDone(p, y, x);
    }

    rotateRight(y) {
      const p = y.parent;
      const x = y.left;
      const b = x.right;
      x.right = y;
      y.parent = x;
      y.left = b;
      if (b) b.parent = y;
      this.rotateDone(p, x, y);
    }

    /* ---------------- pool ---------------- */

    putBlock(block) {
      block.right = null;
      block.left = null;
      block.parent = this.pool;
      this.pool = block;
    }

    getBlock(start, end) {
      if (this.pool == null) return new Block(++this.seq, start, end);
      const block = this.pool;
      this.pool = block.parent;
      block.parent = null;
      block.start = start;
      block.end = end;
      return block;
    }

    /* ---------------- find ---------------- */

    find(b, size, path) {
      if (!b) return null;
      path.push(b.id);
      const bs = b.size;
      if (bs >= size) {
        if (bs === size) {
          this.emit('find', 'exact', '块 ' + fmt(b) + ' 大小 ' + bs + ' 恰好等于 ' + size + '，精确命中！',
            this.pathHL(path, b.id, 'target'), { kind: 'visit' });
          return b;
        }
        this.emit('find', 'ge', '块 ' + fmt(b) + ' 大小 ' + bs + ' ≥ ' + size + '，可用；继续向左子树寻找更贴合的块',
          this.pathHL(path, b.id, 'cand'), { kind: 'visit' });
        const left = this.find(b.left, size, path);
        if (left) return left;
        this.emit('find', 'self', '左侧没有更合适的块，选用 ' + fmt(b),
          this.pathHL(path, b.id, 'target'), { kind: 'visit' });
        return b;
      }
      this.emit('find', 'right', '块 ' + fmt(b) + ' 大小 ' + bs + ' < ' + size + '，不够用，转向右子树',
        this.pathHL(path, b.id, 'visit'), { kind: 'visit' });
      return this.find(b.right, size, path);
    }

    /* ---------------- deleteBlock ---------------- */

    deleteBlock(block) {
      const id = block.id;
      this.emit('deleteBlock', 'loop', '从树中删除 ' + fmt(block) + '：先通过旋转把它“下沉”为叶子节点',
        { [id]: 'target' });
      while (block.left || block.right) {
        if (!block.right || (block.left && block.left.size > block.right.size)) {
          const c = block.left;
          const why = !block.right ? '只有左孩子' : '左孩子(' + c.size + ') 比右孩子(' + block.right.size + ') 大';
          this.emit('deleteBlock', 'rr', why + ' → 右旋，左孩子 ' + fmt(c) + ' 上浮',
            { [id]: 'target', [c.id]: 'rotate' });
          this.rotateRight(block);
          this.emit('deleteBlock', 'rr', '右旋完成，' + fmt(block) + ' 下沉一层',
            { [id]: 'target', [c.id]: 'rotate' }, { kind: 'struct' });
        } else {
          const c = block.right;
          const why = !block.left ? '只有右孩子' : '右孩子(' + c.size + ') 不小于左孩子(' + block.left.size + ')';
          this.emit('deleteBlock', 'rl', why + ' → 左旋，右孩子 ' + fmt(c) + ' 上浮',
            { [id]: 'target', [c.id]: 'rotate' });
          this.rotateLeft(block);
          this.emit('deleteBlock', 'rl', '左旋完成，' + fmt(block) + ' 下沉一层',
            { [id]: 'target', [c.id]: 'rotate' }, { kind: 'struct' });
        }
      }
      const p = block.parent;
      if (p) {
        if (p.left === block) p.left = null;
        else p.right = null;
      } else {
        this.root = null;
      }
      this.floating.add(block);
      this.emit('deleteBlock', p ? 'leaf' : 'empty',
        p ? fmt(block) + ' 已是叶子，直接从父节点上摘除' : fmt(block) + ' 是唯一节点，摘除后 sizeTree = nil',
        { [id]: 'target' }, { kind: 'struct' });
    }

    /* ---------------- insert ---------------- */

    insert(block) {
      if (!this.root) {
        this.root = block;
        this.floating.delete(block);
        this.emit('insert', 'attach', '树为空，' + fmt(block) + ' 直接成为根节点',
          { [block.id]: 'new' }, { kind: 'struct' });
        return;
      }
      this.emit('insert', 'cmp', '插入 ' + fmt(block) + '：先按 Start 走二叉搜索树路径',
        { [block.id]: 'new' });
      this.root = this.insertAt(this.root, block);
    }

    insertAt(b, block) {
      if (!b) return block;
      if (block.end === block.start) throw new Error('empty block');
      const goLeft = block.start < b.start;
      this.emit('insert', goLeft ? 'goleft' : 'goright',
        'Start ' + block.start + (goLeft ? ' < ' : ' ≥ ') + b.start + '，进入 ' + fmt(b) + ' 的' + (goLeft ? '左' : '右') + '子树',
        { [b.id]: 'visit', [block.id]: 'new' }, { kind: 'visit' });

      if (goLeft) {
        const wasEmpty = !b.left;
        b.left = this.insertAt(b.left, block);
        if (b.left) b.left.parent = b;
        if (wasEmpty) {
          this.floating.delete(block);
          this.emit('insert', 'attach', '挂载为 ' + fmt(b) + ' 的左孩子',
            { [b.id]: 'visit', [block.id]: 'new' }, { kind: 'struct' });
        }
      } else {
        const wasEmpty = !b.right;
        b.right = this.insertAt(b.right, block);
        if (b.right) b.right.parent = b;
        if (wasEmpty) {
          this.floating.delete(block);
          this.emit('insert', 'attach', '挂载为 ' + fmt(b) + ' 的右孩子',
            { [b.id]: 'visit', [block.id]: 'new' }, { kind: 'struct' });
        }
      }

      const isChild = block === b.left || block === b.right;
      if (block.size < b.size) {
        if (block === b.left) {
          this.emit('insert', 'heap', '堆性质：' + block.size + ' < 父节点 ' + b.size + '，需要右旋上浮',
            { [b.id]: 'rotate', [block.id]: 'new' });
          this.rotateRight(b);
          this.emit('insert', 'rr', '右旋完成，' + fmt(block) + ' 上浮一层',
            { [b.id]: 'rotate', [block.id]: 'new' }, { kind: 'struct' });
          return block;
        } else if (block === b.right) {
          this.emit('insert', 'heap', '堆性质：' + block.size + ' < 父节点 ' + b.size + '，需要左旋上浮',
            { [b.id]: 'rotate', [block.id]: 'new' });
          this.rotateLeft(b);
          this.emit('insert', 'rl', '左旋完成，' + fmt(block) + ' 上浮一层',
            { [b.id]: 'rotate', [block.id]: 'new' }, { kind: 'struct' });
          return block;
        }
      }
      if (isChild) {
        this.emit('insert', 'stop', '堆性质：' + block.size + ' ≥ 父节点 ' + b.size + '，无需旋转，插入结束',
          { [b.id]: 'visit', [block.id]: 'new' });
      }
      return b;
    }

    /* ---------------- Allocate ---------------- */

    allocate(size) {
      this.curMemHL = null;
      this.emit('Allocate', 'find', 'Allocate(' + size + ')：在 Treap 中查找能容纳 ' + size + ' 字节的空闲块');
      const block = this.root ? this.find(this.root, size, []) : null;
      if (!block) {
        this.emit('Allocate', 'fail', '没有找到 ≥ ' + size + ' 的空闲块，返回 -1（ScalableMemoryAllocator 会在此时新建子分配器扩容）',
          {}, { fail: true });
        return -1;
      }
      const offset = block.start;
      this.emit('Allocate', 'delete', '选中 ' + fmt(block) + '，offset = ' + offset + '。先把它从树中摘下',
        { [block.id]: 'target' });
      this.deleteBlock(block);
      this.curMemHL = { offset: offset, size: size, cls: 'alloc' };
      if (block.size === size) {
        this.floating.delete(block);
        this.putBlock(block);
        this.emit('Allocate', 'put', '整块被用完：Block 节点放回对象池等待复用（零 GC 分配）',
          {}, { kind: 'struct' });
      } else {
        block.start += size;
        this.emit('Allocate', 'shrink', '从块头部切下 ' + size + ' 字节，剩余 ' + fmt(block) + '（' + block.size + '）需要重新插入',
          { [block.id]: 'new' }, { kind: 'struct' });
        this.insert(block);
      }
      return offset;
    }

    /* ---------------- Free ---------------- */

    findLeftAdjacent(offset) {
      let curr = this.root;
      const path = [];
      while (curr) {
        path.push(curr.id);
        if (curr.end === offset) {
          this.emit('findLeftAdjacent', 'hit', fmt(curr) + '.End == ' + offset + '，找到左邻空闲块',
            this.pathHL(path, curr.id, 'merge'), { kind: 'visit' });
          return curr;
        }
        if (curr.end < offset) {
          this.emit('findLeftAdjacent', 'right', '找左邻：' + fmt(curr) + '.End ' + curr.end + ' < ' + offset + '，向右',
            this.pathHL(path, curr.id, 'visit'), { kind: 'visit' });
          curr = curr.right;
        } else {
          this.emit('findLeftAdjacent', 'left', '找左邻：' + fmt(curr) + '.End ' + curr.end + ' > ' + offset + '，向左',
            this.pathHL(path, curr.id, 'visit'), { kind: 'visit' });
          curr = curr.left;
        }
      }
      this.emit('findLeftAdjacent', 'miss', '没有 End == ' + offset + ' 的空闲块：左侧不相邻',
        this.pathHL(path), { kind: 'visit' });
      return null;
    }

    findRightAdjacent(offset) {
      let curr = this.root;
      const path = [];
      while (curr) {
        path.push(curr.id);
        if (curr.start === offset) {
          this.emit('findRightAdjacent', 'hit', fmt(curr) + '.Start == ' + offset + '，找到右邻空闲块',
            this.pathHL(path, curr.id, 'merge'), { kind: 'visit' });
          return curr;
        }
        if (curr.start > offset) {
          this.emit('findRightAdjacent', 'left', '找右邻：' + fmt(curr) + '.Start ' + curr.start + ' > ' + offset + '，向左',
            this.pathHL(path, curr.id, 'visit'), { kind: 'visit' });
          curr = curr.left;
        } else {
          this.emit('findRightAdjacent', 'right', '找右邻：' + fmt(curr) + '.Start ' + curr.start + ' < ' + offset + '，向右',
            this.pathHL(path, curr.id, 'visit'), { kind: 'visit' });
          curr = curr.right;
        }
      }
      this.emit('findRightAdjacent', 'miss', '没有 Start == ' + offset + ' 的空闲块：右侧不相邻',
        this.pathHL(path), { kind: 'visit' });
      return null;
    }

    free(offset, size) {
      this.curMemHL = { offset: offset, size: size, cls: 'free' };
      this.emit('Free', 'find', 'Free(' + offset + ', ' + size + ')：归还 [' + offset + ', ' + (offset + size) + ')，先找左右相邻的空闲块');
      const left = this.findLeftAdjacent(offset);
      const right = this.findRightAdjacent(offset + size);
      const hl = {};
      if (left) hl[left.id] = 'merge';
      if (right) hl[right.id] = 'merge';

      if (left && right) {
        this.emit('Free', 'both', '左邻 ' + fmt(left) + ' 与右邻 ' + fmt(right) + ' 都存在 → 三块合并为一块', hl);
        this.deleteBlock(right);
        this.deleteBlock(left);
        left.end = right.end;
        this.emit('Free', 'bothmerge', '合并：左块 End 延伸到 ' + right.end + '，得到 ' + fmt(left),
          { [left.id]: 'new', [right.id]: 'dead' }, { kind: 'struct' });
        this.insert(left);
        this.floating.delete(right);
        this.putBlock(right);
        this.emit('Free', 'put', '右块节点已被吞并，放回对象池复用', {}, { kind: 'struct' });
        return 'both';
      }
      if (!left && !right) {
        this.emit('Free', 'none', '两侧都没有相邻空闲块，需要一个新 Block 节点' + (this.pool ? '（从对象池取出）' : ''));
        const block = this.getBlock(offset, offset + size);
        this.floating.add(block);
        this.emit('Free', 'new', '得到节点 ' + fmt(block) + '，准备插入', { [block.id]: 'new' }, { kind: 'struct' });
        this.insert(block);
        return 'none';
      }
      if (left) {
        this.emit('Free', 'left', '只有左邻 ' + fmt(left) + '：与之合并', hl);
        this.deleteBlock(left);
        left.end = offset + size;
        this.emit('Free', 'leftmerge', '左块 End 延伸到 ' + (offset + size) + '，得到 ' + fmt(left),
          { [left.id]: 'new' }, { kind: 'struct' });
        this.insert(left);
        return 'left';
      }
      this.emit('Free', 'right', '只有右邻 ' + fmt(right) + '：与之合并', hl);
      this.deleteBlock(right);
      right.start = offset;
      this.emit('Free', 'rightmerge', '右块 Start 前移到 ' + offset + '，得到 ' + fmt(right),
        { [right.id]: 'new' }, { kind: 'struct' });
      this.insert(right);
      return 'right';
    }
  }

  /* Go source snippets shown in the code panel. `//@mark` tags map step marks to lines. */
  const SNIPPETS = {
    Allocate: [
      'func (a *Allocator) Allocate(size int) (offset int) {',
      '\tblock := a.sizeTree.find(size) //@find',
      '\tif block == nil {',
      '\t\treturn -1 //@fail',
      '\t}',
      '\toffset = block.Start',
      '\ta.deleteBlock(block) //@delete',
      '\tif blockSize := block.End - block.Start; blockSize == size {',
      '\t\ta.putBlock(block) //@put',
      '\t} else {',
      '\t\tblock.Start += size //@shrink',
      '\t\ta.insert(block)',
      '\t}',
      '\treturn offset //@done',
      '}',
    ],
    find: [
      'func (b *Block) find(size int) (block *Block) {',
      '\tif b == nil {',
      '\t\treturn nil',
      '\t}',
      '\tif blockSize := b.End - b.Start; blockSize >= size { //@ge',
      '\t\tif blockSize == size {',
      '\t\t\treturn b //@exact',
      '\t\t}',
      '\t\tif left := b.left.find(size); left != nil {',
      '\t\t\treturn left',
      '\t\t}',
      '\t\treturn b //@self',
      '\t}',
      '\treturn b.right.find(size) //@right',
      '}',
    ],
    deleteBlock: [
      'func (a *Allocator) deleteBlock(block *Block) {',
      '\t// Rotate block down to leaf',
      '\tfor block.left != nil || block.right != nil { //@loop',
      '\t\tif block.right == nil || (block.left != nil &&',
      '\t\t\t(block.left.End-block.left.Start) > (block.right.End-block.right.Start)) {',
      '\t\t\tblock.rotateRight(a) //@rr',
      '\t\t} else {',
      '\t\t\tblock.rotateLeft(a) //@rl',
      '\t\t}',
      '\t}',
      '\tif p := block.parent; p != nil { //@leaf',
      '\t\tif p.left == block {',
      '\t\t\tp.left = nil',
      '\t\t} else {',
      '\t\t\tp.right = nil',
      '\t\t}',
      '\t} else {',
      '\t\ta.sizeTree = nil //@empty',
      '\t}',
      '}',
    ],
    insert: [
      'func (b *Block) insert(block *Block, allocator *Allocator) *Block {',
      '\tif b == nil {',
      '\t\treturn block //@attach',
      '\t}',
      '\tif block.Start < b.Start { //@cmp',
      '\t\tb.left = b.left.insert(block, allocator) //@goleft',
      '\t\tif b.left != nil { b.left.parent = b }',
      '\t} else {',
      '\t\tb.right = b.right.insert(block, allocator) //@goright',
      '\t\tif b.right != nil { b.right.parent = b }',
      '\t}',
      '\tblockSize := block.End - block.Start',
      '\tnodeSize := b.End - b.Start',
      '\tif blockSize < nodeSize { //@heap',
      '\t\tif block == b.left {',
      '\t\t\tb.rotateRight(allocator) //@rr',
      '\t\t\treturn block',
      '\t\t} else if block == b.right {',
      '\t\t\tb.rotateLeft(allocator) //@rl',
      '\t\t\treturn block',
      '\t\t}',
      '\t}',
      '\treturn b //@stop',
      '}',
    ],
    Free: [
      'func (a *Allocator) Free(offset, size int) {',
      '\tswitch left, right := a.findLeftAdjacent(offset), //@find',
      '\t\ta.findRightAdjacent(offset+size); true {',
      '\tcase left != nil && right != nil: //@both',
      '\t\ta.deleteBlock(right)',
      '\t\ta.deleteBlock(left)',
      '\t\tleft.End = right.End //@bothmerge',
      '\t\ta.insert(left)',
      '\t\ta.putBlock(right) //@put',
      '\tcase left == nil && right == nil: //@none',
      '\t\tblock := a.getBlock(offset, offset+size) //@new',
      '\t\ta.insert(block)',
      '\tcase left != nil: //@left',
      '\t\ta.deleteBlock(left)',
      '\t\tleft.End = offset + size //@leftmerge',
      '\t\ta.insert(left)',
      '\tcase right != nil: //@right',
      '\t\ta.deleteBlock(right)',
      '\t\tright.Start = offset //@rightmerge',
      '\t\ta.insert(right)',
      '\t}',
      '} //@end',
    ],
    findLeftAdjacent: [
      'func (a *Allocator) findLeftAdjacent(offset int) (curr *Block) {',
      '\tcurr = a.sizeTree',
      '\tfor curr != nil {',
      '\t\tif curr.End == offset {',
      '\t\t\treturn //@hit',
      '\t\t}',
      '\t\tif curr.End < offset {',
      '\t\t\tcurr = curr.right //@right',
      '\t\t} else {',
      '\t\t\tcurr = curr.left //@left',
      '\t\t}',
      '\t}',
      '\treturn //@miss',
      '}',
    ],
    findRightAdjacent: [
      'func (a *Allocator) findRightAdjacent(offset int) (curr *Block) {',
      '\tcurr = a.sizeTree',
      '\tfor curr != nil {',
      '\t\tif curr.Start == offset {',
      '\t\t\treturn //@hit',
      '\t\t}',
      '\t\tif curr.Start > offset {',
      '\t\t\tcurr = curr.left //@left',
      '\t\t} else {',
      '\t\t\tcurr = curr.right //@right',
      '\t\t}',
      '\t}',
      '\treturn //@miss',
      '}',
    ],
  };

  global.GoMemTreap = { TreapAllocator: TreapAllocator, SNIPPETS: SNIPPETS };
})(window);

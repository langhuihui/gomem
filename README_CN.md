# GoMem

GoMem 是一个为 Go 语言设计的高性能内存分配器库，从 Monibuca 项目中提取而来。

## 特性

- **多种分配策略**: 支持单树和双树（AVL）分配算法
- **伙伴分配器**: 可选的伙伴系统，用于高效的内存池管理
- **可回收内存**: 支持内存回收，具有自动清理功能
- **可扩展分配器**: 动态增长的内存分配器
- **内存读取器**: 高效的多缓冲区读取器，支持零拷贝操作

## 构建标签

该库支持多个构建标签来自定义行为：

- `twotree`: 使用双树（AVL）实现替代单树 treap
- `enable_buddy`: 启用伙伴分配器进行内存池管理
- `disable_rm`: 禁用可回收内存功能以减少开销

## 安装

```bash
go get github.com/langhuihui/gomem
```

## 使用方法

### 基本内存分配

```go
package main

import "github.com/langhuihui/gomem"

func main() {
    // 创建一个可扩展的内存分配器
    allocator := gomem.NewScalableMemoryAllocator(1024)
    
    // 分配内存
    buf := allocator.Malloc(256)
    
    // 使用缓冲区...
    copy(buf, []byte("Hello, World!"))
    
    // 释放内存
    allocator.Free(buf)
}
```

### 可回收内存

```go
// 为批量操作创建可回收内存
allocator := gomem.NewScalableMemoryAllocator(1024)
rm := gomem.NewRecyclableMemory(allocator)

// 分配多个缓冲区
buf1 := rm.NextN(128)
buf2 := rm.NextN(256)

// 使用缓冲区...
copy(buf1, []byte("Buffer 1"))
copy(buf2, []byte("Buffer 2"))

// 一次性回收所有内存
rm.Recycle()
```

### 内存缓冲区操作

```go
// 创建一个内存缓冲区
mem := gomem.NewMemory([]byte{1, 2, 3, 4, 5})

// 添加更多数据
mem.PushOne([]byte{6, 7, 8})

// 获取总大小和缓冲区数量
fmt.Printf("Size: %d, Buffers: %d\n", mem.Size, mem.Count())

// 转换为字节数组
data := mem.ToBytes()
```

### 内存读取器

```go
// 创建一个内存读取器
reader := gomem.NewReadableBuffersFromBytes([]byte{1, 2, 3}, []byte{4, 5, 6})

// 读取数据
buf := make([]byte, 6)
n, err := reader.Read(buf)
// buf 现在包含 [1, 2, 3, 4, 5, 6]
```

## 性能考虑

- 在高吞吐量场景中使用 `enable_buddy` 构建标签以获得更好的内存池性能
- 当不需要可回收内存时使用 `disable_rm` 构建标签以减少开销
- 使用 `twotree` 构建标签以获得更平衡的分配性能

## 许可证

MIT

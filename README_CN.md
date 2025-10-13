# GoMem

<div align="center">
  <img src="logo.png" alt="GoMem Logo" width="200"/>
</div>

[![Go 版本](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![许可证](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/langhuihui/gomem?style=flat-square)](https://goreportcard.com/report/github.com/langhuihui/gomem)

> **语言**: [English](README.md) | [中文](README_CN.md)

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
- `enable_mmap`: 启用内存映射分配以提高内存效率（支持 Linux/macOS/Windows）
  - **Linux**: 自动启用透明大页（THP）支持，使用 2MB 大页替代 4KB 页面，显著减少 TLB 缺失并提升内存访问性能

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

### 分段内存释放

```go
package main

import "github.com/langhuihui/gomem"

func main() {
    // 创建一个可扩展的内存分配器
    allocator := gomem.NewScalableMemoryAllocator(1024)
    
    // 分配一大块内存
    buf := allocator.Malloc(1024)
    
    // 使用内存的不同部分
    part1 := buf[0:256]    // 前256字节
    part2 := buf[256:512]  // 中间256字节  
    part3 := buf[512:1024] // 后512字节
    
    // 填充数据
    copy(part1, []byte("Part 1 data"))
    copy(part2, []byte("Part 2 data"))
    copy(part3, []byte("Part 3 data"))
    
    // 分段释放内存 - 可以释放部分内存
    allocator.Free(part1)  // 释放前256字节
    allocator.Free(part2)  // 释放中间256字节
    
    // 继续使用剩余内存
    copy(part3, []byte("Updated part 3"))
    
    // 最后释放剩余内存
    allocator.Free(part3)
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

## 并发安全

⚠️ **重要**: Malloc 和 Free 操作必须在同一个协程中调用，以避免竞态问题。为了更优雅的使用，建议使用 [gotask](https://github.com/langhuihui/gotask)，可以在 `Start` 方法中申请内存，在 `Dispose` 方法中释放内存。

```go
// ❌ 错误：不同的协程
go func() {
    buf := allocator.Malloc(256)
    // ... 使用缓冲区
}()

go func() {
    allocator.Free(buf) // 竞态条件！
}()

// ✅ 正确：同一个协程
buf := allocator.Malloc(256)
// ... 使用缓冲区
allocator.Free(buf)

// ✅ 优雅：使用 gotask
type MyTask struct {
    allocator *gomem.ScalableMemoryAllocator
    buffer []byte
}

func (t *MyTask) Start() {
    t.allocator = gomem.NewScalableMemoryAllocator(1024)
    t.buffer = t.allocator.Malloc(256)
}

func (t *MyTask) Dispose() {
    t.allocator.Free(t.buffer)
}
```

## 性能考虑

- **使用 `enable_mmap` 构建标签可获得显著的性能提升**：分配器创建快100-400倍，内存使用减少99.98%
- 在高吞吐量场景中使用 `enable_buddy` 构建标签以获得更好的内存池性能
- **启用 RecyclableMemory 比禁用版本快53%**，且内存使用更少
- 仅在不需要内存管理功能时使用 `disable_rm` 构建标签（减少复杂度但牺牲性能）
- **单树分配器比双树分配器显著更快**（分配操作快77-86%）
- 仅在需要更快查找操作时使用 `twotree` 构建标签（比单树快100%）

## 基准测试结果

以下基准测试结果在 Apple M2 Pro (ARM64) 和 Go 1.23.0 环境下获得：

### MMAP vs 默认实现性能对比

MMAP 实现在内存效率方面提供了**显著的改进**，且性能开销极小：

| 指标 | 默认实现 | MMAP | 提升幅度 |
|------|---------|------|----------|
| **整体性能（几何平均）** | 234.1 ns/op | 94.21 ns/op | **快59.8%** ⚡ |
| **内存使用（几何平均）** | - | - | **减少86.6%** 💾 |
| **1MB 分配器创建** | 80.5 µs<br/>1,048,763 B | 799 ns<br/>216 B | **快100倍**<br/>**内存减少99.98%** 🚀 |
| **16MB 分配器创建** | 317.2 µs<br/>16,777,405 B | 777 ns<br/>216 B | **快408倍**<br/>**内存减少99.999%** 🚀 |
| **单次分配（1KB）** | 13.25 ns/op | 13.89 ns/op | 慢4.8% |
| **内存访问（写入）** | 441 ns/op | 458 ns/op | 慢3.8% |
| **内存访问（读取）** | 320 ns/op | 333 ns/op | 慢4.2% |

**关键发现：**
- **分配器创建**：MMAP 快100-408倍，内存使用减少99.98-99.999%
- **内存效率**：仅使用216字节元数据，而非立即分配整个缓冲区
- **分配操作**：仅慢3-6%（< 1纳秒开销）- 在大多数场景下可以忽略
- **虚拟内存**：MMAP 预留地址空间但不立即分配物理内存（惰性分配）

**何时使用 MMAP：**
- ✅ 创建多个或大型分配器
- ✅ 内存效率至关重要
- ✅ 频繁创建/销毁分配器
- ✅ 处理稀疏数据（不是所有内存都会立即使用）
- ✅ 需要预留大地址空间

**何时使用默认实现：**
- ⚠️ 分配操作中每纳秒都很重要（高频交易等）
- ⚠️ 所有分配的内存会立即使用
- ⚠️ 在不支持高效 mmap 的系统上运行

**启用 MMAP：**
```bash
go build -tags=enable_mmap
```

**Linux THP 支持：**
在 Linux 上使用 `enable_mmap` 时，会自动启用透明大页（Transparent Huge Pages，THP）：
- 使用 2MB 大页替代 4KB 小页（x86_64 架构）
- 显著减少 TLB（Translation Lookaside Buffer）缺失
- 提升大块内存访问性能
- 通过 `madvise(MADV_HUGEPAGE)` 系统调用实现
- 如果系统不支持 THP，会静默降级到常规页面，不影响程序运行

### 单树 vs 双树分配器性能比较

| 操作类型 | 单树 (ns/op) | 双树 (ns/op) | 性能差异 | 胜出者 |
|---------|-------------|-------------|---------|--------|
| **基础分配** | 12.33 | 22.71 | **快84%** | 单树 |
| **小内存分配 (64B)** | 12.32 | 22.60 | **快84%** | 单树 |
| **大内存分配 (8KB)** | 12.14 | 22.61 | **快86%** | 单树 |
| **顺序分配** | 1961 | 3467 | **快77%** | 单树 |
| **随机分配** | 12.47 | 23.02 | **快85%** | 单树 |
| **查找操作** | 3.03 | 1.51 | **快100%** | 双树 |
| **获取空闲大小** | 3.94 | 4.27 | **快8%** | 单树 |

**关键发现：**
- 单树分配器在内存分配操作上**快77-86%**
- 双树分配器仅在查找操作上**快100%**
- 由于分配性能更优，推荐在大多数用例中使用单树分配器

### RecyclableMemory 性能比较（启用 vs 禁用）

| 操作类型 | 启用 RM (ns/op) | 禁用 RM (ns/op) | 性能差异 | 内存使用 |
|---------|----------------|----------------|---------|---------|
| **基础操作** | 335.2 | 511.9 | **快53%** | 启用: 1536B/2 allocs, 禁用: 1788B/2 allocs |
| **多个分配** | - | 1035.1 | - | 禁用: 3875B/10 allocs |
| **Clone操作** | - | 53.7 | - | 禁用: 240B/1 alloc |

**关键发现：**
- 启用 RecyclableMemory 在基础操作上**快53%**
- 启用 RM 内存使用更少（1536B vs 1788B 基础操作）
- 启用 RM 提供真正的内存管理和回收功能
- 禁用 RM 使用简单的 `make([]byte, size)` 无内存池

### 内存分配器性能（单树）

| 基准测试 | 操作次数/秒 | 每次操作时间 | 内存/操作 | 分配次数/操作 |
|----------|-------------|-------------|-----------|---------------|
| Allocate | 96,758,520 | 15.08 ns | 0 B | 0 |
| AllocateSmall | 98,864,434 | 12.49 ns | 0 B | 0 |
| AllocateLarge | 100,000,000 | 12.65 ns | 0 B | 0 |
| SequentialAlloc | 1,321,965 | 942.2 ns | 0 B | 0 |
| RandomAlloc | 96,241,566 | 12.79 ns | 0 B | 0 |
| GetFreeSize | 303,367,089 | 3.934 ns | 0 B | 0 |

### 内存操作性能

| 基准测试 | 操作次数/秒 | 每次操作时间 | 内存/操作 | 分配次数/操作 |
|----------|-------------|-------------|-----------|---------------|
| PushOne | 31,982,593 | 35.05 ns | 143 B | 0 |
| Push | 17,666,751 | 70.40 ns | 259 B | 0 |
| ToBytes | 119,496 | 11,806 ns | 106,496 B | 1 |
| CopyTo | 417,379 | 2,905 ns | 0 B | 0 |
| Append | 979,598 | 1,859 ns | 7,319 B | 0 |
| Count | 1,000,000,000 | 0.3209 ns | 0 B | 0 |
| Range | 32,809,593 | 36.08 ns | 0 B | 0 |

### 内存读取器性能

| 基准测试 | 操作次数/秒 | 每次操作时间 | 内存/操作 | 分配次数/操作 |
|----------|-------------|-------------|-----------|---------------|
| Read | 10,355,643 | 112.4 ns | 112 B | 2 |
| ReadByte | 536,228 | 2,235 ns | 56 B | 2 |
| ReadBytes | 2,556,602 | 608.7 ns | 1,080 B | 18 |
| ReadBE | 408,663 | 3,587 ns | 56 B | 2 |
| Skip | 8,762,934 | 125.8 ns | 56 B | 2 |
| Range | 15,608,808 | 70.99 ns | 80 B | 2 |
| RangeN | 20,101,638 | 79.09 ns | 80 B | 2 |
| LEB128Unmarshal | 356,560 | 3,052 ns | 56 B | 2 |

### 伙伴分配器性能

| 基准测试 | 操作次数/秒 | 每次操作时间 | 内存/操作 | 分配次数/操作 |
|----------|-------------|-------------|-----------|---------------|
| Alloc | 4,017,826 | 388.2 ns | 0 B | 0 |
| AllocSmall | 3,092,535 | 410.7 ns | 0 B | 0 |
| AllocLarge | 3,723,950 | 276.4 ns | 0 B | 0 |
| SequentialAlloc | 62,786 | 17,997 ns | 0 B | 0 |
| RandomAlloc | 3,249,220 | 357.8 ns | 0 B | 0 |
| Pool | 27,800 | 56,846 ns | 196,139 B | 0 |
| NonPowerOf2 | 3,167,425 | 317.8 ns | 0 B | 0 |

### 性能总结

- **单树分配器**: 极快的分配/释放操作，每次操作约12ns，零内存分配
- **双树分配器**: 分配较慢（约23ns每次操作），但查找操作更快（约1.5ns vs 3ns）
- **启用 RecyclableMemory**: 比禁用版本快53%，内存效率更高
- **禁用 RecyclableMemory**: 实现更简单但性能较慢，内存使用更高
- **内存操作**: 高效的缓冲区管理，开销最小
- **内存读取器**: 高性能读取，支持零拷贝操作
- **伙伴分配器**: 快速的2的幂次分配，支持池化以减少GC压力

**推荐**: 
- **在大多数应用中使用 `enable_mmap` 标签**以获得60%的性能提升和87%的内存减少
- 由于分配性能更优，推荐在大多数应用中使用单树分配器（默认）
- 保持 RecyclableMemory 启用（默认）以获得更好的性能和内存效率
- 仅在查找操作关键且频繁时才使用双树分配器
- 仅在不需要内存管理功能时才使用 `disable_rm` 标签

## 许可证

MIT

---

## 贡献

欢迎贡献代码！请随时提交 Pull Request。

## 支持

如果您有任何问题或需要帮助，请在 GitHub 上提交 issue。

## Star 历史

[![Star History Chart](https://api.star-history.com/svg?repos=langhuihui/gomem&type=Date)](https://star-history.com/#langhuihui/gomem&Date)

---

<div align="center">
  <sub>由 GoMem 团队用 ❤️ 构建</sub>
</div>

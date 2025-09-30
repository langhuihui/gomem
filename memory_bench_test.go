package gomem

import (
	"testing"
)

// BenchmarkMemoryPushOne benchmarks pushing single buffers
func BenchmarkMemoryPushOne(b *testing.B) {
	mem := NewMemory([]byte{})
	data := make([]byte, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mem.PushOne(data)
	}
}

// BenchmarkMemoryPush benchmarks pushing multiple buffers
func BenchmarkMemoryPush(b *testing.B) {
	mem := NewMemory([]byte{})
	data1 := make([]byte, 512)
	data2 := make([]byte, 512)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mem.Push(data1, data2)
	}
}

// BenchmarkMemoryToBytes benchmarks converting memory to bytes
func BenchmarkMemoryToBytes(b *testing.B) {
	mem := NewMemory([]byte{})
	// Pre-populate with some data
	for i := 0; i < 100; i++ {
		data := make([]byte, 1024)
		mem.PushOne(data)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mem.ToBytes()
	}
}

// BenchmarkMemoryCopyTo benchmarks copying memory to buffer
func BenchmarkMemoryCopyTo(b *testing.B) {
	mem := NewMemory([]byte{})
	// Pre-populate with some data
	for i := 0; i < 100; i++ {
		data := make([]byte, 1024)
		mem.PushOne(data)
	}

	buf := make([]byte, mem.Size)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mem.CopyTo(buf)
	}
}

// BenchmarkMemoryAppend benchmarks appending memory
func BenchmarkMemoryAppend(b *testing.B) {
	mem1 := NewMemory([]byte{})
	mem2 := NewMemory([]byte{})

	// Pre-populate with some data
	for i := 0; i < 50; i++ {
		data := make([]byte, 1024)
		mem1.PushOne(data)
		mem2.PushOne(data)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mem1.Append(mem2)
	}
}

// BenchmarkMemoryCount benchmarks getting buffer count
func BenchmarkMemoryCount(b *testing.B) {
	mem := NewMemory([]byte{})
	// Pre-populate with some data
	for i := 0; i < 100; i++ {
		data := make([]byte, 1024)
		mem.PushOne(data)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mem.Count()
	}
}

// BenchmarkMemoryRange benchmarks iterating over buffers
func BenchmarkMemoryRange(b *testing.B) {
	mem := NewMemory([]byte{})
	// Pre-populate with some data
	for i := 0; i < 100; i++ {
		data := make([]byte, 1024)
		mem.PushOne(data)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mem.Range(func(buf []byte) {
			_ = len(buf)
		})
	}
}

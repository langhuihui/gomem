package gomem

import (
	"testing"
)

// BenchmarkMemoryReaderRead benchmarks reading from memory reader
func BenchmarkMemoryReaderRead(b *testing.B) {
	// Create test data
	buf1 := make([]byte, 1024)
	buf2 := make([]byte, 1024)
	buf3 := make([]byte, 1024)

	reader := NewReadableBuffersFromBytes(buf1, buf2, buf3)
	readBuf := make([]byte, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader = NewReadableBuffersFromBytes(buf1, buf2, buf3)
		for reader.Length > 0 {
			n, _ := reader.Read(readBuf)
			if n == 0 {
				break
			}
		}
	}
}

// BenchmarkMemoryReaderReadByte benchmarks reading single bytes
func BenchmarkMemoryReaderReadByte(b *testing.B) {
	// Create test data
	buf := make([]byte, 1024)
	reader := NewReadableBuffersFromBytes(buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader = NewReadableBuffersFromBytes(buf)
		for reader.Length > 0 {
			_, err := reader.ReadByte()
			if err != nil {
				break
			}
		}
	}
}

// BenchmarkMemoryReaderReadBytes benchmarks reading fixed-size bytes
func BenchmarkMemoryReaderReadBytes(b *testing.B) {
	// Create test data
	buf := make([]byte, 1024)
	reader := NewReadableBuffersFromBytes(buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader = NewReadableBuffersFromBytes(buf)
		for reader.Length > 0 {
			_, err := reader.ReadBytes(64)
			if err != nil {
				break
			}
		}
	}
}

// BenchmarkMemoryReaderReadBE benchmarks reading big-endian integers
func BenchmarkMemoryReaderReadBE(b *testing.B) {
	// Create test data
	buf := make([]byte, 1024)
	reader := NewReadableBuffersFromBytes(buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader = NewReadableBuffersFromBytes(buf)
		for reader.Length >= 4 {
			_, err := reader.ReadBE(4)
			if err != nil {
				break
			}
		}
	}
}

// BenchmarkMemoryReaderSkip benchmarks skipping bytes
func BenchmarkMemoryReaderSkip(b *testing.B) {
	// Create test data
	buf := make([]byte, 1024)
	reader := NewReadableBuffersFromBytes(buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader = NewReadableBuffersFromBytes(buf)
		for reader.Length > 0 {
			err := reader.Skip(64)
			if err != nil {
				break
			}
		}
	}
}

// BenchmarkMemoryReaderRange benchmarks iterating over buffers
func BenchmarkMemoryReaderRange(b *testing.B) {
	// Create test data
	buf1 := make([]byte, 512)
	buf2 := make([]byte, 512)
	reader := NewReadableBuffersFromBytes(buf1, buf2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader = NewReadableBuffersFromBytes(buf1, buf2)
		reader.Range(func(buf []byte) {
			_ = len(buf)
		})
	}
}

// BenchmarkMemoryReaderRangeN benchmarks iterating over N buffers
func BenchmarkMemoryReaderRangeN(b *testing.B) {
	// Create test data
	buf1 := make([]byte, 512)
	buf2 := make([]byte, 512)
	reader := NewReadableBuffersFromBytes(buf1, buf2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader = NewReadableBuffersFromBytes(buf1, buf2)
		reader.RangeN(2, func(buf []byte) {
			_ = len(buf)
		})
	}
}

// BenchmarkMemoryReaderLEB128Unmarshal benchmarks LEB128 unmarshaling
func BenchmarkMemoryReaderLEB128Unmarshal(b *testing.B) {
	// Create test data with LEB128 encoded numbers
	buf := make([]byte, 1024)
	// Fill with some LEB128 encoded data
	for i := 0; i < len(buf); i++ {
		buf[i] = byte(i % 128)
	}

	reader := NewReadableBuffersFromBytes(buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader = NewReadableBuffersFromBytes(buf)
		for reader.Length > 0 {
			_, _, err := reader.LEB128Unmarshal()
			if err != nil {
				break
			}
		}
	}
}

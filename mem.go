/*
Package gomem provides efficient memory management and buffer operations for Go applications.
It offers a Memory struct that can handle multiple byte buffers with optimized operations
for reading, writing, copying, and manipulating data across multiple buffer segments.

The Memory struct is designed for scenarios where you need to work with data that may be
split across multiple byte slices, such as network protocols, file I/O, or streaming data.
It provides methods for appending buffers, copying data, comparing memory contents,
and creating readers for sequential data access.

Key features:
- Efficient handling of multiple byte buffers
- Memory-safe operations with proper size tracking
- Support for negative indexing in buffer operations
- Integration with io.Writer and io.Reader interfaces
- Optimized copying and comparison operations
- Support for method chaining in append operations

Example usage:

	// Create a new Memory from a byte slice
	mem := NewMemory([]byte("Hello, "))

	// Append more data
	mem.PushOne([]byte("World!"))
	mem.Push([]byte(" This"), []byte(" is"), []byte(" a"), []byte(" test"))

	// Get all data as a single byte slice
	allData := mem.ToBytes()
	fmt.Println(string(allData)) // Output: "Hello, World! This is a test"

	// Create a reader for sequential access
	reader := mem.NewReader()
	buf := make([]byte, 5)
	reader.Read(buf)
	fmt.Println(string(buf)) // Output: "Hello"

	// Copy data to another Memory
	var mem2 Memory
	mem2.CopyFrom(&mem)

	// Check if memories are equal
	if mem.Equal(&mem2) {
		fmt.Println("Memories are equal")
	}

	// Reset memory
	mem.Reset()
	fmt.Println(mem.Size) // Output: 0
*/
package gomem

import (
	"io"
	"net"
	"slices"
)

const (
	MaxBlockSize   = 1 << 22
	BuddySize      = MaxBlockSize << 7
	MinPowerOf2    = 10
	defaultBufSize = 1 << 14
)

type Memory struct {
	Size    int
	Buffers [][]byte
}

// NewMemory creates a new Memory instance with the provided buffer.
// The buffer is wrapped in net.Buffers and the size is set to the buffer length.
func NewMemory(buf []byte) Memory {
	return Memory{
		Buffers: net.Buffers{buf},
		Size:    len(buf),
	}
}

// WriteTo writes all buffers in the Memory to the provided io.Writer.
// Returns the number of bytes written and any error that occurred.
func (m *Memory) WriteTo(w io.Writer) (n int64, err error) {
	copy := net.Buffers(slices.Clone(m.Buffers))
	return copy.WriteTo(w)
}

// Reset clears all buffers and resets the size to zero.
func (m *Memory) Reset() {
	m.Buffers = m.Buffers[:0]
	m.Size = 0
}

// UpdateBuffer updates the buffer at the specified index with the new buffer.
// Negative indices are supported (counting from the end).
// The total size is adjusted based on the difference in buffer lengths.
func (m *Memory) UpdateBuffer(index int, buf []byte) {
	if index < 0 {
		index = len(m.Buffers) + index
	}
	m.Size = len(buf) - len(m.Buffers[index])
	m.Buffers[index] = buf
}

// CopyFrom copies all data from the source Memory to this Memory.
// The data is copied into a new buffer and appended to this Memory.
func (m *Memory) CopyFrom(b *Memory) {
	buf := make([]byte, b.Size)
	b.CopyTo(buf)
	m.PushOne(buf)
}

// Equal compares this Memory with another Memory for equality.
// Returns true if both Memory instances have the same size, buffer count, and identical buffer contents.
func (m *Memory) Equal(b *Memory) bool {
	if m.Size != b.Size || len(m.Buffers) != len(b.Buffers) {
		return false
	}
	for i, buf := range m.Buffers {
		if !slices.Equal(buf, b.Buffers[i]) {
			return false
		}
	}
	return true
}

// CopyTo copies all buffer data into the provided destination buffer.
// The destination buffer must have sufficient capacity to hold all data.
func (m *Memory) CopyTo(buf []byte) {
	for _, b := range m.Buffers {
		l := len(b)
		copy(buf, b)
		buf = buf[l:]
	}
}

// ToBytes returns a new byte slice containing all data from all buffers.
// This creates a contiguous copy of all buffer data.
func (m *Memory) ToBytes() []byte {
	buf := make([]byte, m.Size)
	m.CopyTo(buf)
	return buf
}

// PushOne appends a single buffer to the Memory.
// The buffer is added to the Buffers slice and the total size is updated.
func (m *Memory) PushOne(b []byte) {
	m.Buffers = append(m.Buffers, b)
	m.Size += len(b)
}

// Push appends multiple buffers to the Memory.
// All provided buffers are added to the Buffers slice and the total size is updated.
func (m *Memory) Push(b ...[]byte) {
	m.Buffers = append(m.Buffers, b...)
	for _, level0 := range b {
		m.Size += len(level0)
	}
}

// Append appends all buffers from another Memory to this Memory.
// Returns the receiver for method chaining.
func (m *Memory) Append(mm Memory) *Memory {
	m.Buffers = append(m.Buffers, mm.Buffers...)
	m.Size += mm.Size
	return m
}

// Count returns the number of buffers in the Memory.
func (m *Memory) Count() int {
	return len(m.Buffers)
}

// Range iterates over all buffers in the Memory, calling the yield function for each buffer.
// This provides a way to process each buffer without exposing the internal structure.
func (m *Memory) Range(yield func([]byte)) {
	for i := range m.Count() {
		yield(m.Buffers[i])
	}
}

// NewReader creates a new MemoryReader for reading data from this Memory.
// The reader is initialized with the current Memory and its total size.
func (m *Memory) NewReader() MemoryReader {
	return MemoryReader{
		Memory: m,
		Length: m.Size,
	}
}

/*
Package gomem provides efficient memory management and buffer operations for Go applications.
This file contains the MemoryReader implementation, which provides sequential reading
capabilities for Memory structs with support for various data formats and reading patterns.

MemoryReader implements the io.Reader interface and extends it with additional functionality
for reading structured data, handling different byte orders, and managing reading state
across multiple buffer segments. It's designed for scenarios where you need to parse
or process data sequentially from a Memory struct.

Key features:
- Implements io.Reader interface for standard Go I/O compatibility
- Support for reading individual bytes, byte arrays, and structured data
- Big-endian and little-endian number reading capabilities
- LEB128 (Little Endian Base 128) variable-length integer decoding
- Efficient skipping and seeking operations
- Support for unreading (moving backwards in the data stream)
- Range operations for iterating over remaining data
- Memory clipping operations for removing processed data

Example usage:

	// Create a Memory with some data
	mem := NewMemory([]byte("Hello, World!"))
	reader := mem.NewReader()

	// Read individual bytes
	b1, _ := reader.ReadByte()
	b2, _ := reader.ReadByte()
	fmt.Printf("First two bytes: %c%c\n", b1, b2) // Output: "He"

	// Read multiple bytes at once
	buf := make([]byte, 5)
	n, _ := reader.Read(buf)
	fmt.Printf("Read %d bytes: %s\n", n, string(buf)) // Output: "Read 5 bytes: llo, "

	// Skip some bytes
	reader.Skip(2) // Skip " W"

	// Read remaining data
	remaining, _ := reader.ReadBytes(reader.Length)
	fmt.Printf("Remaining: %s\n", string(remaining)) // Output: "Remaining: orld!"

	// Example with LEB128 decoding
	leb128Data := []byte{0xE5, 0x8E, 0x26} // LEB128 encoded 624485
	reader2 := NewReadableBuffersFromBytes(leb128Data)
	value, bytesRead, _ := reader2.LEB128Unmarshal()
	fmt.Printf("LEB128 value: %d, bytes read: %d\n", value, bytesRead)

	// Example with big-endian reading
	beData := []byte{0x12, 0x34, 0x56, 0x78} // Big-endian 32-bit integer
	reader3 := NewReadableBuffersFromBytes(beData)
	beValue, _ := reader3.ReadBE(4)
	fmt.Printf("Big-endian value: 0x%X\n", beValue) // Output: "Big-endian value: 0x12345678"
*/
package gomem

import (
	"io"
	"slices"
)

type MemoryReader struct {
	*Memory
	Length, offset0, offset1 int
}

// NewReadableBuffersFromBytes creates a new MemoryReader from multiple byte slices.
// All provided byte slices are combined into a single Memory and wrapped in a MemoryReader.
func NewReadableBuffersFromBytes(b ...[]byte) MemoryReader {
	buf := &Memory{Buffers: b}
	for _, level0 := range b {
		buf.Size += len(level0)
	}
	return MemoryReader{Memory: buf, Length: buf.Size}
}

var _ io.Reader = (*MemoryReader)(nil)

// Offset returns the current reading position offset from the beginning.
// This represents how many bytes have been read from the Memory.
func (r *MemoryReader) Offset() int {
	return r.Size - r.Length
}

// Pop is not supported for MemoryReader and will panic if called.
// This method exists to satisfy interface requirements but should not be used.
func (r *MemoryReader) Pop() []byte {
	panic("ReadableBuffers Pop not allowed")
}

// GetCurrent returns the current buffer slice starting from the current reading position.
// This provides access to the remaining data in the current buffer.
func (r *MemoryReader) GetCurrent() []byte {
	return r.Memory.Buffers[r.offset0][r.offset1:]
}

// MoveToEnd moves the reading position to the end of all data.
// This effectively marks all data as read and sets the remaining length to zero.
func (r *MemoryReader) MoveToEnd() {
	r.offset0 = r.Count()
	r.offset1 = 0
	r.Length = 0
}

// Read implements the io.Reader interface, reading data into the provided buffer.
// Returns the number of bytes read and any error that occurred.
// Will return io.EOF when no more data is available.
func (r *MemoryReader) Read(buf []byte) (actual int, err error) {
	if r.Length == 0 {
		return 0, io.EOF
	}
	n := len(buf)
	curBuf := r.GetCurrent()
	curBufLen := len(curBuf)
	if n > r.Length {
		if curBufLen > 0 {
			actual += copy(buf, curBuf)
			r.offset0++
			r.offset1 = 0
		}
		for _, b := range r.Memory.Buffers[r.offset0:] {
			actual += copy(buf[actual:], b)
		}
		r.MoveToEnd()
		return
	}
	l := n
	for n > 0 {
		curBuf = r.GetCurrent()
		curBufLen = len(curBuf)
		if n < curBufLen {
			actual += n
			copy(buf[l-n:], curBuf[:n])
			r.forward(n)
			break
		}
		copy(buf[l-n:], curBuf)
		n -= curBufLen
		actual += curBufLen
		r.skipBuf()
		if r.Length == 0 && n > 0 {
			err = io.EOF
			return
		}
	}
	return
}

// ReadByteTo reads multiple bytes and stores them in the provided byte pointers.
// Returns an error if any byte cannot be read (including io.EOF).
func (r *MemoryReader) ReadByteTo(b ...*byte) (err error) {
	for i := range b {
		if r.Length == 0 {
			return io.EOF
		}
		*b[i], err = r.ReadByte()
		if err != nil {
			return
		}
	}
	return
}

// ReadByteMask reads a byte and applies a bit mask to it.
// Returns the masked byte value and any error that occurred during reading.
func (r *MemoryReader) ReadByteMask(mask byte) (byte, error) {
	b, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	return b & mask, nil
}

// ReadByte reads a single byte from the current position.
// Returns the byte value and any error that occurred (including io.EOF).
func (r *MemoryReader) ReadByte() (b byte, err error) {
	if r.Length == 0 {
		return 0, io.EOF
	}
	curBuf := r.GetCurrent()
	b = curBuf[0]
	if len(curBuf) == 1 {
		r.skipBuf()
	} else {
		r.forward(1)
	}
	return
}

// LEB128Unmarshal decodes a LEB128 (Little Endian Base 128) variable-length integer.
// Returns the decoded value, the number of bytes consumed, and any error.
// LEB128 is commonly used in binary protocols for encoding variable-length integers.
func (r *MemoryReader) LEB128Unmarshal() (uint, int, error) {
	v := uint(0)
	n := 0
	for i := 0; i < 8; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return 0, 0, err
		}
		v |= uint(b&0b01111111) << (i * 7)
		n++

		if (b & 0b10000000) == 0 {
			break
		}
	}

	return v, n, nil
}
func (r *MemoryReader) getCurrentBufLen() int {
	return len(r.Memory.Buffers[r.offset0]) - r.offset1
}

// Skip advances the reading position by n bytes without reading the data.
// Returns an error if trying to skip beyond the available data (io.EOF).
func (r *MemoryReader) Skip(n int) error {
	if n <= 0 {
		return nil
	}
	if n > r.Length {
		return io.EOF
	}
	curBufLen := r.getCurrentBufLen()
	for n > 0 {
		if n < curBufLen {
			r.forward(n)
			break
		}
		n -= curBufLen
		r.skipBuf()
		if r.Length == 0 && n > 0 {
			return io.EOF
		}
	}
	return nil
}

// Unread moves the reading position backwards by n bytes.
// This allows "unreading" data that was previously read.
func (r *MemoryReader) Unread(n int) {
	r.Length += n
	r.offset1 -= n
	for r.offset1 < 0 {
		r.offset0--
		r.offset1 += len(r.Memory.Buffers[r.offset0])
	}
}

func (r *MemoryReader) forward(n int) {
	r.Length -= n
	r.offset1 += n
}

func (r *MemoryReader) skipBuf() {
	curBufLen := r.getCurrentBufLen()
	r.Length -= curBufLen
	r.offset0++
	r.offset1 = 0
}

// ReadBytes reads exactly n bytes and returns them as a new byte slice.
// Returns an error if fewer than n bytes are available (io.EOF).
func (r *MemoryReader) ReadBytes(n int) ([]byte, error) {
	if n > r.Length {
		return nil, io.EOF
	}
	b := make([]byte, n)
	actual, err := r.Read(b)
	return b[:actual], err
}

// ReadBE reads n bytes in big-endian order and returns them as a uint32.
// The most significant byte is read first, followed by less significant bytes.
func (r *MemoryReader) ReadBE(n int) (num uint32, err error) {
	for i := range n {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		num += uint32(b) << ((n - i - 1) << 3)
	}
	return
}

// Range iterates over all remaining data, calling the yield function for each buffer segment.
// If yield is nil, it moves the reading position to the end without processing data.
func (r *MemoryReader) Range(yield func([]byte)) {
	if yield != nil {
		for r.Length > 0 {
			yield(r.GetCurrent())
			r.skipBuf()
		}
	} else {
		r.MoveToEnd()
	}
}

// RangeN iterates over up to n bytes of remaining data, calling the yield function for each buffer segment.
// Stops when n bytes have been processed or no more data is available.
func (r *MemoryReader) RangeN(n int, yield func([]byte)) {
	for good := yield != nil; r.Length > 0 && n > 0; r.skipBuf() {
		curBuf := r.GetCurrent()
		if curBufLen := len(curBuf); curBufLen > n {
			if r.forward(n); good {
				yield(curBuf[:n])
			}
			return
		} else if n -= curBufLen; good {
			yield(curBuf)
		}
	}
}

// ClipFront removes the already-read data from the front of the Memory.
// The yield function is called for each buffer segment being removed.
// This is useful for memory management when you want to free up processed data.
func (r *MemoryReader) ClipFront(yield func([]byte) bool) {
	offset := r.Size - r.Length
	if offset == 0 {
		return
	}
	if m := r.Memory; r.Length == 0 {
		for _, buf := range m.Buffers {
			yield(buf)
		}
		m.Buffers = m.Buffers[:0]
	} else {
		for _, buf := range m.Buffers[:r.offset0] {
			yield(buf)
		}
		if r.offset1 > 0 {
			yield(m.Buffers[r.offset0][:r.offset1])
			m.Buffers[r.offset0] = r.GetCurrent()
		}
		if r.offset0 > 0 {
			m.Buffers = slices.Delete(m.Buffers, 0, r.offset0)
		}
	}
	r.Size -= offset
	r.offset0 = 0
	r.offset1 = 0
}

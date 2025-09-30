package gomem

import (
	"testing"
)

func TestMemory(t *testing.T) {
	// Test NewMemory
	buf := []byte{1, 2, 3, 4, 5}
	mem := NewMemory(buf)
	if mem.Size != len(buf) {
		t.Errorf("Expected size %d, got %d", len(buf), mem.Size)
	}
	if len(mem.Buffers) != 1 {
		t.Errorf("Expected 1 buffer, got %d", len(mem.Buffers))
	}

	// Test PushOne
	mem.PushOne([]byte{6, 7})
	if mem.Size != 7 {
		t.Errorf("Expected size 7, got %d", mem.Size)
	}
	if len(mem.Buffers) != 2 {
		t.Errorf("Expected 2 buffers, got %d", len(mem.Buffers))
	}

	// Test Reset
	mem.Reset()
	if mem.Size != 0 {
		t.Errorf("Expected size 0, got %d", mem.Size)
	}
	if len(mem.Buffers) != 0 {
		t.Errorf("Expected 0 buffers, got %d", len(mem.Buffers))
	}
}

func TestMemoryReader(t *testing.T) {
	// Test NewReadableBuffersFromBytes
	buf1 := []byte{1, 2, 3}
	buf2 := []byte{4, 5, 6}
	reader := NewReadableBuffersFromBytes(buf1, buf2)

	if reader.Length != 6 {
		t.Errorf("Expected length 6, got %d", reader.Length)
	}

	// Test Read
	result := make([]byte, 6)
	n, err := reader.Read(result)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if n != 6 {
		t.Errorf("Expected to read 6 bytes, got %d", n)
	}
	expected := []byte{1, 2, 3, 4, 5, 6}
	for i, b := range result {
		if b != expected[i] {
			t.Errorf("Expected %d at position %d, got %d", expected[i], i, b)
		}
	}
}

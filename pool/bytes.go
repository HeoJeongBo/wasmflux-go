package pool

// ByteBuffer is a reusable byte buffer from the pool.
type ByteBuffer struct {
	B []byte
}

// Reset clears the buffer without releasing underlying memory.
func (b *ByteBuffer) Reset() {
	b.B = b.B[:0]
}

// Write appends bytes to the buffer.
func (b *ByteBuffer) Write(p []byte) (int, error) {
	b.B = append(b.B, p...)
	return len(p), nil
}

// WriteString appends a string to the buffer.
func (b *ByteBuffer) WriteString(s string) {
	b.B = append(b.B, s...)
}

// WriteByte appends a single byte.
func (b *ByteBuffer) WriteByte(c byte) error {
	b.B = append(b.B, c)
	return nil
}

// Len returns the current length of the buffer content.
func (b *ByteBuffer) Len() int {
	return len(b.B)
}

// String returns the buffer content as a string.
func (b *ByteBuffer) String() string {
	return string(b.B)
}

// Bytes is a pool of reusable byte buffers.
// Default initial capacity is 256 bytes.
var Bytes = New(func() *ByteBuffer {
	return &ByteBuffer{B: make([]byte, 0, 256)}
})

// GetBuffer retrieves a byte buffer from the pool, reset and ready to use.
func GetBuffer() *ByteBuffer {
	buf := Bytes.Get()
	buf.Reset()
	return buf
}

// PutBuffer returns a byte buffer to the pool.
// Buffers larger than 64KB are discarded to prevent pool bloat.
func PutBuffer(buf *ByteBuffer) {
	if cap(buf.B) > 65536 {
		return
	}
	Bytes.Put(buf)
}

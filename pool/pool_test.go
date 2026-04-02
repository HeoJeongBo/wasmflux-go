package pool

import "testing"

func TestPool_GetPut(t *testing.T) {
	p := New(func() *int {
		v := 42
		return &v
	})

	v := p.Get()
	if *v != 42 {
		t.Errorf("got %d, want 42", *v)
	}

	*v = 100
	p.Put(v)

	v2 := p.Get()
	// May or may not be the same pointer (sync.Pool makes no guarantees),
	// but the constructor should work.
	if v2 == nil {
		t.Error("got nil from pool")
	}
}

func TestByteBuffer(t *testing.T) {
	buf := GetBuffer()
	defer PutBuffer(buf)

	buf.WriteString("hello")
	buf.WriteByte(' ')
	buf.WriteString("world")

	if buf.String() != "hello world" {
		t.Errorf("got %q, want %q", buf.String(), "hello world")
	}
	if buf.Len() != 11 {
		t.Errorf("len: %d, want 11", buf.Len())
	}
}

func TestByteBuffer_Reset(t *testing.T) {
	buf := GetBuffer()
	defer PutBuffer(buf)

	buf.WriteString("data")
	buf.Reset()

	if buf.Len() != 0 {
		t.Errorf("len after reset: %d, want 0", buf.Len())
	}
}

func TestByteBuffer_Write(t *testing.T) {
	buf := GetBuffer()
	defer PutBuffer(buf)

	n, err := buf.Write([]byte("test"))
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if n != 4 {
		t.Errorf("n: %d, want 4", n)
	}
	if buf.String() != "test" {
		t.Errorf("got %q", buf.String())
	}
}

func TestPutBuffer_LargeDiscard(t *testing.T) {
	buf := &ByteBuffer{B: make([]byte, 0, 100_000)} // > 64KB
	buf.WriteString("data")
	PutBuffer(buf)
	// Can't directly verify discard, but ensure it doesn't panic.
	// The pool may or may not return this buffer.
}

func TestPool_ConcurrentGetPut(t *testing.T) {
	p := New(func() int { return 42 })

	done := make(chan struct{}, 100)
	for i := 0; i < 100; i++ {
		go func() {
			v := p.Get()
			p.Put(v)
			done <- struct{}{}
		}()
	}
	for i := 0; i < 100; i++ {
		<-done
	}
}

func BenchmarkByteBuffer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := GetBuffer()
		buf.WriteString("key=value")
		PutBuffer(buf)
	}
}

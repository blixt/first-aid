package syncbuffer_test

import (
	"bytes"
	"io"
	"sync"
	"testing"

	"github.com/blixt/first-aid/syncbuffer"
)

func TestBasicWriteAndRead(t *testing.T) {
	sb := syncbuffer.New(10)
	data := []byte("hello")

	n, err := sb.Write(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(data) {
		t.Fatalf("expected to write %d bytes, wrote %d", len(data), n)
	}

	readBuf := make([]byte, len(data))
	n, err = sb.Read(readBuf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(data) {
		t.Fatalf("expected to read %d bytes, read %d", len(data), n)
	}
	if !bytes.Equal(data, readBuf) {
		t.Fatalf("expected %q, got %q", data, readBuf)
	}
}

func TestConcurrency(t *testing.T) {
	sb := syncbuffer.New(100)
	data := []byte("concurrency test data")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_, err := sb.Write(data)
			if err != nil {
				t.Errorf("write error: %v", err)
			}
		}
	}()

	go func() {
		defer wg.Done()
		readBuf := make([]byte, len(data))
		for i := 0; i < 100; i++ {
			_, err := sb.Read(readBuf)
			if err != nil && err != io.EOF {
				t.Errorf("read error: %v", err)
			}
			if !bytes.Equal(data, readBuf) {
				t.Errorf("iteration %d: expected %q, got %q", i, data, readBuf)
			}
		}
	}()

	wg.Wait()
}

func TestClosedBuffer(t *testing.T) {
	sb := syncbuffer.New(10)
	sb.Close()

	_, err := sb.Write([]byte("data"))
	if err != io.ErrClosedPipe {
		t.Errorf("expected error %v, got %v", io.ErrClosedPipe, err)
	}

	readBuf := make([]byte, 4)
	_, err = sb.Read(readBuf)
	if err != io.EOF {
		t.Errorf("expected error %v, got %v", io.EOF, err)
	}
}

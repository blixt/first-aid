package syncbuffer

import (
	"io"
	"sync"
)

// SyncBuffer is a thread-safe circular buffer.
type SyncBuffer struct {
	buf    []byte
	size   int
	rpos   int // read position
	wpos   int // write position
	mu     sync.Mutex
	cond   *sync.Cond
	closed bool
}

// New creates a new SyncBuffer with the given size.
func New(size int) *SyncBuffer {
	// Create the buffer with one extra byte since rpos and wpos can never be
	// the same while reading and writing, meaning in reality that byte position
	// cannot be used.
	size++
	sb := &SyncBuffer{
		buf:  make([]byte, size),
		size: size,
	}
	sb.cond = sync.NewCond(&sb.mu)
	return sb
}

func (sb *SyncBuffer) Size() int {
	return sb.size - 1
}

func (sb *SyncBuffer) Free() int {
	return sb.size - (sb.wpos-sb.rpos+sb.size)%sb.size - 1
}

func (sb *SyncBuffer) Used() int {
	return (sb.wpos - sb.rpos + sb.size) % sb.size
}

// Write writes data to the buffer.
func (sb *SyncBuffer) Write(p []byte) (int, error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.closed {
		return 0, io.ErrClosedPipe
	}

	n := len(p)
	totalWritten := 0

	for totalWritten < n {
		if sb.closed {
			return totalWritten, io.ErrClosedPipe
		}

		spaceLeft := sb.size - (sb.wpos-sb.rpos+sb.size)%sb.size - 1
		if spaceLeft == 0 {
			sb.cond.Wait()
			continue
		}

		toWrite := n - totalWritten
		if toWrite > spaceLeft {
			toWrite = spaceLeft
		}

		end := sb.wpos + toWrite
		if end <= sb.size {
			copy(sb.buf[sb.wpos:end], p[totalWritten:totalWritten+toWrite])
		} else {
			part1 := sb.size - sb.wpos
			part2 := toWrite - part1
			copy(sb.buf[sb.wpos:], p[totalWritten:totalWritten+part1])
			copy(sb.buf[0:part2], p[totalWritten+part1:totalWritten+toWrite])
		}

		sb.wpos = (sb.wpos + toWrite) % sb.size
		totalWritten += toWrite
		sb.cond.Broadcast()
	}

	return totalWritten, nil
}

// Read reads data from the buffer.
func (sb *SyncBuffer) Read(p []byte) (int, error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	for sb.rpos == sb.wpos && !sb.closed {
		sb.cond.Wait()
	}

	if sb.rpos == sb.wpos && sb.closed {
		return 0, io.EOF
	}

	totalRead := 0
	n := len(p)
	for totalRead < n && (sb.rpos != sb.wpos || !sb.closed) {
		toRead := n - totalRead

		availableBytes := (sb.wpos - sb.rpos + sb.size) % sb.size
		if availableBytes == 0 {
			sb.cond.Wait()
			continue
		}

		if toRead > availableBytes {
			toRead = availableBytes
		}

		end := sb.rpos + toRead
		if end <= sb.size {
			copy(p[totalRead:totalRead+toRead], sb.buf[sb.rpos:end])
		} else {
			part1 := sb.size - sb.rpos
			part2 := toRead - part1
			copy(p[totalRead:totalRead+part1], sb.buf[sb.rpos:])
			copy(p[totalRead+part1:totalRead+toRead], sb.buf[0:part2])
		}

		sb.rpos = (sb.rpos + toRead) % sb.size
		totalRead += toRead
		sb.cond.Broadcast()
	}

	if totalRead == 0 && sb.closed {
		return 0, io.EOF
	}

	return totalRead, nil
}

// Close closes the buffer from being written to, and makes future Reads return EOF.
func (sb *SyncBuffer) Close() error {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.closed = true
	sb.cond.Broadcast()
	return nil
}

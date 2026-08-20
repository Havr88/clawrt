package sys

import (
	"bytes"
	"sync"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// GetBuffer acquires a clean bytes.Buffer from the pool
func GetBuffer() *bytes.Buffer {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// PutBuffer returns the buffer to the pool
func PutBuffer(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	// Do not keep extremely large buffers in pool to avoid memory leaks on embedded devices
	if buf.Cap() > 64*1024 {
		return
	}
	bufPool.Put(buf)
}

package contract

import (
	"bytes"
	"sync"
)

var jsonBufferPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, 1024))
	},
}

func getJSONBuffer() *bytes.Buffer {
	return jsonBufferPool.Get().(*bytes.Buffer)
}

func putJSONBuffer(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	buf.Reset()
	jsonBufferPool.Put(buf)
}

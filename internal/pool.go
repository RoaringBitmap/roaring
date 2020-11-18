package internal

import (
	"sync"
)

var (
	ByteBufferPool = sync.Pool{
		New: func() interface{} {
			return &ByteBuffer{}
		},
	}

	ByteInputAdapterPool = sync.Pool{
		New: func() interface{} {
			return &ByteInputAdapter{}
		},
	}
)

package grasure

import (
	"io"
)

type NodeWriter struct {
	writers io.WriteCloser
	offset  uint64
	size    uint64
}

//ParallelReader is a writer that write data bytes to multiple nodes
type ParallelWriter struct {
	readers    io.ReadCloser
	remoteAddr string
}

//ParallelReader is a reader that handles parallel disk reads
type ParallelReader struct {
	readers      []io.ReaderAt
	dataShards   int
	parityShards int
	offset       uint64
	size         uint64
	paraBuf      [][]byte
	buf          []byte
	degrade      bool
}

func NewParallelReader(pool *ErasurePool, offset, size uint64, degrade bool) *ParallelReader {
	return &ParallelReader{
		dataShards:   pool.K,
		parityShards: pool.M,
		readers:      make([]io.ReaderAt, pool.K+pool.M),
		offset:       offset,
		size:         size,
		paraBuf:      nil,
		buf:          nil,
		degrade:      degrade,
	}
}

func (pr *ParallelReader) ReadBlock(offset, size uint64) {

}

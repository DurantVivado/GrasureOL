package grasure

import (
	"context"
	"io"
)

//Customer APIs
type StorageAPI interface {
	ListDir(ctx context.Context, dirPath string, level int) ([]string, error)
	ReadFile(ctx context.Context, path string, offset, size int64, buf []byte) (n int64, err error)
	ReadFileStream(ctx context.Context, path string, offset, length int64) (io.ReadCloser, error)

	WriteFile(ctx context.Context, path string, size int64, buf []byte) error
	WriteFileStream(ctx context.Context, path string, size int64, reader io.Reader) error
	RenameFile(ctx context.Context, srcPath, dstPath string) error

	Delete(ctx context.Context, path string, recursive bool) (err error)
}
package filemanager

import "io"

// writerAtAdapter adapts an io.WriterAt to an io.Writer
type writerAtAdapter struct {
	w      io.WriterAt
	offset int64
}

func (a *writerAtAdapter) Write(p []byte) (int, error) {
	n, err := a.w.WriteAt(p, a.offset)
	a.offset += int64(n)
	return n, err
}

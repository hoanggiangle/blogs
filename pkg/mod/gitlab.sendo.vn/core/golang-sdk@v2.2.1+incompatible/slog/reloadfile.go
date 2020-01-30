package slog

import (
	"os"
	"sync/atomic"
	"unsafe"
)

// a simple file that have reload for logrotate
//
// used unsafe.Pointer to work around race bug
type reloadFile struct {
	pFile unsafe.Pointer
	fname string
}

func newReloadFile(path string) (*reloadFile, error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &reloadFile{
		pFile: unsafe.Pointer(f),
		fname: path,
	}, nil
}

func (r *reloadFile) ReOpen() error {
	f, err := os.OpenFile(r.fname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

	pOld := atomic.SwapPointer(&r.pFile, unsafe.Pointer(f))
	f = (*os.File)(pOld)
	f.Close()

	return err
}

func (r *reloadFile) Write(p []byte) (int, error) {
	pFile := atomic.LoadPointer(&r.pFile)
	f := (*os.File)(pFile)
	return f.Write(p)
}

func (r *reloadFile) Sync() {
	pFile := atomic.LoadPointer(&r.pFile)
	f := (*os.File)(pFile)
	f.Sync()
}

func (r *reloadFile) Close() {
	pFile := atomic.LoadPointer(&r.pFile)
	f := (*os.File)(pFile)
	f.Close()
}

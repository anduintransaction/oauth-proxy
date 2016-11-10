package packer

import (
	"bytes"
	"errors"
	"io"
	"os"
	"time"
)

var (
	errBadFileDescriptor = errors.New("bad file descriptor")
	errInvalidDirectory  = errors.New("not a folder")
	errNotFound          = errors.New("no such file or directory")
)

type File interface {
	io.Reader
	io.Seeker
	io.Closer
	Name() string
	Stat() (os.FileInfo, error)
}

// A FileInfo describes a file, and implements os.FileInfo interface.
type FileInfo struct {
	FileName    string
	FileSize    int64
	FileMode    os.FileMode
	FileModTime time.Time
}

// Name returns base name of the file.
func (fi *FileInfo) Name() string {
	return fi.FileName
}

// Size returns length in bytes for regular files; system-dependent for others.
func (fi *FileInfo) Size() int64 {
	return fi.FileSize
}

// Mode returns file mode bits
func (fi *FileInfo) Mode() os.FileMode {
	return fi.FileMode
}

// ModTime returns modification time
func (fi *FileInfo) ModTime() time.Time {
	return fi.FileModTime
}

// IsDir is abbreviation for Mode().IsDir()
func (fi *FileInfo) IsDir() bool {
	return fi.FileMode.IsDir()
}

// Sys is always nil
func (fi *FileInfo) Sys() interface{} {
	return nil
}

type file struct {
	data   []byte
	r      *bytes.Reader
	stat   os.FileInfo
	closed bool
}

// NewPackerFile creates a File. Use internally by generator.
func NewPackerFile(data []byte, info os.FileInfo) File {
	return &file{
		data: data,
		r:    bytes.NewReader(data),
		stat: info,
	}
}

func (f *file) Clone() *file {
	return &file{
		data: f.data,
		r:    bytes.NewReader(f.data),
		stat: f.stat,
	}
}

func (f *file) Close() error {
	f.closed = true
	return nil
}

func (f *file) Name() string {
	return f.stat.Name()
}

func (f *file) Read(b []byte) (n int, err error) {
	if f.closed {
		return 0, &os.PathError{"read", f.Name(), errBadFileDescriptor}
	}
	return f.r.Read(b)
}

func (f *file) Seek(offset int64, whence int) (ret int64, err error) {
	if f.closed {
		return 0, &os.PathError{"seek", f.Name(), errBadFileDescriptor}
	}
	return f.r.Seek(offset, whence)
}

func (f *file) Stat() (fi os.FileInfo, err error) {
	if f.closed {
		return nil, &os.PathError{"stat", f.Name(), errBadFileDescriptor}
	}
	return f.stat, nil
}

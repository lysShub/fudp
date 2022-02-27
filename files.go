package fudp

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"go.uber.org/atomic"
)

type file struct {
	fh   *os.File
	name string // 分割符为"/"
	size int64
	mode os.FileMode

	// 文件状态
	stat *atomic.Uint32

	back *file
	next *file
}

func (f *file) add() {
	f.stat.Store(f.stat.Load()<<1 + 1)
}

func (f *fudp) prepare() (err error) {
	if f.wpath, err = filepath.Abs(f.wpath); err != nil {
		return err
	}

	defer func() {
		if f.files == nil {
			err = errors.New("can't find any readable file")
		} else { // reset f.files head
			for {
				if f.files.back != nil {
					f.files = f.files.back
				} else {
					break
				}
			}
		}
	}()

	return filepath.Walk(f.wpath, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			if err == nil {
				if fh, e := os.Open(path); e == nil {
					name, _ := filepath.Rel(f.wpath, path)
					name = filepath.ToSlash(name)

					if f.files == nil {
						f.files = &file{fh: fh, size: info.Size(), name: name}
					} else {
						f.files.next = &file{
							fh:   fh,
							name: name,
							size: info.Size(),
							back: f.files,
						}
						f.files = f.files.next
					}
					return nil
				} else {
					err = e
				}
			}

			if f.strict {
				return err
			}
		}
		return nil
	})
}

func (f *fudp) getFile() (file *file) {
	f.Lock()
	defer func() {
		if file != nil {
			file.add()
		}
		f.Unlock()
	}()

	for {
		if f.files.stat.Load()&0b1 == 0 {
			return f.files
		}
		if f.files.next != nil {
			f.files = f.files.next
		} else {
			break
		}
	}

	for {
		if f.files.stat.Load()&0b1 == 0 {
			return f.files
		}
		if f.files.back != nil {
			f.files = f.files.next
		} else {
			break
		}
	}
	return nil
}

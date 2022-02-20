package fudp

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

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
						f.files = &file{
							fh:   fh,
							name: name,
						}
					} else {
						f.files.next = &file{fh: fh, name: name, back: f.files}
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

type file struct {
	fh   *os.File
	name string // 分割符为"/"

	back *file
	next *file
}

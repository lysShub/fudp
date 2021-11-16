package fudp

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

func verifyPath(path string, isSend bool) error {
	if isSend {
		fi, err := os.Stat(path)
		if os.IsNotExist(err) {
			return errors.New("invalid path: not exist")
		} else {
			if !fi.IsDir() {
				if fi.Size() == 0 {
					return errors.New("invalid path: file empty")
				}
			} else {
				var s int64
				filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
					s = s + info.Size()
					if s > 0 {
						return errors.New("null")
					}
					return nil
				})
				if s == 0 {
					return errors.New("invalid path: path empty")
				}
			}
		}
	} else {
		fi, err := os.Stat(path)
		if os.IsNotExist(err) {
			return os.MkdirAll(path, 0666)
		} else if !fi.IsDir() {
			return errors.New("invalid path: is file path, expcet floder path")
		}
	}
	return nil
}

func formatPath(path string) string {

	var p rune = os.PathSeparator
	switch p {
	case '\\':
		path = filepath.FromSlash(path)
	case '/':
		path = filepath.ToSlash(path)
	default:
		path = ""
	}

	path, err := filepath.Abs(path)
	if err != nil {
		path = getExePath()
	}
	return path
}

// getExePath 获取可执行文件路径
func getExePath() string {
	ex, err := os.Executable()
	if err != nil {
		exReal, err := filepath.EvalSymlinks(ex)
		if err != nil {
			dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
			if err != nil {
				return "./"
			}
			return dir
		}
		return filepath.Dir(exReal)
	}
	return filepath.Dir(ex)
}

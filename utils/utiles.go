package utils

import (
	"os"
	"path/filepath"
)

// GetExePath 获取可执行文件路径
func GetExePath() string {
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

// GetExeName 格式化为当前系统的路径格式
func FormatPath(path string) string {
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
		path = GetExePath()
	}
	return path
}

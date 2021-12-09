package log

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"
)

var instance string

func init() {
	instance = strconv.Itoa(int(time.Now().Unix())) + "    "
}

var LogOut, WranOut, ErrorOut *os.File = os.Stdout, os.Stdout, os.Stderr

// Redirect 重定向日志输出
func Redirect(log, warn, err *os.File) error {
	if log != nil {
		if n, err := log.Write([]byte{0}); err != nil || n != 0 {
			return err
		}
		LogOut = log
	}
	if warn != nil {
		if n, err := warn.Write([]byte{0}); err != nil || n != 0 {
			return err
		}
		WranOut = warn
	}
	if err != nil {
		if n, err := err.Write([]byte{0}); err != nil || n != 0 {
			return err
		}
		ErrorOut = err
	}
	return nil
}

// Log 记录日志信息
// 	@msg: 日志信息
//  当msg不为nil时返回true
func Log(msg error) error {

	if msg == nil {
		return nil
	} else {
		_, fp, ln, _ := runtime.Caller(1)
		fmt.Fprintln(LogOut, time.Now().Format(time.RFC3339), instance+"Logging    "+fp+":"+strconv.Itoa(ln)+"    "+msg.Error())

		// fmt.Fprintln(LogOut, time.Now().Format(time.RFC3339), instance+"Logging    "+msg.Error()) // 打包时使用, 此时日志不包含路径信息

		return msg
	}
}

// Warn 记录警告信息
// 	@msg: 警告信息
// 当war不为nil时返回true
func Warn(war error) error {

	if war == nil {
		return nil
	} else {
		_, fp, ln, _ := runtime.Caller(1)
		fmt.Fprintln(WranOut, time.Now().Format(time.RFC3339), instance+"Warning    "+fp+":"+strconv.Itoa(ln)+"    "+war.Error())

		// fmt.Fprintln(WranOut, time.Now().Format(time.RFC3339), instance+"Warning    "+msg.Error()) // 打包时使用, 此时日志不包含路径信息

		return war
	}
}

// Error 记录错误信息
// 	@err: 错误
//  当msg不为nil时返回true
func Error(err error) error {
	if err == nil {
		return nil
	} else {
		_, fp, ln, _ := runtime.Caller(1)
		fmt.Fprintln(ErrorOut, time.Now().Format(time.RFC3339), instance+"Error    "+fp+":"+strconv.Itoa(ln)+"    "+err.Error())

		// fmt.Fprintln(ErrorOut, time.Now().Format(time.RFC3339), instance+"Error    "+msg.Error()) // 打包时使用, 此时日志不包含路径信息

		return err
	}
}

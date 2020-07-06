package utils

import (
	"fmt"
	"runtime/debug"
)

func TryCatch() {
	if r := recover(); r != nil {
		TLog.Error(fmt.Sprintf("%v\r\n", r))
		TLog.Error(string(debug.Stack()[:]))
	}
}

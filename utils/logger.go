package utils

import (
	"fmt"
	"github.com/kataras/golog"
)

var _logger *golog.Logger

func SetLogger(logger *golog.Logger) {
	_logger = logger
}

func GetLogger() *golog.Logger {
	if _logger == nil {
		fmt.Println("[WARN] Logger not set")
	}

	return _logger
}

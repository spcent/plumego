package commands

import (
	"flag"
	"io"
	"strings"
)

// AppError 自定义错误类型
type AppError struct {
	ErrorCode int
	Message   string
	Detail    string
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	return e.Message
}

// Code 返回错误码
func (e *AppError) Code() int {
	return e.ErrorCode
}

// containsHelpFlag checks if the arguments contain help flags
func containsHelpFlag(args []string) bool {
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

// ParseFlags 解析命令行参数，返回剩余的位置参数
func ParseFlags(name string, args []string, fn func(*flag.FlagSet)) ([]string, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	fn(fs)

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	return fs.Args(), nil
}

// NewAppError 创建新的应用错误
func NewAppError(code int, message string, detail string) *AppError {
	return &AppError{
		ErrorCode: code,
		Message:   message,
		Detail:    detail,
	}
}

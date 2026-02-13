// Package builtin - 脚本引擎内置对象和模块
// 提供丰富的内置对象供 JavaScript 脚本使用

package builtin

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dop251/goja"
)

// LogEntry 日志条目
type LogEntry struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Time    int64  `json:"time"`
}

// InjectBuiltins 注入内置对象到 VM
func InjectBuiltins(vm *goja.Runtime, logs *[]LogEntry) {
	injectConsole(vm, logs)
	injectJSON(vm)
	injectDate(vm)
	injectSleep(vm)
}

// injectConsole 注入 console 对象
func injectConsole(vm *goja.Runtime, logs *[]LogEntry) {
	console := map[string]interface{}{
		"log": func(args ...interface{}) {
			*logs = append(*logs, LogEntry{
				Level:   "info",
				Message: fmt.Sprint(args...),
				Time:    time.Now().Unix(),
			})
		},
		"debug": func(args ...interface{}) {
			*logs = append(*logs, LogEntry{
				Level:   "debug",
				Message: fmt.Sprint(args...),
				Time:    time.Now().Unix(),
			})
		},
		"info": func(args ...interface{}) {
			*logs = append(*logs, LogEntry{
				Level:   "info",
				Message: fmt.Sprint(args...),
				Time:    time.Now().Unix(),
			})
		},
		"warn": func(args ...interface{}) {
			*logs = append(*logs, LogEntry{
				Level:   "warn",
				Message: fmt.Sprint(args...),
				Time:    time.Now().Unix(),
			})
		},
		"error": func(args ...interface{}) {
			*logs = append(*logs, LogEntry{
				Level:   "error",
				Message: fmt.Sprint(args...),
				Time:    time.Now().Unix(),
			})
		},
	}
	vm.Set("console", console)
}

// injectJSON 注入 JSON 辅助函数
func injectJSON(vm *goja.Runtime) {
	vm.Set("JSON", map[string]interface{}{
		"parse": func(s string) (interface{}, error) {
			var v interface{}
			err := json.Unmarshal([]byte(s), &v)
			return v, err
		},
		"stringify": func(v interface{}) (string, error) {
			b, err := json.Marshal(v)
			return string(b), err
		},
	})
}

// injectDate 注入 Date 对象
func injectDate(vm *goja.Runtime) {
	vm.Set("Date", map[string]interface{}{
		"now":   time.Now().UnixMilli,
		"parse": time.Parse,
	})
}

// injectSleep 注入 sleep 函数
func injectSleep(vm *goja.Runtime) {
	vm.Set("sleep", func(ms int) {
		time.Sleep(time.Duration(ms) * time.Millisecond)
	})
}

// Package builtin - 模块接口定义

package builtin

import "github.com/dop251/goja"

// Module 脚本模块接口
type Module interface {
	// Name 模块名称
	Name() string
	// Register 注册模块到 VM
	Register(vm *goja.Runtime) error
}

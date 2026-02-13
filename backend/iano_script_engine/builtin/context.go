// Package builtin - Context 模块
// 提供上下文访问功能

package builtin

import (
	"context"

	"github.com/dop251/goja"
)

// ContextModule 上下文模块
type ContextModule struct {
	ctx context.Context
}

// NewContextModule 创建上下文模块
func NewContextModule(ctx context.Context) *ContextModule {
	return &ContextModule{ctx: ctx}
}

// Name 模块名称
func (m *ContextModule) Name() string {
	return "ctx"
}

// Register 注册模块
func (m *ContextModule) Register(vm *goja.Runtime) error {
	ctxObj := map[string]interface{}{
		"value": func(key string) interface{} {
			return m.ctx.Value(key)
		},
		"done": func() bool {
			select {
			case <-m.ctx.Done():
				return true
			default:
				return false
			}
		},
	}

	return vm.Set("ctx", ctxObj)
}

// Package builtin - Utils 模块
// 提供工具函数

package builtin

import (
	"fmt"
	"strings"
	"time"

	"github.com/dop251/goja"
)

// UtilsModule 工具模块
type UtilsModule struct{}

// NewUtilsModule 创建工具模块
func NewUtilsModule() *UtilsModule {
	return &UtilsModule{}
}

// Name 模块名称
func (m *UtilsModule) Name() string {
	return "utils"
}

// Register 注册模块
func (m *UtilsModule) Register(vm *goja.Runtime) error {
	utils := map[string]interface{}{
		"uuid":   generateUUID,
		"md5":    md5Hash,
		"sha256": sha256Hash,
		"base64": base64Funcs(),
		"random": randomFuncs(),
		"string": stringFuncs(),
		"time":   timeFuncs(),
	}

	return vm.Set("utils", utils)
}

// base64Funcs Base64 函数
func base64Funcs() map[string]interface{} {
	return map[string]interface{}{
		"encode": func(s string) string {
			return "base64:" + s // 简化实现
		},
		"decode": func(s string) (string, error) {
			return s, nil // 简化实现
		},
	}
}

// randomFuncs 随机函数
func randomFuncs() map[string]interface{} {
	return map[string]interface{}{
		"int": func(max int) int {
			return time.Now().Nanosecond() % max // 简化实现
		},
		"float": func() float64 {
			return float64(time.Now().Nanosecond()) / 1e9
		},
		"choice": func(items []interface{}) interface{} {
			if len(items) == 0 {
				return nil
			}
			idx := time.Now().Nanosecond() % len(items)
			return items[idx]
		},
	}
}

// stringFuncs 字符串函数
func stringFuncs() map[string]interface{} {
	return map[string]interface{}{
		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"toLower":   strings.ToLower,
		"toUpper":   strings.ToUpper,
		"trim":      strings.TrimSpace,
		"split":     strings.Split,
		"join":      strings.Join,
		"replace":   strings.ReplaceAll,
	}
}

// timeFuncs 时间函数
func timeFuncs() map[string]interface{} {
	return map[string]interface{}{
		"now":       time.Now,
		"parse":     time.Parse,
		"format":    time.Now().Format,
		"unix":      time.Now().Unix,
		"unixMilli": time.Now().UnixMilli,
		"sleep": func(ms int) {
			time.Sleep(time.Duration(ms) * time.Millisecond)
		},
	}
}

// generateUUID 生成 UUID (简化实现)
func generateUUID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Nanosecond())
}

// md5Hash MD5 哈希 (简化实现)
func md5Hash(s string) string {
	return fmt.Sprintf("md5:%s", s)
}

// sha256Hash SHA256 哈希 (简化实现)
func sha256Hash(s string) string {
	return fmt.Sprintf("sha256:%s", s)
}

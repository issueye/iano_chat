package web

import (
	"net/http"
	"sync"
)

// contextPool Context 对象池
var contextPool = sync.Pool{
	New: func() interface{} {
		return &Context{
			Params: make(map[string]string),
			data:   make(map[string]interface{}),
		}
	},
}

// paramsPool Params map 池
var paramsPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]string)
	},
}

// dataMapPool data map 池
var dataMapPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]interface{})
	},
}

// handlersPool handlers 切片池
var handlersPool = sync.Pool{
	New: func() interface{} {
		return make([]HandlerFunc, 0, 8)
	},
}

// acquireContext 从池中获取 Context
func acquireContext(w http.ResponseWriter, r *http.Request) *Context {
	c := contextPool.Get().(*Context)
	c.Writer = w
	c.Request = r
	c.Path = r.URL.Path
	c.Method = r.Method
	c.query = r.URL.Query()
	c.index = -1
	c.statusCode = 0

	// 清空 Params
	for k := range c.Params {
		delete(c.Params, k)
	}

	// 清空 data
	for k := range c.data {
		delete(c.data, k)
	}

	// 清空 handlers
	c.handlers = c.handlers[:0]

	return c
}

// releaseContext 释放 Context 回池中
func releaseContext(c *Context) {
	// 重置字段
	c.Writer = nil
	c.Request = nil
	c.Path = ""
	c.Method = ""
	c.query = nil
	c.handlers = nil
	c.paramKeys = nil
	c.paramVals = nil
	c.index = -1
	c.statusCode = 0

	// 清空 map
	for k := range c.Params {
		delete(c.Params, k)
	}
	for k := range c.data {
		delete(c.data, k)
	}

	contextPool.Put(c)
}

// Reset 重置 Context 状态
func (c *Context) Reset() {
	c.index = -1
	c.statusCode = 0

	// 清空 Params
	for k := range c.Params {
		delete(c.Params, k)
	}

	// 清空 data
	for k := range c.data {
		delete(c.data, k)
	}

	// 清空 handlers
	c.handlers = c.handlers[:0]
}

// bytePool 字节切片池
var bytePool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 1024)
	},
}

// acquireBytes 获取字节切片
func acquireBytes() []byte {
	return bytePool.Get().([]byte)
}

// releaseBytes 释放字节切片
func releaseBytes(b []byte) {
	if cap(b) <= 4096 { // 只回收小切片
		bytePool.Put(b[:0])
	}
}

// bufferPool 缓冲区池
var bufferPool = sync.Pool{
	New: func() interface{} {
		return &buffer{}
	},
}

type buffer struct {
	data []byte
}

func (b *buffer) Write(p []byte) (n int, err error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *buffer) Reset() {
	b.data = b.data[:0]
}

// Stats 返回池统计信息（用于调试）
type PoolStats struct {
	ContextPoolSize int
	ParamsPoolSize  int
	DataPoolSize    int
	HandlersSize    int
	BytePoolSize    int
}

// GetPoolStats 获取池统计（近似值）
// 注意：sync.Pool 没有提供直接获取大小的方法，这只是为了调试
func GetPoolStats() PoolStats {
	// 由于 sync.Pool 不暴露内部状态，这里返回 0
	// 实际大小由运行时管理
	return PoolStats{}
}

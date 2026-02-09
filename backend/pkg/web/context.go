package web

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

// validate 是全局验证器实例
var validate *validator.Validate
var validateOnce sync.Once

// Validator 返回全局验证器实例（懒加载）
func Validator() *validator.Validate {
	validateOnce.Do(func() {
		validate = validator.New()
	})
	return validate
}

type Context struct {
	Writer     http.ResponseWriter
	Request    *http.Request
	Path       string
	Method     string
	Params     map[string]string
	query      url.Values
	handlers   []HandlerFunc
	paramKeys  []string
	paramVals  []string
	index      int
	statusCode int
	data       map[string]interface{}
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	c := &Context{
		Writer:  w,
		Request: r,
		Path:    r.URL.Path,
		Method:  r.Method,
		query:   r.URL.Query(),
		index:   -1,
		data:    make(map[string]interface{}),
	}
	c.Params = make(map[string]string)
	return c
}

// Set 在上下文中存储数据
func (c *Context) Set(key string, value interface{}) {
	c.data[key] = value
}

// Get 从上下文中获取数据
func (c *Context) Get(key string) (interface{}, bool) {
	value, exists := c.data[key]
	return value, exists
}

func (c *Context) Next() {
	c.index++
	for c.index < len(c.handlers) {
		c.handlers[c.index](c)
		c.index++
	}
}

func (c *Context) Status(code int) {
	c.statusCode = code
	c.Writer.WriteHeader(code)
}

// GetStatus 获取当前状态码
func (c *Context) GetStatus() int {
	return c.statusCode
}

func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

func (c *Context) SetHeader(key, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) Query(key string) string {
	return c.query.Get(key)
}

func (c *Context) DefaultQuery(key, defaultValue string) string {
	if value := c.Query(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Context) AllQuery() map[string][]string {
	return c.query
}

func (c *Context) Param(key string) string {
	return c.Params[key]
}

func (c *Context) PostForm(key string) string {
	return c.Request.FormValue(key)
}

func (c *Context) DefaultPostForm(key, defaultValue string) string {
	if value := c.PostForm(key); value != "" {
		return value
	}
	return defaultValue
}

// Body 读取请求体，限制最大 32MB
func (c *Context) Body() []byte {
	if c.Request.Body == nil {
		return nil
	}
	const maxBodySize = 32 << 20 // 32MB
	body, _ := io.ReadAll(io.LimitReader(c.Request.Body, maxBodySize))
	c.Request.Body = io.NopCloser(strings.NewReader(string(body)))
	return body
}

// BodyWithLimit 读取请求体，可自定义大小限制
func (c *Context) BodyWithLimit(maxSize int64) ([]byte, error) {
	if c.Request.Body == nil {
		return nil, nil
	}
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxSize))
	if err != nil {
		return nil, err
	}
	c.Request.Body = io.NopCloser(strings.NewReader(string(body)))
	return body, nil
}

func (c *Context) Bind(obj interface{}) error {
	if err := json.Unmarshal(c.Body(), obj); err != nil {
		return err
	}
	return nil
}

// BindAndValidate 解析请求体并验证数据
// 使用标签: validate:"required", validate:"email", validate:"min=3,max=50" 等
func (c *Context) BindAndValidate(obj interface{}) error {
	// 先绑定数据
	if err := c.Bind(obj); err != nil {
		return err
	}
	// 再验证数据
	if err := Validator().Struct(obj); err != nil {
		return err
	}
	return nil
}

func (c *Context) JSON(code int, obj interface{}) error {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	return encoder.Encode(obj)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}

func (c *Context) JSONString(obj interface{}) (string, error) {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (c *Context) Redirect(code int, url string) {
	http.Redirect(c.Writer, c.Request, url, code)
}

func (c *Context) Abort() {
	c.index = len(c.handlers)
}

func (c *Context) IsAjax() bool {
	return c.GetHeader("X-Requested-With") == "XMLHttpRequest"
}

func (c *Context) Cookie(name string) (*http.Cookie, error) {
	return c.Request.Cookie(name)
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Writer, cookie)
}

func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}
	file, header, err := c.Request.FormFile(name)
	if err != nil {
		return nil, err
	}
	file.Close()
	return header, nil
}

func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

// ValidationError 验证错误信息
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors 验证错误列表
type ValidationErrors []ValidationError

// Error 实现 error 接口
func (v ValidationErrors) Error() string {
	var msgs []string
	for _, e := range v {
		msgs = append(msgs, fmt.Sprintf("%s: %s", e.Field, e.Message))
	}
	return strings.Join(msgs, "; ")
}

// FormatValidationErrors 将验证错误格式化为友好的错误信息
func FormatValidationErrors(err error) ValidationErrors {
	if err == nil {
		return nil
	}

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errors ValidationErrors
		for _, e := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   e.Field(),
				Message: formatValidationErrorMessage(e),
			})
		}
		return errors
	}

	return ValidationErrors{{Field: "", Message: err.Error()}}
}

// formatValidationErrorMessage 格式化单个验证错误信息
func formatValidationErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "该字段是必填项"
	case "email":
		return "请输入有效的邮箱地址"
	case "min":
		return "长度不能少于 " + e.Param()
	case "max":
		return "长度不能超过 " + e.Param()
	case "gte":
		return "值必须大于或等于 " + e.Param()
	case "lte":
		return "值必须小于或等于 " + e.Param()
	case "len":
		return "长度必须等于 " + e.Param()
	case "oneof":
		return "值必须是以下之一: " + e.Param()
	case "alphanum":
		return "只能包含字母和数字"
	case "numeric":
		return "必须是数字"
	default:
		return "验证失败: " + e.Tag()
	}
}

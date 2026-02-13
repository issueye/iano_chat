// Package builtin - 文件模块测试

package builtin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
)

// toJSPath 将 Windows 路径转换为 JavaScript 兼容路径
func toJSPath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

func TestFileModule_Name(t *testing.T) {
	module := NewFileModule(nil)
	assert.Equal(t, "file", module.Name())
}

func TestFileModule_Register(t *testing.T) {
	vm := goja.New()
	module := NewFileModule(nil)
	err := module.Register(vm)
	assert.NoError(t, err)

	// 检查 file 对象是否存在
	fileObj := vm.Get("file")
	assert.NotNil(t, fileObj)
	assert.False(t, goja.IsUndefined(fileObj))
}

func TestFileModule_ReadWrite(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "file_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 创建模块
	module := NewFileModule(&FileModuleConfig{
		AllowedDirs: []string{tempDir},
	})
	vm := goja.New()
	_ = module.Register(vm)

	// 测试写入
	testFile := filepath.Join(tempDir, "test.txt")
	jsPath := toJSPath(testFile)
	writeScript := `
		var result = file.write("` + jsPath + `", "Hello World");
		result;
	`
	value, err := vm.RunString(writeScript)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))

	// 测试读取
	readScript := `
		var result = file.read("` + jsPath + `");
		result;
	`
	value, err = vm.RunString(readScript)
	assert.NoError(t, err)
	result = value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))
	assert.Equal(t, "Hello World", result["data"].(string))
}

func TestFileModule_Exists(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "file_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	module := NewFileModule(&FileModuleConfig{
		AllowedDirs: []string{tempDir},
	})
	vm := goja.New()
	_ = module.Register(vm)

	// 测试不存在的文件
	testFile := filepath.Join(tempDir, "notexist.txt")
	script := `file.exists("` + toJSPath(testFile) + `")`
	value, err := vm.RunString(script)
	assert.NoError(t, err)
	assert.False(t, value.Export().(bool))

	// 创建文件
	os.WriteFile(testFile, []byte("test"), 0644)

	// 测试存在的文件
	value, err = vm.RunString(script)
	assert.NoError(t, err)
	assert.True(t, value.Export().(bool))
}

func TestFileModule_Delete(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "file_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	module := NewFileModule(&FileModuleConfig{
		AllowedDirs: []string{tempDir},
	})
	vm := goja.New()
	_ = module.Register(vm)

	// 创建测试文件
	testFile := filepath.Join(tempDir, "delete_test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	// 删除文件
	script := `file.delete("` + toJSPath(testFile) + `")`
	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))

	// 验证文件已删除
	_, err = os.Stat(testFile)
	assert.True(t, os.IsNotExist(err))
}

func TestFileModule_MkdirRmdir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "file_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	module := NewFileModule(&FileModuleConfig{
		AllowedDirs: []string{tempDir},
	})
	vm := goja.New()
	_ = module.Register(vm)

	// 创建目录
	testDir := filepath.Join(tempDir, "subdir", "nested")
	script := `file.mkdir("` + toJSPath(testDir) + `")`
	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))

	// 验证目录存在
	info, err := os.Stat(testDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())

	// 删除目录
	script = `file.rmdir("` + toJSPath(filepath.Join(tempDir, "subdir")) + `")`
	value, err = vm.RunString(script)
	assert.NoError(t, err)
	result = value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))
}

func TestFileModule_List(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "file_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 创建测试文件和目录
	os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("test1"), 0644)
	os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("test2"), 0644)
	os.Mkdir(filepath.Join(tempDir, "subdir"), 0755)

	module := NewFileModule(&FileModuleConfig{
		AllowedDirs: []string{tempDir},
	})
	vm := goja.New()
	_ = module.Register(vm)

	script := `file.list("` + toJSPath(tempDir) + `")`
	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))

	// files 可能是 []interface{} 或 []map[string]interface{}
	filesData := result["data"]
	switch v := filesData.(type) {
	case []interface{}:
		assert.Equal(t, 3, len(v))
	case []map[string]interface{}:
		assert.Equal(t, 3, len(v))
	default:
		t.Fatalf("unexpected files type: %T", filesData)
	}
}

func TestFileModule_Stat(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "file_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "stat_test.txt")
	os.WriteFile(testFile, []byte("Hello World"), 0644)

	module := NewFileModule(&FileModuleConfig{
		AllowedDirs: []string{tempDir},
	})
	vm := goja.New()
	_ = module.Register(vm)

	script := `file.stat("` + toJSPath(testFile) + `")`
	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))

	data := result["data"].(map[string]interface{})
	assert.Equal(t, "stat_test.txt", data["name"])
	assert.Equal(t, int64(11), data["size"])
	assert.False(t, data["isDir"].(bool))
}

func TestFileModule_JSON(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "file_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	module := NewFileModule(&FileModuleConfig{
		AllowedDirs: []string{tempDir},
	})
	vm := goja.New()
	_ = module.Register(vm)

	testFile := filepath.Join(tempDir, "test.json")

	// 写入 JSON
	writeScript := `
		var data = { name: "test", value: 42, items: [1, 2, 3] };
		file.writeJSON("` + toJSPath(testFile) + `", data);
	`
	value, err := vm.RunString(writeScript)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))

	// 读取 JSON
	readScript := `file.readJSON("` + toJSPath(testFile) + `")`
	value, err = vm.RunString(readScript)
	assert.NoError(t, err)
	result = value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))

	data := result["data"].(map[string]interface{})
	assert.Equal(t, "test", data["name"])
}

func TestFileModule_ReadOnly(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "file_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	module := NewFileModule(&FileModuleConfig{
		AllowedDirs: []string{tempDir},
		ReadOnly:    true,
	})
	vm := goja.New()
	_ = module.Register(vm)

	testFile := filepath.Join(tempDir, "readonly_test.txt")
	script := `file.write("` + toJSPath(testFile) + `", "test")`
	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["error"], "read-only")
}

func TestFileModule_PathNotAllowed(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "file_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 只允许访问 tempDir
	module := NewFileModule(&FileModuleConfig{
		AllowedDirs: []string{tempDir},
	})
	vm := goja.New()
	_ = module.Register(vm)

	// 尝试访问不允许的路径
	script := `file.read("/etc/passwd")`
	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["error"], "not allowed")
}

func TestFileModule_Append(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "file_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	module := NewFileModule(&FileModuleConfig{
		AllowedDirs: []string{tempDir},
	})
	vm := goja.New()
	_ = module.Register(vm)

	testFile := filepath.Join(tempDir, "append_test.txt")

	// 第一次写入
	script := `file.append("` + toJSPath(testFile) + `", "Hello")`
	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))

	// 追加内容
	script = `file.append("` + toJSPath(testFile) + `", " World")`
	value, err = vm.RunString(script)
	assert.NoError(t, err)
	result = value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))

	// 验证内容
	content, err := os.ReadFile(testFile)
	assert.NoError(t, err)
	assert.Equal(t, "Hello World", string(content))
}

func TestFileModule_Copy(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "file_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	module := NewFileModule(&FileModuleConfig{
		AllowedDirs: []string{tempDir},
	})
	vm := goja.New()
	_ = module.Register(vm)

	srcFile := filepath.Join(tempDir, "source.txt")
	dstFile := filepath.Join(tempDir, "copy.txt")

	// 创建源文件
	os.WriteFile(srcFile, []byte("Copy Test"), 0644)

	// 复制文件
	script := `file.copy("` + toJSPath(srcFile) + `", "` + toJSPath(dstFile) + `")`
	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))

	// 验证目标文件
	content, err := os.ReadFile(dstFile)
	assert.NoError(t, err)
	assert.Equal(t, "Copy Test", string(content))
}

func TestFileModule_Rename(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "file_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	module := NewFileModule(&FileModuleConfig{
		AllowedDirs: []string{tempDir},
	})
	vm := goja.New()
	_ = module.Register(vm)

	oldPath := filepath.Join(tempDir, "old_name.txt")
	newPath := filepath.Join(tempDir, "new_name.txt")

	// 创建源文件
	os.WriteFile(oldPath, []byte("Rename Test"), 0644)

	// 重命名
	script := `file.rename("` + toJSPath(oldPath) + `", "` + toJSPath(newPath) + `")`
	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))

	// 验证旧文件不存在，新文件存在
	_, err = os.Stat(oldPath)
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(newPath)
	assert.NoError(t, err)
}

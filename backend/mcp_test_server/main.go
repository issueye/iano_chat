package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	s := server.NewMCPServer("test-mcp-server", "1.0.0")

	toolEcho := mcp.NewTool(
		"echo",
		mcp.WithDescription("回显输入的文本"),
		mcp.WithString("text", mcp.Required(), mcp.Description("要回显的文本")),
	)
	s.AddTool(toolEcho, echoHandler)

	toolAdd := mcp.NewTool(
		"add",
		mcp.WithDescription("计算两个数的和"),
		mcp.WithNumber("a", mcp.Required(), mcp.Description("第一个数字")),
		mcp.WithNumber("b", mcp.Required(), mcp.Description("第二个数字")),
	)
	s.AddTool(toolAdd, addHandler)

	toolGetEnvInfo := mcp.NewTool(
		"get_environment_info",
		mcp.WithDescription("获取当前运行环境的信息"),
	)
	s.AddTool(toolGetEnvInfo, getEnvInfoHandler)

	toolReadFile := mcp.NewTool(
		"read_file",
		mcp.WithDescription("读取指定文件的内容"),
		mcp.WithString("path", mcp.Required(), mcp.Description("文件路径")),
	)
	s.AddTool(toolReadFile, readFileHandler)

	toolListDirectory := mcp.NewTool(
		"list_directory",
		mcp.WithDescription("列出指定目录的内容"),
		mcp.WithString("path", mcp.Description("目录路径，默认为当前目录")),
	)
	s.AddTool(toolListDirectory, listDirHandler)

	if err := server.ServeStdio(s); err != nil {
		log.Fatal(err)
	}
}

func echoHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, ok := request.Params.Arguments["text"].(string)
	if !ok {
		return nil, errors.New("text must be a string")
	}
	return mcp.NewToolResultText(name), nil
}

func addHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	a, ok := request.Params.Arguments["a"].(float64)
	if !ok {
		return nil, errors.New("a must be a number")
	}
	b, ok := request.Params.Arguments["b"].(float64)
	if !ok {
		return nil, errors.New("b must be a number")
	}
	result := a + b
	return mcp.NewToolResultText(fmt.Sprintf("结果: %.2f", result)), nil
}

func getEnvInfoHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	hostname, _ := os.Hostname()
	wd, _ := os.Getwd()
	info := fmt.Sprintf("主机名: %s\n工作目录: %s\n", hostname, wd)
	return mcp.NewToolResultText(info), nil
}

func readFileHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, ok := request.Params.Arguments["path"].(string)
	if !ok {
		return nil, errors.New("path must be a string")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	return mcp.NewToolResultText(string(content)), nil
}

func listDirHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := "."
	if p, ok := request.Params.Arguments["path"].(string); ok {
		path = p
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %v", err)
	}

	result := "目录内容:\n"
	for _, entry := range entries {
		fileType := "文件"
		if entry.IsDir() {
			fileType = "目录"
		}
		result += fmt.Sprintf("  [%s] %s\n", fileType, entry.Name())
	}

	return mcp.NewToolResultText(result), nil
}

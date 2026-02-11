package tools

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type ArchiveCreateTool struct {
	basePath string
}

func NewArchiveCreateTool(basePath string) *ArchiveCreateTool {
	if basePath == "" {
		basePath, _ = os.Getwd()
	}
	return &ArchiveCreateTool{basePath: basePath}
}

func (t *ArchiveCreateTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "archive_create",
		Desc: "创建 ZIP 压缩包",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"source": {
				Type:     schema.String,
				Desc:     "要压缩的文件或目录路径",
				Required: true,
			},
			"output": {
				Type:     schema.String,
				Desc:     "输出 ZIP 文件路径",
				Required: true,
			},
		}),
	}, nil
}

func (t *ArchiveCreateTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Source string `json:"source"`
		Output string `json:"output"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	absSource, err := t.resolvePath(args.Source)
	if err != nil {
		return "", err
	}

	absOutput := args.Output
	if !filepath.IsAbs(absOutput) {
		absOutput = filepath.Join(t.basePath, absOutput)
	}

	zipFile, err := os.Create(absOutput)
	if err != nil {
		return "", fmt.Errorf("创建 ZIP 文件失败: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	var fileCount int

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(absSource, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		if err != nil {
			return err
		}

		fileCount++
		return nil
	}

	if err := filepath.Walk(absSource, walkFn); err != nil {
		return "", fmt.Errorf("遍历源目录失败: %w", err)
	}

	return fmt.Sprintf("已创建压缩包: %s (%d 个文件)", args.Output, fileCount), nil
}

func (t *ArchiveCreateTool) resolvePath(path string) (string, error) {
	absPath := path
	if !filepath.IsAbs(path) {
		absPath = filepath.Join(t.basePath, path)
	}
	absPath = filepath.Clean(absPath)

	if t.basePath != "" {
		rel, err := filepath.Rel(t.basePath, absPath)
		if err != nil || strings.HasPrefix(rel, "..") {
			return "", fmt.Errorf("路径超出允许范围")
		}
	}

	return absPath, nil
}

type ArchiveExtractTool struct {
	basePath string
}

func NewArchiveExtractTool(basePath string) *ArchiveExtractTool {
	if basePath == "" {
		basePath, _ = os.Getwd()
	}
	return &ArchiveExtractTool{basePath: basePath}
}

func (t *ArchiveExtractTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "archive_extract",
		Desc: "解压 ZIP 压缩包",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"source": {
				Type:     schema.String,
				Desc:     "ZIP 文件路径",
				Required: true,
			},
			"output": {
				Type:     schema.String,
				Desc:     "解压目标目录（默认为当前目录）",
				Required: false,
			},
		}),
	}, nil
}

func (t *ArchiveExtractTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Source string `json:"source"`
		Output string `json:"output"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	absSource, err := t.resolvePath(args.Source)
	if err != nil {
		return "", err
	}

	outputDir := args.Output
	if outputDir == "" {
		outputDir = filepath.Dir(absSource)
	} else if !filepath.IsAbs(outputDir) {
		outputDir = filepath.Join(t.basePath, outputDir)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("创建目标目录失败: %w", err)
	}

	zipReader, err := zip.OpenReader(absSource)
	if err != nil {
		return "", fmt.Errorf("打开 ZIP 文件失败: %w", err)
	}
	defer zipReader.Close()

	var fileCount int

	for _, file := range zipReader.File {
		outputPath := filepath.Join(outputDir, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(outputPath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return "", fmt.Errorf("创建目录失败: %w", err)
		}

		outputFile, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return "", fmt.Errorf("创建文件失败: %w", err)
		}

		inputFile, err := file.Open()
		if err != nil {
			outputFile.Close()
			return "", fmt.Errorf("读取 ZIP 内容失败: %w", err)
		}

		_, err = io.Copy(outputFile, inputFile)
		inputFile.Close()
		outputFile.Close()

		if err != nil {
			return "", fmt.Errorf("写入文件失败: %w", err)
		}

		fileCount++
	}

	return fmt.Sprintf("已解压 %d 个文件到: %s", fileCount, outputDir), nil
}

func (t *ArchiveExtractTool) resolvePath(path string) (string, error) {
	absPath := path
	if !filepath.IsAbs(path) {
		absPath = filepath.Join(t.basePath, path)
	}
	absPath = filepath.Clean(absPath)

	if t.basePath != "" {
		rel, err := filepath.Rel(t.basePath, absPath)
		if err != nil || strings.HasPrefix(rel, "..") {
			return "", fmt.Errorf("路径超出允许范围")
		}
	}

	return absPath, nil
}

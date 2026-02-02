package disk_filesystem

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/adk/middlewares/filesystem"
	"github.com/yargevad/filepathx"
)

// InDiskBackend implements the filesystem.Backend interface for local disk operations.
type InDiskBackend struct{}

// NewInDiskBackend creates a new InDiskBackend instance.
func NewInDiskBackend() filesystem.Backend {
	return &InDiskBackend{}
}

// LsInfo lists file information under the given path.
func (b *InDiskBackend) LsInfo(ctx context.Context, req *filesystem.LsInfoRequest) ([]filesystem.FileInfo, error) {
	path := req.Path
	if path == "" {
		path = "/"
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", path, err)
	}

	var files []filesystem.FileInfo
	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())
		files = append(files, filesystem.FileInfo{
			Path: fullPath,
		})
	}

	return files, nil
}

// Read reads file content with support for line-based offset and limit.
func (b *InDiskBackend) Read(ctx context.Context, req *filesystem.ReadRequest) (string, error) {
	filePath := req.FilePath

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	lineNum := 0

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 200
	}

	for scanner.Scan() {
		if lineNum >= offset && len(lines) < limit {
			lines = append(lines, scanner.Text())
		}
		lineNum++
		if len(lines) >= limit {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return strings.Join(lines, "\n"), nil
}

// GrepRaw searches for content matching the specified pattern in files.
func (b *InDiskBackend) GrepRaw(ctx context.Context, req *filesystem.GrepRequest) ([]filesystem.GrepMatch, error) {
	pattern := req.Pattern
	searchPath := req.Path
	glob := req.Glob

	if searchPath == "" {
		searchPath = "/"
	}

	var matches []filesystem.GrepMatch

	var files []string
	if glob != "" {
		patternPath := filepath.Join(searchPath, glob)
		var err error
		files, err = filepathx.Glob(patternPath)
		if err != nil {
			return nil, fmt.Errorf("failed to glob pattern %s: %w", patternPath, err)
		}
	} else {
		err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk directory %s: %w", searchPath, err)
		}
	}

	for _, file := range files {
		fileMatches, err := b.grepFile(file, pattern)
		if err != nil {
			continue
		}
		matches = append(matches, fileMatches...)
	}

	return matches, nil
}

func (b *InDiskBackend) grepFile(filePath, pattern string) ([]filesystem.GrepMatch, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []filesystem.GrepMatch
	scanner := bufio.NewScanner(file)
	lineNum := 1

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, pattern) {
			matches = append(matches, filesystem.GrepMatch{
				Path:    filePath,
				Line:    lineNum,
				Content: line,
			})
		}
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return matches, nil
}

// GlobInfo returns file information matching the glob pattern.
func (b *InDiskBackend) GlobInfo(ctx context.Context, req *filesystem.GlobInfoRequest) ([]filesystem.FileInfo, error) {
	pattern := req.Pattern
	path := req.Path

	if path == "" {
		path = "/"
	}

	fullPattern := filepath.Join(path, pattern)
	matches, err := filepathx.Glob(fullPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob pattern %s: %w", fullPattern, err)
	}

	var files []filesystem.FileInfo
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}
		if !info.IsDir() {
			files = append(files, filesystem.FileInfo{
				Path: match,
			})
		}
	}

	return files, nil
}

// Write creates or updates file content.
func (b *InDiskBackend) Write(ctx context.Context, req *filesystem.WriteRequest) error {
	filePath := req.FilePath
	content := req.Content

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("file already exists: %s", filePath)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

// Edit replaces string occurrences in a file.
func (b *InDiskBackend) Edit(ctx context.Context, req *filesystem.EditRequest) error {
	filePath := req.FilePath
	oldString := req.OldString
	newString := req.NewString
	replaceAll := req.ReplaceAll

	if oldString == "" {
		return fmt.Errorf("oldString cannot be empty")
	}

	if oldString == newString {
		return fmt.Errorf("oldString and newString cannot be the same")
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	contentStr := string(content)

	if !replaceAll {
		count := strings.Count(contentStr, oldString)
		if count == 0 {
			return fmt.Errorf("oldString not found in file: %s", filePath)
		}
		if count > 1 {
			return fmt.Errorf("oldString appears multiple times (%d) in file: %s", count, filePath)
		}
	}

	newContent := strings.ReplaceAll(contentStr, oldString, newString)

	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

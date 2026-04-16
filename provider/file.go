package provider

import (
	"context"
	"fmt"
	"io"
	"os"
)

type FileProvider struct {
	path string
	cfg  *FileConfig
}

type FileConfig struct {
	FilePath string
}

func NewFileProvider(cfg *FileConfig) (*FileProvider, error) {
	if cfg.FilePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	return &FileProvider{
		path: cfg.FilePath,
		cfg:  cfg,
	}, nil
}

// Get the names of all tags in the repository
func (p *FileProvider) ListVersions(ctx context.Context) ([]string, error) {
	return []string{"file"}, nil
}

func (p *FileProvider) ListFiles(ctx context.Context, version string) ([]string, error) {
	entries, err := os.ReadDir(p.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read root directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

func (p *FileProvider) DownloadFile(ctx context.Context, version, file string) ([]byte, error) {
	f, err := os.Open(p.path + "/" + file)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// TODO: Handle
func (p *FileProvider) GetPath(version, file string) string {
	panic("not implemented")
}

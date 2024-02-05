package app

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
)

type image struct {
	data []byte

	contentType string
}

func loadImage(location string) (*image, error) {
	slog.Debug("checking if resource is a file")
	// Prepend "public/" to the path because that's where the static files are
	file, err := os.Open(filepath.Join(publicFilesPath, location))
	if err == nil {
		slog.Info("resource from file")
		defer file.Close()
		b, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}

		return &image{
			data:        b,
			contentType: http.DetectContentType(b),
		}, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	slog.Debug("resource is not a file")

	slog.Debug("checking if resource is an URL")
	resp, err := http.Get(location)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	slog.Info("resource loaded from URL")
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &image{
		data:        b,
		contentType: resp.Header.Get("Content-Type"),
	}, nil
}

package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

var (
	defaultAddr         = "localhost:3000"
	defaultPollInterval = 15 * time.Minute
)

var (
	ErrMissingOwner = errors.New("missing owner")
	ErrMissingRepo  = errors.New("missing repo")
)

type Provider interface {
	ListVersions(ctx context.Context) ([]string, error)
	ListFiles(ctx context.Context, version, path string) ([]string, error)
}

func validateConfig(cfg *Config) error {
	if cfg.Owner == "" {
		return ErrMissingOwner
	}

	if cfg.Repo == "" {
		return ErrMissingRepo
	}

	if cfg.PollInterval == 0 {
		cfg.PollInterval = defaultPollInterval
	}

	if cfg.PathPrefix != "" && !strings.HasSuffix(cfg.PathPrefix, "/") {
		log.Printf("Path %s does not end with a slash, appending one\n", cfg.PathPrefix)
		cfg.PathPrefix += "/"
	}

	return nil
}

func New(cfg *Config, provider Provider) (*Server, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	s := &Server{
		provider: provider,
		cfg:      cfg,
	}

	return s, nil
}

type Server struct {
	provider Provider

	cfg *Config

	docsRWLock sync.RWMutex
	// The dicumentation files and their versions that are available
	docs []*Documentation
}

type Documentation struct {
	// The version of this documentation
	Version string
	// The different files in this version
	Files []string
}

func (s *Server) Owner() string {
	return s.cfg.Owner
}

func (s *Server) Repo() string {
	return s.cfg.Repo
}

func (s *Server) Path() string {
	return s.cfg.PathPrefix
}

func (s *Server) FileSuffix() string {
	return s.cfg.FileSuffix
}

func (s *Server) Run(ctx context.Context) error {
	if err := s.Poll(); err != nil {
		return err
	}

	errChan := make(chan error, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				errChan <- nil
				return
			case <-time.After(s.cfg.PollInterval):
				if err := s.Poll(); err != nil {
					errChan <- err
					return
				}
			}
		}
	}()

	return <-errChan
}

// Poll polls the provider for new versions and files.
func (s *Server) Poll() error {
	versions, err := s.provider.ListVersions(context.Background())
	if err != nil {
		return err
	}

	docs := make([]*Documentation, 0, len(versions))
	for _, version := range versions {
		files, err := s.provider.ListFiles(context.Background(), version, s.cfg.PathPrefix)
		if err != nil {
			return err
		}

		for i, file := range files {
			files[i] = strings.TrimPrefix(file, s.cfg.PathPrefix)
			files[i] = strings.TrimSuffix(files[i], s.cfg.FileSuffix)
		}

		docs = append(docs, &Documentation{
			Version: version,
			Files:   files,
		})
	}

	s.docsRWLock.Lock()
	s.docs = append(s.docs, docs...)
	s.docsRWLock.Unlock()

	return nil
}

func (s *Server) GetVersions() []string {
	s.docsRWLock.RLock()
	defer s.docsRWLock.RUnlock()

	versions := make([]string, len(s.docs))
	for i, d := range s.docs {
		versions[i] = d.Version
	}

	return versions
}

func (s *Server) GetVersion(version string) *Documentation {
	s.docsRWLock.RLock()
	defer s.docsRWLock.RUnlock()

	for _, d := range s.docs {
		if d.Version == version {
			return d
		}
	}

	return nil
}

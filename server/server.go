package server

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v58/github"
)

var (
	defaultPollInterval = 15 * time.Minute
)

type Provider interface {
	ListVersions(ctx context.Context) ([]string, error)
	ListFiles(ctx context.Context, version, path string) ([]string, error)
}

func validateConfig(cfg *Config) error {
	if cfg.PollInterval == 0 {
		slog.Info("no poll interval set, using default", "default", defaultPollInterval)
		cfg.PollInterval = defaultPollInterval
	}

	if cfg.PathPrefix != "" && !strings.HasSuffix(cfg.PathPrefix, "/") {
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

func (s *Server) Path() string {
	return s.cfg.PathPrefix
}

func (s *Server) FileSuffix() string {
	return s.cfg.FileSuffix
}

func (s *Server) Run(ctx context.Context) error {
	slog.Info("starting server")

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
		if err, ok := err.(*github.RateLimitError); ok {
			slog.Warn("rate limit reached, skipping poll", "rate", err.Rate)
			return nil
		}
		if err, ok := err.(*github.AbuseRateLimitError); ok {
			slog.Warn("abuse rate limit reached, skipping poll", "retry-after", err.RetryAfter)
			return nil
		}
		return err
	}

	newVersions, removedVersions := s.calculateVersionDiffs(versions)

	for _, version := range newVersions {
		slog.Info("found new version", "version", version)
		if err := s.FetchVersion(version); err != nil {
			return err
		}
	}

	for _, version := range removedVersions {
		slog.Info("removed version", "version", version)
		s.RemoveVersion(version)
	}

	return nil
}

func (s *Server) FetchVersion(version string) error {
	files, err := s.provider.ListFiles(context.Background(), version, s.cfg.PathPrefix)
	if err != nil {
		return err
	}

	for i := range files {
		f, ok := strings.CutSuffix(files[i], s.cfg.FileSuffix)
		if !ok {
			slog.Warn("file does not end with the suffix, skipping", "file", files[i], "version", version, "suffix", s.cfg.FileSuffix)
			continue
		}
		f = strings.TrimPrefix(f, s.cfg.PathPrefix)
		files[i] = f
	}

	s.docsRWLock.Lock()
	defer s.docsRWLock.Unlock()

	// If the version already exists, update the files
	for _, d := range s.docs {
		if d.Version == version {
			d.Files = files
			return nil
		}
	}

	// Otherwise, append a new version
	s.docs = append(s.docs, &Documentation{
		Version: version,
		Files:   files,
	})

	return nil
}

func (s *Server) RemoveVersion(version string) {
	s.docsRWLock.Lock()
	defer s.docsRWLock.Unlock()

	for i, d := range s.docs {
		if d.Version == version {
			s.docs = append(s.docs[:i], s.docs[i+1:]...)
			break
		}
	}
}

// calculateVersionDiffs calculates the differences between the currently
// available versions and the ones that were found by the provider.
func (s *Server) calculateVersionDiffs(foundVersions []string) (newVersions []string, removedVersions []string) {
	foundVersionsMap := make(map[string]struct{}, len(foundVersions))
	for _, t := range foundVersions {
		foundVersionsMap[t] = struct{}{}
	}

	s.docsRWLock.RLock()
	for _, d := range s.docs {
		if _, ok := foundVersionsMap[d.Version]; !ok {
			removedVersions = append(removedVersions, d.Version)
		}
	}
	s.docsRWLock.RUnlock()

	for _, t := range foundVersions {
		if !slices.ContainsFunc(s.docs, func(d *Documentation) bool {
			return d.Version == t
		}) {
			newVersions = append(newVersions, t)
		}
	}

	return newVersions, removedVersions
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

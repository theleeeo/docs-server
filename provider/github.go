package provider

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/go-github/v58/github"
)

const (
	defaultMaxTags = 10
)

type GithubProvider struct {
	cfg    *GithubConfig
	client *github.Client
}

type GithubConfig struct {
	Owner     string
	Repo      string
	MaxTags   int
	AuthToken string
}

func NewGithub(cfg *GithubConfig) (*GithubProvider, error) {
	if cfg.Owner == "" {
		return nil, fmt.Errorf("owner cannot be empty")
	}

	if cfg.Repo == "" {
		return nil, fmt.Errorf("repo cannot be empty")
	}

	if cfg.MaxTags <= 0 {
		slog.Info("max tags not set, using default", "default", defaultMaxTags)
		cfg.MaxTags = defaultMaxTags
	}

	cl := github.NewClient(nil)
	if cfg.AuthToken != "" {
		cl.WithAuthToken(cfg.AuthToken)
	}

	return &GithubProvider{
		cfg:    cfg,
		client: cl,
	}, nil
}

// RootURL returns the url of where to get the swagger files from
func (p *GithubProvider) RootURL() string {
	return fmt.Sprintf("raw.githubusercontent.com/%s/%s", p.cfg.Owner, p.cfg.Repo)
}

// Get the names of all tags in the repository
func (p *GithubProvider) ListVersions(ctx context.Context) ([]string, error) {
	tags, _, err := p.client.Repositories.ListTags(ctx, p.cfg.Owner, p.cfg.Repo, &github.ListOptions{PerPage: p.cfg.MaxTags})
	if err != nil {
		return nil, err
	}

	var versions []string
	for _, tag := range tags {
		versions = append(versions, *tag.Name)
	}

	return versions, nil
}

func (p *GithubProvider) ListFiles(ctx context.Context, tag, path string) ([]string, error) {
	tree, _, err := p.client.Git.GetTree(ctx, p.cfg.Owner, p.cfg.Repo, tag, true)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range tree.Entries {
		if *entry.Type == "blob" && strings.HasPrefix(*entry.Path, path) {
			files = append(files, *entry.Path)
		}
	}

	return files, nil
}

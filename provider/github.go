package provider

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/google/go-github/v58/github"
)

const (
	defaultMaxTags = 10
)

type GithubProvider struct {
	client *github.Client

	cfg     *GithubConfig
	rootUrl string
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

	rootUrl, err := url.Parse(fmt.Sprintf("https://raw.githubusercontent.com/%s/%s", cfg.Owner, cfg.Repo))
	if err != nil {
		return nil, fmt.Errorf("invalid root url: %w", err)
	}

	cl := github.NewClient(nil)
	if cfg.AuthToken != "" {
		cl.WithAuthToken(cfg.AuthToken)
	}

	return &GithubProvider{
		cfg:     cfg,
		client:  cl,
		rootUrl: rootUrl.String(),
	}, nil
}

// RootURL returns the url of where to get the swagger files from
func (p *GithubProvider) RootURL() string {
	return p.rootUrl
}

// Get the names of all tags in the repository
func (p *GithubProvider) ListVersions(ctx context.Context) ([]string, error) {
	tags, _, err := p.client.Repositories.ListTags(ctx, p.cfg.Owner, p.cfg.Repo, &github.ListOptions{PerPage: p.cfg.MaxTags})
	if err != nil {
		return nil, handleError(err)
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
		return nil, handleError(err)
	}

	var files []string
	for _, entry := range tree.Entries {
		if *entry.Type == "blob" && strings.HasPrefix(*entry.Path, path) {
			files = append(files, *entry.Path)
		}
	}

	return files, nil
}

func handleError(err error) error {
	if err, ok := err.(*github.RateLimitError); ok {
		return NewRateLimitError("rate limit reached", "limit", err.Rate.Limit, "reset", err.Rate.Reset)
	}
	if err, ok := err.(*github.AbuseRateLimitError); ok {
		return NewRateLimitError("abuse rate limit reached", "retry_after", err.RetryAfter)
	}
	return err
}

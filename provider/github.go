package provider

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strings"

	"github.com/google/go-github/v58/github"
)

const (
	defaultMaxTags = 10
)

var (
	ErrNotFound = fmt.Errorf("not found")
)

type GithubProvider struct {
	client *github.Client

	cfg     *GithubConfig
	rootUrl string
}

type GithubConfig struct {
	Owner      string
	Repo       string
	PathPrefix string
	FileSuffix string
	MaxTags    int
	AuthToken  string
}

func NewGithub(cfg *GithubConfig) (*GithubProvider, error) {
	if cfg.Owner == "" {
		return nil, fmt.Errorf("owner cannot be empty")
	}

	if cfg.Repo == "" {
		return nil, fmt.Errorf("repo cannot be empty")
	}

	if cfg.PathPrefix != "" {
		cfg.PathPrefix = strings.Trim(cfg.PathPrefix, "/")
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

func (p *GithubProvider) ListFiles(ctx context.Context, version string) ([]string, error) {
	tree, _, err := p.client.Git.GetTree(ctx, p.cfg.Owner, p.cfg.Repo, version, true)
	if err != nil {
		return nil, handleError(err)
	}

	var files []string
	for _, entry := range tree.Entries {
		// Only include files that are in the path prefix and are blobs
		if *entry.Type == "blob" && strings.HasPrefix(*entry.Path, p.cfg.PathPrefix) {
			f := strings.TrimPrefix(*entry.Path, p.cfg.PathPrefix)
			f = strings.TrimPrefix(f, "/")

			// If the file does not end with the suffix, skip it
			f, ok := strings.CutSuffix(f, p.cfg.FileSuffix)
			if !ok {
				slog.Warn("file does not end with the suffix, skipping", "file", f, "version", version, "suffix", p.cfg.FileSuffix)
				continue
			}

			files = append(files, f)
		}
	}

	return files, nil
}

func (p *GithubProvider) GetPath(version, file string) string {
	return fmt.Sprint(p.rootUrl, "/", version, "/", p.cfg.PathPrefix, "/", file, p.cfg.FileSuffix)
}

func (p *GithubProvider) DownloadFile(ctx context.Context, version, file string) ([]byte, error) {
	path := fmt.Sprint(p.cfg.PathPrefix, "/", file, p.cfg.FileSuffix)
	content, resp, err := p.client.Repositories.DownloadContents(ctx, p.cfg.Owner, p.cfg.Repo, path, &github.RepositoryContentGetOptions{Ref: version})
	if err != nil {
		if strings.Contains(err.Error(), "No commit found") {
			return nil, fmt.Errorf("%w: version=%s", ErrNotFound, version)
		}
		if strings.Contains(err.Error(), "no file named") {
			return nil, fmt.Errorf("%w: file=%s", ErrNotFound, file)
		}
		return nil, handleError(err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	defer content.Close()
	return io.ReadAll(content)
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

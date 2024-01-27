package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v58/github"
)

type githubProvider struct {
	owner  string
	repo   string
	client *github.Client
}

func NewGithub(owner, repo string) (*githubProvider, error) {
	if owner == "" {
		return nil, fmt.Errorf("owner cannot be empty")
	}

	if repo == "" {
		return nil, fmt.Errorf("repo cannot be empty")
	}

	return &githubProvider{
		owner:  owner,
		repo:   repo,
		client: github.NewClient(nil),
	}, nil
}

// Get the names of all tags in the repository
func (p *githubProvider) ListVersions(ctx context.Context) ([]string, error) {
	tags, _, err := p.client.Repositories.ListTags(ctx, p.owner, p.repo, &github.ListOptions{PerPage: 10})
	if err != nil {
		return nil, err
	}

	var versions []string
	for _, tag := range tags {
		versions = append(versions, *tag.Name)
	}

	return versions, nil
}

func (p *githubProvider) ListFiles(ctx context.Context, tag, path string) ([]string, error) {
	tree, _, err := p.client.Git.GetTree(ctx, p.owner, p.repo, tag, true)
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

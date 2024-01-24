package provider

import (
	"context"
	"log"
	"strings"

	"github.com/google/go-github/v58/github"
)

type githubProvider struct {
	owner  string
	repo   string
	client *github.Client
}

func NewGithub(owner, repo string) *githubProvider {
	return &githubProvider{
		owner:  owner,
		repo:   repo,
		client: github.NewClient(nil),
	}
}

// Get the names of all tags in the repository
func (c *githubProvider) ListVersions(ctx context.Context) ([]string, error) {
	tags, _, err := c.client.Repositories.ListTags(ctx, c.owner, c.repo, &github.ListOptions{PerPage: 10})
	if err != nil {
		return nil, err
	}

	var versions []string
	for _, tag := range tags {
		versions = append(versions, *tag.Name)
	}

	return versions, nil
}

func (c *githubProvider) ListFiles(ctx context.Context, tag, path string) ([]string, error) {
	tree, _, err := c.client.Git.GetTree(ctx, c.owner, c.repo, tag, true)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range tree.Entries {
		if *entry.Type == "blob" && strings.HasPrefix(*entry.Path, path) {
			files = append(files, *entry.Path)
		} else {
			log.Printf("Skipping %s", *entry.Path)
		}
	}

	return files, nil
}

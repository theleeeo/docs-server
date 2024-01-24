package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/TheLeeeo/docs-server/provider"
	"github.com/TheLeeeo/docs-server/server"
)

func loadConfig() (*server.Config, error) {
	pollInterval, err := parsePollInterval(os.Getenv("POLL_INTERVAL"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse poll interval: %w", err)
	}

	return &server.Config{
		Owner:        os.Getenv("OWNER"),
		Repo:         os.Getenv("REPO"),
		PathPrefix:   os.Getenv("PATH_PREFIX"),
		FileSuffix:   os.Getenv("FILE_SUFFIX"),
		PollInterval: pollInterval,
	}, nil
}

func parsePollInterval(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}

	return d, nil
}

func main() {
	serverConfig, err := loadConfig()
	if err != nil {
		panic(err)
	}

	ghClient := provider.NewGithub(serverConfig.Owner, serverConfig.Repo)

	s, err := server.New(serverConfig, ghClient)
	if err != nil {
		panic(err)
	}

	app := NewApp(s)

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "localhost:3000"
	}

	go func() {
		if err := app.Listen(addr); err != nil {
			panic(err)
		}
	}()

	if err := s.Run(context.TODO()); err != nil {
		panic(err)
	}

	app.Shutdown()

}

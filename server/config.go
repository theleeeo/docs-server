package server

import "time"

type Config struct {
	Owner string
	Repo  string

	PathPrefix string
	FileSuffix string

	PollInterval time.Duration
}

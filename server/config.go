package server

import "time"

type Config struct {
	Owner string
	Repo  string

	CompanyName string

	PathPrefix string
	FileSuffix string

	PollInterval time.Duration
}

package server

import "time"

type Config struct {
	PollInterval time.Duration

	PathPrefix string
	FileSuffix string
}

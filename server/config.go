package server

import "time"

type Config struct {
	PollInterval time.Duration
	Proxy        bool
}

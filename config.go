package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	LogLevel string `yaml:"log_level"`

	Provider struct {
		Github *struct {
			Owner      string `yaml:"owner"`
			Repo       string `yaml:"repo"`
			PathPrefix string `yaml:"path_prefix"`
			FileSuffix string `yaml:"file_suffix"`
			MaxTags    int    `yaml:"max_tags"`
			AuthToken  string `yaml:"auth_token"`
		} `yaml:"github"`
	} `yaml:"provider"`

	Server struct {
		PollInterval string `yaml:"poll_interval"`
		Proxy        bool   `yaml:"proxy"`
	} `yaml:"server"`

	App struct {
		Address    string `yaml:"address"`
		PathPrefix string `yaml:"path_prefix"`
	} `yaml:"app"`

	Design struct {
		HeaderTitle string `yaml:"header_name"`
		HeaderImage string `yaml:"header_image"`
		HeaderColor string `yaml:"header_color"`
		Favicon     string `yaml:"favicon"`
	} `yaml:"design"`
}

func loadConfig() (*Config, error) {
	content, err := os.ReadFile("./config.yml")
	if err != nil {
		log.Fatal(err)
	}

	var config Config
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return &config, nil
}

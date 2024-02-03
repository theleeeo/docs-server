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
			Owner     string `yaml:"owner"`
			Repo      string `yaml:"repo"`
			MaxTags   int    `yaml:"max_tags"`
			AuthToken string `yaml:"auth_token"`
		} `yaml:"github"`
	} `yaml:"provider"`

	Server struct {
		PollInterval string `yaml:"poll_interval"`
		PathPrefix   string `yaml:"path_prefix"`
		FileSuffix   string `yaml:"file_suffix"`
	} `yaml:"server"`

	App struct {
		Address     string `yaml:"address"`
		DocsUseHttp bool   `yaml:"docs_use_http"`
	} `yaml:"app"`

	Design struct {
		HeaderTitle string `yaml:"header_name"`
		HeaderLogo  string `yaml:"header_logo"`
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

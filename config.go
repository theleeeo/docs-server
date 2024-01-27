package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Provider struct {
		Github struct {
			Owner string `yaml:"owner"`
			Repo  string `yaml:"repo"`
		} `yaml:"github"`
	} `yaml:"provider"`

	App struct {
		Address string `yaml:"address"`
		RootUrl string `yaml:"root_url"`
	} `yaml:"app"`

	Server struct {
		PollInterval string `yaml:"poll_interval"`
		PathPrefix   string `yaml:"path_prefix"`
		FileSuffix   string `yaml:"file_suffix"`
	} `yaml:"server"`

	Design struct {
		CompanyName string `yaml:"company_name"`
		CompanyLogo string `yaml:"company_logo"`
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

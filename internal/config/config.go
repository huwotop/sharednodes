package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Subscription Subscription `yaml:"subscription"`
	Check        Check        `yaml:"check"`
	Filter       Filter       `yaml:"filter"`
	Output       Output       `yaml:"output"`
}

type Subscription struct {
	URLs []string `yaml:"urls"`
}

type Check struct {
	AliveTestURL      string `yaml:"alive_test_url"`
	AliveTestStatusCode int  `yaml:"alive_test_status_code"`
	AliveTimeout      int    `yaml:"alive_timeout"`
	MaxThreads        int    `yaml:"max_threads"`
	NodePoolSize      int    `yaml:"node_pool_size"`
}

type Filter struct {
	MinAlive         bool     `yaml:"min_alive"`
	MaxDelay         uint16   `yaml:"max_delay"`
	Countries        []string `yaml:"countries"`
	ExcludeCountries bool     `yaml:"exclude_countries"`
}

type Output struct {
	FileName string `yaml:"file_name"`
}

var cfg *Config

func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	cfg = &Config{}
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func Get() *Config {
	return cfg
}

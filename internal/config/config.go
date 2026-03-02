package config

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Target struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type Config struct {
	Port          int           `yaml:"port"`
	CheckInterval time.Duration `yaml:"check_interval"`
	CheckTimeout  time.Duration `yaml:"check_timeout"`
	RedisAddr     string        `yaml:"redis_addr"`
	RedisPassword string        `yaml:"redis_password"`
	Targets       []Target      `yaml:"targets"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	expanded := os.ExpandEnv(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", c.Port)
	}

	if c.CheckInterval <= 0 {
		return fmt.Errorf("check_interval must be positive, got %s", c.CheckInterval)
	}

	if c.CheckTimeout <= 0 {
		return fmt.Errorf("check_timeout must be positive, got %s", c.CheckTimeout)
	}

	if c.RedisAddr == "" {
		return fmt.Errorf("redis_addr must not be empty")
	}

	if len(c.Targets) == 0 {
		return fmt.Errorf("at least one target is required")
	}

	for i, t := range c.Targets {
		if t.Name == "" {
			return fmt.Errorf("target[%d]: name must not be empty", i)
		}
		if t.URL == "" {
			return fmt.Errorf("target[%d] %q: url must not be empty", i, t.Name)
		}
		u, err := url.Parse(t.URL)
		if err != nil {
			return fmt.Errorf("target[%d] %q: invalid url: %w", i, t.Name, err)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return fmt.Errorf("target[%d] %q: url scheme must be http or https, got %q", i, t.Name, u.Scheme)
		}
		if u.Host == "" {
			return fmt.Errorf("target[%d] %q: url must have a host", i, t.Name)
		}
	}

	return nil
}

package main

import (
	"fmt"
	"net/url"
	"os"
	"time"

	kenko "github.com/aidantrabs/kenko"
	"github.com/aidantrabs/kenko/prommetrics"
	"github.com/aidantrabs/kenko/redisstore"
	"gopkg.in/yaml.v3"
)

type target struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type config struct {
	Port          int           `yaml:"port"`
	CheckInterval time.Duration `yaml:"check_interval"`
	CheckTimeout  time.Duration `yaml:"check_timeout"`
	RedisAddr     string        `yaml:"redis_addr"`
	RedisPassword string        `yaml:"redis_password"`
	Targets       []target      `yaml:"targets"`
}

func loadConfig(path string) (*config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	expanded := os.ExpandEnv(string(data))

	var cfg config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *config) validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", c.Port)
	}

	if c.CheckInterval <= 0 {
		return fmt.Errorf("check_interval must be positive, got %s", c.CheckInterval)
	}

	if c.CheckTimeout <= 0 {
		return fmt.Errorf("check_timeout must be positive, got %s", c.CheckTimeout)
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

func configToOptions(cfg *config) []kenko.Option {
	opts := make([]kenko.Option, 0, len(cfg.Targets)+4)

	for _, t := range cfg.Targets {
		opts = append(opts, kenko.WithTarget(t.Name, t.URL))
	}

	opts = append(opts,
		kenko.WithInterval(cfg.CheckInterval),
		kenko.WithTimeout(cfg.CheckTimeout),
	)

	if cfg.RedisAddr != "" {
		var rsOpts []redisstore.Option
		if cfg.RedisPassword != "" {
			rsOpts = append(rsOpts, redisstore.WithPassword(cfg.RedisPassword))
		}
		opts = append(opts, kenko.WithStore(redisstore.New(cfg.RedisAddr, rsOpts...)))
	}

	opts = append(opts, kenko.WithMetrics(prommetrics.New()))

	return opts
}

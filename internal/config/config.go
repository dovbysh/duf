package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	defaultMaxOpenConns    = 8
	defaultMaxIdleConns    = 8
	defaultConnMaxLifetime = 30 * time.Minute
	defaultConnMaxIdleTime = 5 * time.Minute
)

type Config struct {
	Storage struct {
		ScanPaths       []string `yaml:"scan_paths"`
		ExcludePatterns []string `yaml:"exclude_patterns"`
	} `yaml:"storage"`
	Database struct {
		DSN             string `yaml:"dsn"`
		BatchSize       int    `yaml:"batch_size"`
		MaxOpenConns    int    `yaml:"max_open_conns"`
		MaxIdleConns    int    `yaml:"max_idle_conns"`
		ConnMaxLifetime string `yaml:"conn_max_lifetime"`
		ConnMaxIdleTime string `yaml:"conn_max_idle_time"`
	} `yaml:"database"`
	LMStudio struct {
		AuthToken     string `yaml:"auth_token"`
		APIURL        string `yaml:"api_url"`
		DeleteURL     string `yaml:"delete_url"`
		ModelName     string `yaml:"model_name"`
		StatefulChats bool   `yaml:"stateful_chats"`
	} `yaml:"lmstudio"`
	Performance struct {
		HashWorkers int `yaml:"hash_workers"`
	} `yaml:"performance"`
}

type DatabasePoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) DatabasePoolConfig() (DatabasePoolConfig, error) {
	pool := DatabasePoolConfig{
		MaxOpenConns:    defaultMaxOpenConns,
		MaxIdleConns:    defaultMaxIdleConns,
		ConnMaxLifetime: defaultConnMaxLifetime,
		ConnMaxIdleTime: defaultConnMaxIdleTime,
	}

	if c.Database.MaxOpenConns > 0 {
		pool.MaxOpenConns = c.Database.MaxOpenConns
	}
	if c.Database.MaxIdleConns > 0 {
		pool.MaxIdleConns = c.Database.MaxIdleConns
	}
	if c.Database.ConnMaxLifetime != "" {
		duration, err := time.ParseDuration(c.Database.ConnMaxLifetime)
		if err != nil {
			return pool, fmt.Errorf("parse database.conn_max_lifetime: %w", err)
		}
		pool.ConnMaxLifetime = duration
	}
	if c.Database.ConnMaxIdleTime != "" {
		duration, err := time.ParseDuration(c.Database.ConnMaxIdleTime)
		if err != nil {
			return pool, fmt.Errorf("parse database.conn_max_idle_time: %w", err)
		}
		pool.ConnMaxIdleTime = duration
	}

	return pool, nil
}

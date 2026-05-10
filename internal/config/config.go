package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Storage struct {
		ScanPaths       []string `yaml:"scan_paths"`
		ExcludePatterns []string `yaml:"exclude_patterns"`
	} `yaml:"storage"`
	Database struct {
		DSN       string `yaml:"dsn"`
		TableName string `yaml:"table_name"`
		BatchSize int    `yaml:"batch_size"`
	} `yaml:"database"`
	Performance struct {
		HashWorkers int `yaml:"hash_workers"`
	} `yaml:"performance"`
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

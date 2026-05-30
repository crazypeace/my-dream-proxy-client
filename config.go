package main

import (
	"flag"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Bind      string `yaml:"listen"`
	Port      string `yaml:"port"`
	FilesDir   string `yaml:"files-dir"`
	CoreStart string `yaml:"core-start"`
	CoreTest  string `yaml:"core-test"`
	Log       string `yaml:"log"`
}

func LoadConfig() *Config {
	defaultConf := "./mdpc-config.yaml"
	configPath := flag.String("config", envOrDefault("CONFIG", defaultConf), "Config file path")
	flag.Parse()

	cfg := &Config{
		Bind:    "127.0.0.1",
		Port:    "18080",
		FilesDir: "./bin/core/",
	}

	if data, err := os.ReadFile(*configPath); err == nil {
		_ = yaml.Unmarshal(data, cfg)
	}

	if v, err := filepath.Abs(cfg.FilesDir); err == nil {
		cfg.FilesDir = v
	}

	return cfg
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

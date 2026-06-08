package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// CoreConfig holds per-core settings (files-dir, core-start, core-test).
type CoreConfig struct {
	FilesDir  string `yaml:"files-dir"`
	CoreStart string `yaml:"core-start"`
	CoreTest  string `yaml:"core-test"`
}

// Config is the top-level mdpc-config.yaml structure.
// listen, port, log are global; every other top-level key is a core definition.
type Config struct {
	Bind  string
	Port  string
	Log   string
	Cores map[string]*CoreConfig
}

// UnmarshalYAML implements custom unmarshaling: known keys map to global fields,
// everything else is treated as a CoreConfig entry.
func (c *Config) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("config must be a mapping")
	}

	c.Cores = make(map[string]*CoreConfig)

	for i := 0; i < len(value.Content)-1; i += 2 {
		key := value.Content[i].Value
		val := value.Content[i+1]

		switch key {
		case "listen":
			if err := val.Decode(&c.Bind); err != nil {
				return err
			}
		case "port":
			if err := val.Decode(&c.Port); err != nil {
				return err
			}
		case "log":
			if err := val.Decode(&c.Log); err != nil {
				return err
			}
		default:
			var cc CoreConfig
			if err := val.Decode(&cc); err != nil {
				return fmt.Errorf("core %q: %w", key, err)
			}
			c.Cores[key] = &cc
		}
	}
	return nil
}

func LoadConfig() *Config {
	defaultConf := "./mdpc-config.yaml"
	configPath := flag.String("config", envOrDefault("CONFIG", defaultConf), "Config file path")
	flag.Parse()

	cfg := &Config{}

	if data, err := os.ReadFile(*configPath); err == nil {
		_ = yaml.Unmarshal(data, cfg)
	}

	// Apply defaults
	if cfg.Bind == "" {
		cfg.Bind = "127.0.0.1"
	}
	if cfg.Port == "" {
		cfg.Port = "18080"
	}
	if cfg.Cores == nil {
		cfg.Cores = make(map[string]*CoreConfig)
	}

	// Resolve absolute paths for each core
	for _, core := range cfg.Cores {
		if core.FilesDir != "" {
			if v, err := filepath.Abs(core.FilesDir); err == nil {
				core.FilesDir = v
			}
		}
	}

	return cfg
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

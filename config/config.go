package config

import (
	"apt_cacher_go/utils/asserted_path"
	"apt_cacher_go/utils/duration"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

//TODO: add file watcher to reload config on changes

var configPath = asserted_path.Assert("var/config.json")

var Global *Config = func() *Config {
	cfg, err := LoadOrDefault(configPath.GetPath())
	if err != nil {
		log.Panicf("Failed to load global config: %v", err)
	}
	return cfg
}()

type Config struct {
	AlwaysCache             bool              // If true, the proxy will always cache responses, even if the upstream response requests the opposite.
	UpstreamDefaultHttps    bool              // If true, the proxy will always send HTTPS instead of HTTP to the upstream server.
	DefaultCacheMaxAge      duration.Duration // The default cache max age to use if the upstream response does not specify a Cache-Control or Expires header.
	ForceDefaultCacheMaxAge bool              // If true, always use the default cache max age even if the upstream response has a Cache-Control or Expires header.
}

func Default() *Config {
	return &Config{
		AlwaysCache:             true, // This this is primarily targeted at caching apt repositories, we want to cache aggressively by default.
		UpstreamDefaultHttps:    true,
		DefaultCacheMaxAge:      duration.Duration(1 * time.Hour),
		ForceDefaultCacheMaxAge: true, // Since this is again primarily targeted at caching apt repositories, we want to cache aggressively by default.
	}
}

// Writes the configuration to disk.
func (c *Config) Persist() error {
	f, err := os.Create(configPath.GetPath())
	if err != nil {
		return fmt.Errorf("failed to open config file for writing: %v", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ") // Pretty print the JSON output
	if err := enc.Encode(c); err != nil {
		return fmt.Errorf("failed to write config to file: %v", err)
	}
	return nil
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	decoder.DisallowUnknownFields()

	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func LoadOrDefault(path string) (*Config, error) {
	cfg, err := Load(path)
	if err != nil {
		log.Printf("Failed to load config from '%s': '%v' Using default configuration...", path, err)
		cfg = Default()
		if err := cfg.Persist(); err != nil {
			return nil, fmt.Errorf("failed to persist default config: %v", err)
		}
	}
	return cfg, nil
}

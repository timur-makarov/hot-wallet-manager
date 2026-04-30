package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	HTTP  HTTPConfig   `json:"config"`
	Gates []GateConfig `json:"gates"`
}

type HTTPConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type GateConfig struct {
	Name     string `json:"name"`
	Mnemonic string `json:"mnemonic"`
}

func InitConfiguration(file string) (*Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.HTTP.Host == "" {
		return nil, fmt.Errorf("config.host is required")
	}
	if cfg.HTTP.Port <= 0 || cfg.HTTP.Port > 65535 {
		return nil, fmt.Errorf("config.port must be between 1 and 65535")
	}
	if len(cfg.Gates) == 0 {
		return nil, fmt.Errorf("at least one gate is required")
	}

	seen := make(map[string]struct{}, len(cfg.Gates))
	for i, gate := range cfg.Gates {
		if gate.Name == "" {
			return nil, fmt.Errorf("gates[%d].name is required", i)
		}
		if gate.Mnemonic == "" {
			return nil, fmt.Errorf("gates[%d].mnemonic is required", i)
		}
		if _, exists := seen[gate.Name]; exists {
			return nil, fmt.Errorf("duplicate gate name %q", gate.Name)
		}
		seen[gate.Name] = struct{}{}
	}

	return cfg, nil
}

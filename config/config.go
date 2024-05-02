package config

import (
	_ "embed"
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

//go:embed example.yaml
var example []byte

var ErrDatabaseURLRequired = errors.New("no database URL found, specify in configuration or set DATABASE_URL environment variable")

type Config struct {
	Database      Database      `json:"database"`
	Documentation Documentation `json:"documentation"`
}

type Database struct {
	URL           string   `json:"url"`
	Schemas       []string `json:"schemas"`
	ExcludeTables []string `json:"exclude_tables" yaml:"exclude_tables"`
}

type Documentation struct {
	Strategy  string `json:"strategy"`
	Directory string `json:"directory"`
	Filename  string `json:"filename"`
	Stdout    bool   `json:"stdout"`
	Mermaid   bool   `json:"mermaid"`
}

func Default() *Config {
	return &Config{
		Database: Database{
			URL:           os.Getenv("DATABASE_URL"),
			Schemas:       []string{"public"},
			ExcludeTables: []string{},
		},
		Documentation: Documentation{
			Strategy:  "unified",
			Directory: ".",
			Filename:  "schema.md",
			Stdout:    true,
		},
	}
}

func Parse(data []byte) (*Config, error) {
	config := Default()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}
	return config, nil
}

func ParseFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

func Load(path string) (*Config, error) {
	var conf *Config
	if path == "" {
		conf = Default()
	} else {
		var err error
		conf, err = ParseFile(path)
		if err != nil {
			return nil, err
		}
	}
	if dburl := os.Getenv("DATABASE_URL"); dburl != "" {
		conf.Database.URL = dburl
	}
	if conf.Database.URL == "" {
		return nil, ErrDatabaseURLRequired
	}
	return conf, nil
}

func WriteExample(path string) error {
	return os.WriteFile(path, example, 0644)
}

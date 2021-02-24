package config

import (
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

func Parse(configPath string) (*Config, error) {
	log.Debug().Str("path", configPath).Msg("reading configuration file")
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Load Provider data from string input
func LoadProviderFromString(data string) (Provider, error) {
	var p Provider
	err := yaml.Unmarshal([]byte(data), &p)
	if err != nil {
		return Provider{}, err
	}
	return p, nil
}
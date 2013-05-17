package config

import (
  "io/ioutil"
  "github.com/le0pard/go-falcon/log"
  "launchpad.net/goyaml"
)

// Config represents the supported configuration options for a charm,
// as declared in its config.yaml file.
type Config struct {
  Smtp struct {
    Host      string
    Port      string
  }
}

// NewConfig returns a new Config without any options.
func NewConfig() *Config {
  return &Config{}
}

// ReadEnvirons reads the juju config.yaml file
// and returns the result of running ParseEnvironments
// on the file's contents.
func ReadConfig(filename string) (*Config, error) {
  data, err := ioutil.ReadFile(filename)
  if err != nil {
    return nil, log.Errorf("cannot read file %q: %v", filename, err)
  }
  e, err := ReadConfigBytes(data)
  if err != nil {
    return nil, log.Errorf("cannot parse file %q: %v", filename, err)
  }
  return e, nil
}

// ReadConfigBytes parses the contents of an config.yaml file
// and returns its representation. An environment with an unknown type
// will only generate an error when New is called for that environment.
// Attributes for environments with known types are checked.
func ReadConfigBytes(data []byte) (*Config, error) {
  config := NewConfig()
  err := goyaml.Unmarshal(data, &config)
  if err != nil {
    return nil, err
  }
  return config, nil
}



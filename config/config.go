package config

import (
  "io/ioutil"
  "github.com/le0pard/go-falcon/log"
  "launchpad.net/goyaml"
)

// Config represents the supported configuration options for a falcon,
// as declared in its config.yml file.
type Config struct {
  Smtp struct {
    Enabled   bool
    Host      string
    Port      int
  }
  Lmtp struct {
    Enabled   bool
    Host      string
    Port      int
  }
  Storage struct {
    Host      string
    Port      int
    Username  string
    Password  string
  }
}

// NewConfig returns a new Config without any options.
func NewConfig() *Config {
  return &Config{}
}

// ReadEnvirons reads the juju config.yml file
// and returns the result of running Config
// on the file's contents.
func ReadConfig(filename string) (*Config, error) {
  data, err := ioutil.ReadFile(filename)
  if err != nil {
    log.Errorf("cannot read file %q: %v", filename, err)
    return nil, err
  }
  e, err := ReadConfigBytes(data)
  if err != nil {
    log.Errorf("cannot parse file %q: %v", filename, err)
    return nil, err
  }
  return e, nil
}

// ReadConfigBytes parses the contents of an config.yml file
// and returns its representation.
func ReadConfigBytes(data []byte) (*Config, error) {
  config := NewConfig()
  err := goyaml.Unmarshal(data, &config)
  if err != nil {
    return nil, err
  }
  return config, nil
}
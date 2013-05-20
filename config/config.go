package config

import (
  "io/ioutil"
  "github.com/le0pard/go-falcon/log"
  "launchpad.net/goyaml"
)

// Config represents the supported configuration options for a falcon,
// as declared in its config.yml file.
type protocolType string

const (
    protocolSmtp protocolType = "smtp"
    protocolLmtp protocolType = "lmtp"
)

type Config struct {
  Adapter struct {
    Protocol        protocolType
    Host            string
    Port            int
    Welcome_Msg     string
    Allow_Hosts     string
  }
  Storage struct {
    Host          string
    Port          int
    Username      string
    Password      string
    Database      string
    Pool          int
  }
  Proxy struct {
    Enabled       bool
    Host          string
    Port          int
  }
  Log struct {
    Debug         bool
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
  e, err := readConfigBytes(data)
  if err != nil {
    log.Errorf("cannot parse file %q: %v", filename, err)
    return nil, err
  }
  e.setDefaultValues()
  return e, nil
}

// setDefaultValues for yaml config
func (config *Config) setDefaultValues() {
  // default for Adapter
  if config.Adapter.Protocol != protocolSmtp && config.Adapter.Protocol != protocolLmtp {
    config.Adapter.Protocol = protocolSmtp
  }
  if config.Adapter.Host == "" {
    config.Adapter.Host = "localhost"
  }
  if config.Adapter.Port <= 0 {
    config.Adapter.Port = 25
  }
  // default for Storage
  if config.Storage.Host == "" {
    config.Storage.Host = "localhost"
  }
  if config.Storage.Port <= 0 {
    config.Storage.Port = 5432
  }
  if config.Storage.Pool <= 0 {
    config.Storage.Pool = 5
  }
  // default for Proxy
  if config.Proxy.Host == "" {
    config.Proxy.Host = "localhost"
  }
  if config.Proxy.Port <= 0 {
    config.Proxy.Port = 2525
  }
}

// readConfigBytes parses the contents of an config.yml file
// and returns its representation.
func readConfigBytes(data []byte) (*Config, error) {
  config := NewConfig()
  err := goyaml.Unmarshal(data, &config)
  if err != nil {
    return nil, err
  }
  return config, nil
}
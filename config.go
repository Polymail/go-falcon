package main


// Option represents a single configuration option that is declared
// as supported by a charm in its config.yaml file.
type Option struct {
  SmtpHost        string
  SmtpPort        string
}

// Config represents the supported configuration options for a charm,
// as declared in its config.yaml file.
type Config struct {
  Options map[string]Option
}

// NewConfig returns a new Config without any options.
func NewConfig() *Config {
  return &Config{make(map[string]Option)}
}


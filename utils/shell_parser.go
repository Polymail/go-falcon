package utils

import (
  "flag"
  "github.com/le0pard/go-falcon/logger"
  "github.com/kylelemons/go-gypsy/yaml"
)

var (
  yamlConfig = new(yaml.File)
  configFile = flag.String("config", "config.yaml", "YAML config for Falcon")
)

func InitShellParser() {
  flag.Parse()
  logger.Info("Using config file", *configFile)

  yamlConfig, err := yaml.ReadFile(*configFile)
  if err != nil {
    logger.FatalError("Error read file: readfile(%q): %s", *configFile, err)
  }
  val, err := yamlConfig.Get("mapping.key1")
  if err != nil {
    logger.FatalError("%-*s = %s\n", err)
  } else {
    logger.Info("Val", val)
  }
}

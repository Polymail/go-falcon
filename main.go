package main

import (
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/utils"
)

func main() {
  log.StartupInfo()
  utils.InitShellParser()
}

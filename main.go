package main

import (
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/utils"
  stdlog "log"
  "os"
)

func main() {
  log.SetTarget(stdlog.New(os.Stdout, "", stdlog.LstdFlags))
  log.StartupInfo()
  utils.InitShellParser()
}

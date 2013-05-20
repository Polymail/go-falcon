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
  // parse shell and config
  config, err := utils.InitShellParser()
  if err != nil {
    return
  }
  // begin
  log.Noticef("\n%v\n\n", config)
}

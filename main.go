package main

import (
  "github.com/le0pard/go-falcon/logger"
  "github.com/le0pard/go-falcon/utils"
)

func main() {
  logger.StartupInfo()
  utils.initShellParser()
}

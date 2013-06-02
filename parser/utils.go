package parser

import (
  "regexp"
)


func fixInvalidUnquotedAttachmentName(str string) string {
  reg := regexp.MustCompile(`(.)*\s(filename|name)=[^"](.+\s)+[^"]`)
  if reg.MatchString(str) {
    reg := regexp.MustCompile(`[^=]+$`)
    str = reg.ReplaceAllString(str, "\"$0\"")
  }
  return str
}
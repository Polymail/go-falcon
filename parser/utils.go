package parser

import (
  "regexp"
)

func fixMailEncodedHeader(str string) string {
  str = fixInvalidUnquotedAttachmentName(str)
  return str
}


func fixInvalidUnquotedAttachmentName(str string) string {
  reg := regexp.MustCompile(`(.)*\s(filename|name)=[^"](.+\s)+[^"]`)
  if reg.MatchString(str) {
    reg := regexp.MustCompile(`[^=]+$`)
    str = reg.ReplaceAllString(str, "\"$0\"")
  }
  return str
}
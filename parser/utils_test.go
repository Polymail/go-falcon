package parser

import (
  "testing"
)

type mimeUnquotedNameHeaderTest struct {
  From      string
  To        string
}

var mimeUnquotedNameHeaderTests = []mimeUnquotedNameHeaderTest{
  {"Content-Type: text/plain; name=test.txt", "Content-Type: text/plain; name=test.txt"},
  {"Content-Type: image/png; name=test-with-dash.png", "Content-Type: image/png; name=test-with-dash.png"},
  {"Content-Type: text/plain; name=This is a test.txt", "Content-Type: text/plain; name=\"This is a test.txt\""},
  {"Content-Disposition: attachment;\n   filename=This is a test.txt", "Content-Disposition: attachment;\n   filename=\"This is a test.txt\""},
}

func TestMimeUnquotedNameHeader(t *testing.T) {
  for _, header := range mimeUnquotedNameHeaderTests {
    decodedValue := fixInvalidUnquotedAttachmentName(header.From)
    expectEq(t, header.To, decodedValue, "Value of decoded with name header")
  }
}
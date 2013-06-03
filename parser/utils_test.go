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


type mimeHeaderTest struct {
  From      string
  To        string
}

var mimeHeaderTests = []mimeHeaderTest{
  {"=?iso-8859-1?q?J=F6rg_Doe?=", "Jörg Doe"},
  {"=?utf-8?q?J=C3=B6rg_Doe?=", "Jörg Doe"},
  {"=?ISO-8859-1?Q?Andr=E9?=", "André"},
  {"=?ISO-8859-1?B?SvZyZw==?=", "Jörg"},
  {"=?UTF-8?B?SsO2cmc=?=", "Jörg"},
  {"illness notification =?8bit?Q?ALPH=C3=89E?=", "illness notification ALPHÉE"},
  {"=?UTF-8?B?44G+44G/44KA44KB44KC?=", "まみむめも"},
}

func TestMimeHeaderDecode(t *testing.T) {
  for _, header := range mimeHeaderTests {
    decodedHeader, err := MimeHeaderDecode(header.From)
    if err != nil {
      t.Error("Error in parsing header: %v", err)
    } else {
      expectEq(t, header.To, decodedHeader, "Value of decoded header")
    }
  }
}
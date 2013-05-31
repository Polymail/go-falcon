package parser

import (
  "testing"
)


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
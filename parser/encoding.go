package parser

import (
  "bytes"
  "encoding/base64"
  "regexp"
  "io/ioutil"
  "strings"
  "github.com/qiniu/iconv"
  "github.com/sloonz/go-qprintable"
)

// Decode strings in Mime header format
// eg. =?ISO-2022-JP?B?GyRCIVo9dztSOWJAOCVBJWMbKEI=?=
func MimeHeaderDecode(str string) string {
  reg, _ := regexp.Compile(`=\?(.+?)\?([QBqp])\?(.+?)\?=`)
  matched := reg.FindAllStringSubmatch(str, -1)
  var charset, encoding, payload string
  if matched != nil {
    for i := 0; i < len(matched); i++ {
      if len(matched[i]) > 2 {
        charset = matched[i][1]
        encoding = strings.ToUpper(matched[i][2])
        payload = matched[i][3]
        switch encoding {
        case "B":
          str = strings.Replace(str, matched[i][0], mailTransportDecode(payload, "base64", charset), 1)
        case "Q":
          str = strings.Replace(str, matched[i][0], mailTransportDecode(payload, "quoted-printable", charset), 1)
        }
      }
    }
  }
  return str
}

// decode from 7bit to 8bit UTF-8
func mailTransportDecode(pbody string, encodingType string, charset string) string {
  if charset == "" {
    charset = "UTF-8"
  } else {
    charset = strings.ToUpper(charset)
  }
  if strings.ToLower(encodingType) == "base64" {
    pbody = decodeFromBase64(pbody)
  } else if encodingType == "quoted-printable" {
    pbody = fromQuotedP(pbody)
  }
  if charset != "UTF-8" {
    charset = decodeFixCharset(charset)
    cd, err := iconv.Open(charset, "UTF-8")
    if err != nil {
      return pbody
    }
    defer cd.Close()
    return cd.ConvString(pbody)
  }
  return pbody
}

func fromQuotedP(data string) string {
  buf := bytes.NewBufferString(data)
  decoder := qprintable.NewDecoder(qprintable.BinaryEncoding, buf)
  res, _ := ioutil.ReadAll(decoder)
  return string(res)
}

func decodeFromBase64(data string) string {
  buf := bytes.NewBufferString(data)
  decoder := base64.NewDecoder(base64.StdEncoding, buf)
  res, _ := ioutil.ReadAll(decoder)
  return string(res)
}

func decodeFixCharset(charset string) string {
  reg, _ := regexp.Compile(`[_:.\/\\]`)
  fixedCharset := reg.ReplaceAllString(charset, "-")
  // Fix charset
  // borrowed from http://squirrelmail.svn.sourceforge.net/viewvc/squirrelmail/trunk/squirrelmail/include/languages.php?revision=13765&view=markup
  // OE ks_c_5601_1987 > cp949
  fixedCharset = strings.Replace(fixedCharset, "ks-c-5601-1987", "cp949", -1)
  // Moz x-euc-tw > euc-tw
  fixedCharset = strings.Replace(fixedCharset, "x-euc", "euc", -1)
  // Moz x-windows-949 > cp949
  fixedCharset = strings.Replace(fixedCharset, "x-windows_", "cp", -1)
  // windows-125x and cp125x charsets
  fixedCharset = strings.Replace(fixedCharset, "windows-", "cp", -1)
  // ibm > cp
  fixedCharset = strings.Replace(fixedCharset, "ibm", "cp", -1)
  // iso-8859-8-i -> iso-8859-8
  fixedCharset = strings.Replace(fixedCharset, "iso-8859-8-i", "iso-8859-8", -1)
  if charset != fixedCharset {
    return fixedCharset
  }
  return charset
}

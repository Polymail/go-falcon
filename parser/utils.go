package parser

import (
  "bytes"
  "net/url"
  "io/ioutil"
  "strings"
  "regexp"
  "code.google.com/p/mahonia"
  "github.com/sloonz/go-iconv"
  "github.com/sloonz/go-qprintable"
  "github.com/le0pard/go-falcon/utils"
)

var (
  invalidUnquotedRE = regexp.MustCompile(`(.)*\s(filename|name)=[^"](.+\s)+[^"]`)
  invalidUnquotedResRE = regexp.MustCompile(`[^=]+$`)
  invalidEscapedRE = regexp.MustCompile(`name\*[[0-9]*]?=iso-2022-jp'ja'(.*)`)
  mahoniaEnc = mahonia.NewEncoder("iso2022jp")
  invalidContentIdRE = regexp.MustCompile(`<(.*)>`)
)

// fix escaped and unquoted headers values

func FixMailEncodedHeader(str string) string {
  str = fixInvalidUnquotedAttachmentName(str)
  str = fixInvalidEscapedAttachmentName(str)
  return str
}


func fixInvalidUnquotedAttachmentName(str string) string {
  if invalidUnquotedRE.MatchString(str) {
    str = invalidUnquotedResRE.ReplaceAllString(str, "\"$0\"")
  }
  return str
}


func fixInvalidEscapedAttachmentName(str string) string {
  var words []string
  arrayStr := strings.Split(str, " ")
  for _, word := range arrayStr {
    if invalidEscapedRE.MatchString(word) {
      unescapedStr, err := url.QueryUnescape(word)
      if err == nil {
        unescapedStr = mahoniaEnc.ConvertString(unescapedStr)
        unescapedStr = invalidEscapedRE.ReplaceAllString(unescapedStr, "name=\"$1\"")
        word = unescapedStr
      }
    }
    words = append(words, word)
  }
  return strings.Join(words, " ")
}

// encode Mime

func MimeHeaderDecode(str string) string {
  reg := regexp.MustCompile(`=\?(.+?)\?([QBqp])\?(.+?)\?=`)
  matched := reg.FindAllStringSubmatch(str, -1)
  if matched != nil {
    for _, word := range matched {
      if len(word) > 2 {
        switch strings.ToUpper(word[2]) {
          case "B":
            str = strings.Replace(str, word[0], FixEncodingAndCharsetOfPart(word[3], "base64", word[1]), 1)
          case "Q":
            str = strings.Replace(str, word[0], FixEncodingAndCharsetOfPart(word[3], "quoted-printable", word[1]), 1)
        }
      }
    }
  }
  return str
}


// fix email body

func FixEncodingAndCharsetOfPart(data, contentEncoding, contentCharset string) string {
  // encoding
  if contentEncoding == "quoted-printable" {
    data = strings.Replace(fromQuotedP(data), "_", " ", -1)
  } else if contentEncoding == "base64" {
    data = utils.DecodeBase64(data)
  }
  // charset
  if contentCharset == "" {
    contentCharset = "utf-8"
  } else {
    contentCharset = strings.ToLower(contentCharset)
  }
  if contentCharset != "utf-8" {
    switch contentCharset {
    case "iso-8859-1":
      b := new(bytes.Buffer)
      for _, c := range []byte(data) {
        b.WriteRune(rune(c))
      }
      return b.String()
    default:
      // eg. charset can be "ISO-2022-JP"
      convstr, err := iconv.Conv(data, "UTF-8", strings.ToUpper(fixCharset(contentCharset)))
      if err == nil {
        return convstr
      }
    }
  }
  // result
  return data
}

// quoted-printable

func fromQuotedP(data string) string {
  buf := bytes.NewBufferString(data)
  decoder := qprintable.NewDecoder(qprintable.BinaryEncoding, buf)
  res, _ := ioutil.ReadAll(decoder)
  return string(res)
}

// fix charset

func fixCharset(charset string) string {
  reg := regexp.MustCompile(`[_:.\/\\]`)
  fixed_charset := reg.ReplaceAllString(charset, "-")
  // Fix charset
  // borrowed from http://squirrelmail.svn.sourceforge.net/viewvc/squirrelmail/trunk/squirrelmail/include/languages.php?revision=13765&view=markup
  // OE ks_c_5601_1987 > cp949
  fixed_charset = strings.Replace(fixed_charset, "ks-c-5601-1987", "cp949", -1)
  // Moz x-euc-tw > euc-tw
  fixed_charset = strings.Replace(fixed_charset, "x-euc", "euc", -1)
  // Moz x-windows-949 > cp949
  fixed_charset = strings.Replace(fixed_charset, "x-windows_", "cp", -1)
  // windows-125x and cp125x charsets
  fixed_charset = strings.Replace(fixed_charset, "windows-", "cp", -1)
  // ibm > cp
  fixed_charset = strings.Replace(fixed_charset, "ibm", "cp", -1)
  // iso-8859-8-i -> iso-8859-8
  fixed_charset = strings.Replace(fixed_charset, "iso-8859-8-i", "iso-8859-8", -1)
  if charset != fixed_charset {
    return fixed_charset
  }
  return charset
}

// invalid content ID

func getInvalidContentId(contentId string) string {
  if invalidContentIdRE.MatchString(contentId) {
    res := invalidContentIdRE.FindStringSubmatch(contentId)
    if len(res) == 2 {
      contentId = res[1]
    }
  }
  return contentId
}

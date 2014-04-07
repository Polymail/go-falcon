package parser

import (
  "bytes"
  "net/url"
  "io/ioutil"
  "strings"
  "regexp"
  "code.google.com/p/go.text/encoding/charmap"
  "code.google.com/p/go.text/encoding/japanese"
  "code.google.com/p/go.text/encoding/traditionalchinese"
  "code.google.com/p/go.text/encoding/simplifiedchinese"
  "code.google.com/p/go.text/encoding/korean"
  "code.google.com/p/go.text/transform"
  "github.com/sloonz/go-iconv"
  "github.com/le0pard/go-falcon/utils"
)

var (
  invalidUnquotedRE = regexp.MustCompile(`(.)*\s(filename|name)=[^"](.+\s)+[^"]`)
  invalidUnquotedResRE = regexp.MustCompile(`[^=]+$`)
  invalidEscapedRE = regexp.MustCompile(`name\*[[0-9]*]?=iso-2022-jp'ja'(.*)`)
  mimeHeaderRE = regexp.MustCompile(`=\?(.+?)\?([QBqp])\?(.+?)\?=`)
  fixCharsetRE = regexp.MustCompile(`[_:.\/\\]`)
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
        sr := strings.NewReader(unescapedStr)
        tr, err := ioutil.ReadAll(transform.NewReader(sr, japanese.ISO2022JP.NewDecoder()))
        if err == nil {
          unescapedStr = string(tr)
        }
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
  matched := mimeHeaderRE.FindAllStringSubmatch(str, -1)
  if matched != nil {
    for _, word := range matched {
      if len(word) > 2 {
        switch strings.ToUpper(word[2]) {
          case "B":
            str = strings.Replace(str, word[0], FixEncodingAndCharsetOfPart(word[3], "base64", word[1], true), 1)
          case "Q":
            str = strings.Replace(str, word[0], FixEncodingAndCharsetOfPart(strings.Replace(word[3], "_", " ", -1), "quoted-printable", word[1], true), 1)
        }
      }
    }
  }
  return str
}


// fix email body

func FixEncodingAndCharsetOfPart(data, contentEncoding, contentCharset string, checkOnInvalidUtf bool) string {
  // encoding
  if contentEncoding == "quoted-printable" {
    data = fromQuotedP(data)
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
    case "7bit", "8bit":
      return data
    case "iso-8859-1":
      b := new(bytes.Buffer)
      for _, c := range []byte(data) {
        b.WriteRune(rune(c))
      }
      return b.String()
    case "shift-jis", "iso-2022-jp", "big5", "gb2312", "iso-8859-2", "iso-8859-6", "iso-8859-8", "koi8-r", "koi8-u", "windows-1251", "euc-kr":
      decoder := japanese.ShiftJIS.NewDecoder()
      switch contentCharset {
      case "iso-2022-jp":
        decoder = japanese.ISO2022JP.NewDecoder()
      case "big5":
        decoder = traditionalchinese.Big5.NewDecoder()
      case "gb2312":
        decoder = simplifiedchinese.HZGB2312.NewDecoder()
      case "iso-8859-2":
        decoder = charmap.ISO8859_2.NewDecoder()
      case "iso-8859-6":
        decoder = charmap.ISO8859_6.NewDecoder()
      case "iso-8859-8":
        decoder = charmap.ISO8859_8.NewDecoder()
      case "koi8-r":
        decoder = charmap.KOI8R.NewDecoder()
      case "koi8-u":
        decoder = charmap.KOI8U.NewDecoder()
      case "windows-1251":
        decoder = charmap.Windows1251.NewDecoder()
      case "euc-kr":
        decoder = korean.EUCKR.NewDecoder()
      default:
        decoder = japanese.ShiftJIS.NewDecoder()
      }
      tr, err := ioutil.ReadAll(transform.NewReader(strings.NewReader(data), decoder))
      if err == nil {
        data = string(tr)
      } else {
        convstr, err := convertByIconv(data, contentCharset)
        if err == nil {
          data = convstr
        }
      }
    default:
      convstr, err := convertByIconv(data, contentCharset)
      if err == nil {
        data = convstr
      }
    }
  }
  // valid utf
  if checkOnInvalidUtf {
    data = utils.CheckAndFixUtf8(data)
  }
  // result
  return data
}

func convertByIconv(data, contentCharset string) (string, error) {
  return iconv.Conv(data, "UTF-8", strings.ToUpper(fixCharset(contentCharset)))
}

// quoted-printable

func fromQuotedP(data string) string {
  buf := bytes.NewBufferString(data)
  decoder := utils.NewQuotedPrintableReader(buf)
  res, _ := ioutil.ReadAll(decoder)
  return string(res)
}

// fix charset

func fixCharset(charset string) string {
  fixedCharset := fixCharsetRE.ReplaceAllString(charset, "-")
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

// invalid from/to email

func getInvalidFromToHeader(header string) string {
  if invalidContentIdRE.MatchString(header) {
    res := invalidContentIdRE.FindStringSubmatch(header)
    if len(res) == 2 {
      return res[1]
    }
  }
  return header
}

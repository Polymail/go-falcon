package parser

import (
  "bytes"
  "net/url"
  "io/ioutil"
  "strings"
  "errors"
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

func MimeHeaderDecode(str string) (string, error) {
  var (
    words []string
    err error
  )
  arrayStr := strings.Split(str, " ")
  for _, word := range arrayStr {
    if strings.HasPrefix(word, "=?") && strings.HasSuffix(word, "?=") && strings.Count(word, "?") == 4 {
      word, err = decodeRFC2047Word(word)
    }
    if err == nil {
      words = append(words, word)
    }
  }
  phrase := strings.Join(words, " ")
  return phrase, nil
}


func decodeRFC2047Word(s string) (string, error) {
  fields := strings.Split(s, "?")
  if len(fields) != 5 || fields[0] != "=" || fields[4] != "=" {
    return "", errors.New("header not RFC 2047 encoded")
  }
  charset, enc := strings.ToLower(fields[1]), strings.ToLower(fields[2])

  contentEncoding := ""
  switch enc {
  case "b":
    contentEncoding = "base64"
  case "q":
    contentEncoding = "quoted-printable"
  }

  return FixEncodingAndCharsetOfPart(fields[3], contentEncoding, charset), nil
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
    contentCharset = "UTF-8"
  } else {
    contentCharset = strings.ToUpper(contentCharset)
  }
  if contentCharset != "UTF-8" {
    switch contentCharset {
    case "ISO-8859-1":
      b := new(bytes.Buffer)
      for _, c := range []byte(data) {
        b.WriteRune(rune(c))
      }
      return b.String()
    default:
      // eg. charset can be "ISO-2022-JP"
      convstr, err := iconv.Conv(data, "UTF-8", fixCharset(contentCharset))
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
	reg, _ := regexp.Compile(`[_:.\/\\]`)
	fixed_charset := reg.ReplaceAllString(strings.ToLower(charset), "-")
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
  // fix
  fixed_charset = strings.ToUpper(fixed_charset)
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
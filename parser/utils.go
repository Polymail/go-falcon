package parser

import (
  "bytes"
  "encoding/base64"
  "net/url"
  "io"
  "io/ioutil"
  "strings"
  "errors"
  "strconv"
  "regexp"
  "code.google.com/p/mahonia"
)

// fix escaped and unquoted headers values

func FixMailEncodedHeader(str string) string {
  str = fixInvalidUnquotedAttachmentName(str)
  str = fixInvalidEscapedAttachmentName(str)
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

func fixInvalidEscapedAttachmentName(str string) string {
  reg := regexp.MustCompile(`name\*[[0-9]*]?=iso-2022-jp'ja'(.*)`)
  if reg.MatchString(str) {
    unescapedStr, err := url.QueryUnescape(str)
    if err != nil {
      return str
    }
    str = unescapedStr
    enc := mahonia.NewEncoder("iso2022jp")
    str = enc.ConvertString(str)
    str = reg.ReplaceAllString(str, "name=\"$1\"")
    return str
  }
  return str
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

  in := bytes.NewBufferString(fields[3])
  var r io.Reader
  switch enc {
  case "b":
    r = base64.NewDecoder(base64.StdEncoding, in)
  case "q":
    r = qDecoder{r: in}
  default:
    return "", errors.New("mail: RFC 2047 encoding not supported: " + enc)
  }

  dec, err := ioutil.ReadAll(r)
  if err != nil {
    return "", err
  }

  switch charset {
  case "iso-8859-1":
    b := new(bytes.Buffer)
    for _, c := range dec {
      b.WriteRune(rune(c))
    }
    return b.String(), nil
  default:
    return string(dec), nil
  }
  panic("unreachable")
}

type qDecoder struct {
  r       io.Reader
  scratch [2]byte
}

func (qd qDecoder) Read(p []byte) (n int, err error) {
  // This method writes at most one byte into p.
  if len(p) == 0 {
          return 0, nil
  }
  if _, err := qd.r.Read(qd.scratch[:1]); err != nil {
          return 0, err
  }
  switch c := qd.scratch[0]; {
  case c == '=':
    if _, err := io.ReadFull(qd.r, qd.scratch[:2]); err != nil {
      return 0, err
    }
    x, err := strconv.ParseInt(string(qd.scratch[:2]), 16, 64)
    if err != nil {
      return 0, errors.New("mail: invalid RFC 2047 encoding: " + string(qd.scratch[:2]))
    }
    p[0] = byte(x)
  case c == '_':
    p[0] = ' '
  default:
    p[0] = c
  }
  return 1, nil
}

var atextChars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
  "abcdefghijklmnopqrstuvwxyz" +
  "0123456789" +
  "!#$%&'*+-/=?^_`{|}~")

// isAtext returns true if c is an RFC 5322 atext character.
// If dot is true, period is included.
func isAtext(c byte, dot bool) bool {
  if dot && c == '.' {
    return true
  }
  return bytes.IndexByte(atextChars, c) >= 0
}

// isQtext returns true if c is an RFC 5322 qtest character.
func isQtext(c byte) bool {
  // Printable US-ASCII, excluding backslash or quote.
  if c == '\\' || c == '"' {
    return false
  }
  return '!' <= c && c <= '~'
}

// isVchar returns true if c is an RFC 5322 VCHAR character.
func isVchar(c byte) bool {
  // Visible (printing) characters.
  return '!' <= c && c <= '~'
}







/*

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

*/
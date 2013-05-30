package parser

import (
  "bytes"
  "encoding/base64"
  "io"
  "io/ioutil"
  "strings"
  "errors"
  "strconv"
)

func MimeHeaderDecode(str string) (string, error) {
  arrayStr := strings.Split(str, " ")
  var (
    words []string
    err error
  )
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
  if charset != "iso-8859-1" && charset != "utf-8" && charset != "8bit" && charset != "7bit" {
    return "", errors.New("charset not supported: " + charset)
  }

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
  case "utf-8", "8bit", "7bit":
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
package parser

import (
  "bytes"
  "encoding/base64"
  "regexp"
  "io"
  "io/ioutil"
  "strings"
  "errors"
  "strconv"
  "github.com/qiniu/iconv"
)

// decode from 7bit to 8bit UTF-8
func mailTransportDecode(pbody string, encodingType string, charset string) string {
  if charset == "" {
    charset = "UTF-8"
  } else {
    charset = strings.ToUpper(charset)
  }
  if strings.ToLower(encodingType) == "base64" {
    pbody = decodeFromBase64(pbody)
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


func decodeRFC2047Word(s string) (string, error) {
    fields := strings.Split(s, "?")
    if len(fields) != 5 || fields[0] != "=" || fields[4] != "=" {
      return "", errors.New("mail: address not RFC 2047 encoded")
    }
    charset, enc := strings.ToLower(fields[1]), strings.ToLower(fields[2])
    if charset != "iso-8859-1" && charset != "utf-8" {
      return "", errors.New("mail: charset not supported: " + charset)
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
    case "utf-8":
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

package utils

import (
  "bytes"
  "io/ioutil"
  "encoding/base64"
)

func DecodeSMTPAuthPlain(b64 string) (string, string, string) {
  dest := DecodeBase64String(b64)
  // zero byte
  var zero []byte
  zero = make([]byte, 1)
  zero[0] = 0
  f := bytes.Split([]byte(dest), zero)

  if((len(f) == 4) || (len(f) == 3)) {
    return string(f[0]), string(f[1]), string(f[2])
  } else {
    return "","",""
  }

  return "","",""
}

func DecodeBase64String(b64 string) string {
  buf := bytes.NewBufferString(b64)
  encoded := base64.NewDecoder(base64.StdEncoding, buf)
  dest, _ := ioutil.ReadAll(encoded)
  return string(dest)
}
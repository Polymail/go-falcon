package utils

import (
  "bytes"
//  "compress/zlib"
//  "crypto/md5"
//  "crypto/rand"
//  "crypto/tls"
  "encoding/base64"
//  "encoding/hex"
//  "encoding/json"
//  "io"
  "io/ioutil"
)


func Base64ToString(data string) string {
  buf := bytes.NewBufferString(data)
  decoder := base64.NewDecoder(base64.StdEncoding, buf)
  res, _ := ioutil.ReadAll(decoder)
  return string(res)
}
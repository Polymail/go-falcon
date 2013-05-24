package utils

import (
  "bytes"
  "io/ioutil"
  "encoding/base64"
  "strings"
  "crypto/hmac"
  "crypto/md5"
  "strconv"
  "math/rand"
  "time"
  "os"
  "github.com/le0pard/go-falcon/log"
)

func randomString (l int ) string {
    bytes := make([]byte, l)
    for i:=0 ; i<l ; i++ {
        bytes[i] = byte(randInt(65, 90))
    }
    return string(bytes)
}

func randInt(min int , max int) int {
    return min + rand.Intn(max-min)
}

func generateRandString(l int) string {
  rand.Seed(time.Now().UTC().UnixNano())
  return randomString(l)
}

// generate challenge for cram-md5

func GenerateSMTPCramMd5(hostname string) []byte {
  randStr := strconv.Itoa(os.Getppid()) + "." + strconv.Itoa(int(time.Now().UTC().UnixNano()))
  return EncodeBase64("<" + randStr + "@" + hostname + ">")
}

// decode cram-md5

func DecodeSMTPCramMd5(b64 string) (string, string) {
  str := DecodeBase64(b64)
  f := strings.Split(string(str), " ")
  if len(f) == 2 {
    return f[0], f[1]
  }
  return "", ""
}

// check cram-md5

func CheckCramMd5Pass(rawPassword string, cramPassword string, cramSecret []byte) bool {
  if cramSecret != nil {
    d := hmac.New(md5.New, []byte(rawPassword))
    d.Write(cramSecret)
    s := make([]byte, 0, d.Size())
    expectedMAC := d.Sum(s)
    log.Debugf("%s", cramSecret)
    log.Debugf("%x == %x", []byte(cramPassword), expectedMAC)
    return hmac.Equal([]byte(cramPassword), expectedMAC)
  } else {
    return (cramPassword == rawPassword)
  }
}

// Decode smtp plain auth

func DecodeSMTPAuthPlain(b64 string) (string, string, string) {
  dest := DecodeBase64(b64)
  f := bytes.Split(dest, []byte{ 0 })

  if ((len(f) == 4) || (len(f) == 3)) {
    return string(f[0]), string(f[1]), string(f[2])
  }

  return "","",""
}

// encode base64

func EncodeBase64(b64 string) []byte {
  var b bytes.Buffer
  w := base64.NewEncoder(base64.StdEncoding, &b)
  w.Write([]byte(b64))
  w.Close()
  return b.Bytes()
}

// decode base64

func DecodeBase64(b64 string) []byte {
  buf := bytes.NewBufferString(b64)
  encoded := base64.NewDecoder(base64.StdEncoding, buf)
  dest, _ := ioutil.ReadAll(encoded)
  return dest
}
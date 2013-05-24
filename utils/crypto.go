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

func GenerateSMTPCramMd5(hostname string) string {
  randStr := strconv.Itoa(os.Getppid()) + "." + strconv.Itoa(int(time.Now().UTC().UnixNano()))
  return "<" + randStr + "@" + hostname + ">"
}

// decode cram-md5

func DecodeSMTPCramMd5(b64 string) (string, string) {
  str := DecodeBase64String(b64)
  f := strings.Split(str, " ")
  if len(f) == 2 {
    return f[0], f[1]
  }
  return "", ""
}

// check cram-md5

func CheckCramMd5Pass(password, decodedPassword, cramSecret string) bool {
  if cramSecret != "" {
    mac := hmac.New(md5.New, []byte(password))
    mac.Write([]byte(EncodeBase64String(cramSecret)))
    s := make([]byte, 0, mac.Size())
    expectedMAC := mac.Sum(s)
    log.Debugf("%v == %v == %v", password, decodedPassword, cramSecret)
    log.Debugf("%v == %v", decodedPassword, string(expectedMAC))
    return hmac.Equal([]byte(decodedPassword), expectedMAC)
  } else {
    return (decodedPassword == password)
  }
}

// Decode smtp plain auth

func DecodeSMTPAuthPlain(b64 string) (string, string, string) {
  dest := DecodeBase64String(b64)
  f := bytes.Split([]byte(dest), []byte{ 0 })

  if ((len(f) == 4) || (len(f) == 3)) {
    return string(f[0]), string(f[1]), string(f[2])
  }

  return "","",""
}

// encode base64

func EncodeBase64String(b64 string) string {
  var b bytes.Buffer
  w := base64.NewEncoder(base64.StdEncoding, &b)
  w.Write([]byte(b64))
  w.Close()
  return string(b.Bytes())
}

// decode base64

func DecodeBase64String(b64 string) string {
  buf := bytes.NewBufferString(b64)
  encoded := base64.NewDecoder(base64.StdEncoding, buf)
  dest, _ := ioutil.ReadAll(encoded)
  return string(dest)
}
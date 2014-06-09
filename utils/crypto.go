package utils

import (
  "bytes"
  "crypto/hmac"
  "crypto/md5"
  "encoding/base64"
  "fmt"
  "math/rand"
  "os"
  "strconv"
  "strings"
  "time"
)

func randomString(l int) string {
  bytes := make([]byte, l)
  for i := 0; i < l; i++ {
    bytes[i] = byte(randInt(65, 90))
  }
  return string(bytes)
}

func randInt(min int, max int) int {
  return min + rand.Intn(max-min)
}

func GenerateRandString(l int) string {
  rand.Seed(time.Now().UTC().UnixNano())
  return randomString(l)
}

func GetRandFromArray(arr []int) int {
  rand.Seed(time.Now().UTC().UnixNano())
  return arr[rand.Intn(len(arr))]
}

// generate challenge for cram-md5

func GenerateProtocolCramMd5(hostname string) string {
  randStr := strconv.Itoa(os.Getppid()) + "." + strconv.Itoa(int(time.Now().UTC().UnixNano()))
  return "<" + randStr + "@" + hostname + ">"
}

// decode cram-md5

func DecodeProtocolCramMd5(b64 string) (string, string) {
  f := strings.Split(DecodeBase64(b64), " ")
  if len(f) == 2 {
    return f[0], f[1]
  }
  return "", ""
}

// check passwords

func CheckProtocolAuthPass(authMethod, rawPassword, cramPassword, cramSecret string) bool {
  switch authMethod {
  case AUTH_CRAM_MD5:
    d := hmac.New(md5.New, []byte(rawPassword))
    d.Write([]byte(cramSecret))
    s := make([]byte, 0, d.Size())
    expectedMAC := d.Sum(s)
    macIn16bit := []byte(fmt.Sprintf("%x", expectedMAC))
    return hmac.Equal([]byte(cramPassword), macIn16bit)
  case AUTH_APOP:
    var buffer bytes.Buffer
    buffer.WriteString(cramSecret)
    buffer.WriteString(rawPassword)
    h := md5.New()
    h.Write([]byte(buffer.String()))
    macIn16bit := []byte(fmt.Sprintf("%x", h.Sum(nil)))
    return bytes.Equal([]byte(cramPassword), macIn16bit)
  default:
    return (cramPassword == rawPassword)
  }
}

// Decode protocol plain auth

func DecodeProtocolAuthPlain(b64 string) (string, string, string) {
  f := bytes.Split([]byte(DecodeBase64(b64)), []byte{0})

  if (len(f) == 4) || (len(f) == 3) {
    return string(f[0]), string(f[1]), string(f[2])
  }

  return "", "", ""
}

// encode base64

func EncodeBase64(b64 string) string {
  return base64.StdEncoding.EncodeToString([]byte(b64))
}

// decode base64

func DecodeBase64(b64 string) string {
  dest, err := base64.StdEncoding.DecodeString(b64)
  if err != nil {
    return ""
  }
  return string(dest)
}

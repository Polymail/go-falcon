package utils

import (
  "unicode/utf8"
)

// check invalid utf-8 symbols
func CheckAndFixUtf8(data string) string {
  if !utf8.Valid([]byte(data)) {
    v := make([]rune, 0, len(data))
    for i, r := range data {
      if r == utf8.RuneError {
        _, size := utf8.DecodeRuneInString(data[i:])
        if size == 1 {
          continue
        }
      }
      v = append(v, r)
    }
    data = string(v)
  }
  return data
}

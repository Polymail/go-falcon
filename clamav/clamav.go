// http://www.clamav.net/doc/latest/clamdoc.pdf
// TODO: NOT FINISHED
package clamav

import (
  "net"
  "bufio"
  "strconv"
  "io"
  "strings"
  "regexp"
  "github.com/le0pard/go-falcon/config"
)

const CHUNK_SIZE = 1024

type Clamav struct {
  config *config.Config
  RawEmail  []byte
}

func CheckEmailForViruses(config *config.Config, email []byte) (string, error) {
  clamav := &Clamav{
    config: config,
    RawEmail: email,
  }
  output, err := clamav.checkEmail()
  if err != nil {
    return "", err
  }
  return clamav.parseOutput(output), nil
}

// check email by spamassassin

func (ss *Clamav) checkEmail() ([]string, error) {
  var dataArrays []string
  conn, err := net.Dial("tcp", ss.config.Clamav.Host + ":" + strconv.Itoa(ss.config.Clamav.Port))
  if err != nil {
    return dataArrays, err
  }
  // check email
  if len(ss.RawEmail) <= 0 {
    return dataArrays, nil
  }
  // write headers
  _, err = conn.Write([]byte("nINSTREAM\n"))
  if err != nil {
    return dataArrays, err
  }
  chunkPos := 0
  for {
    if chunkPos + CHUNK_SIZE >= len(ss.RawEmail) {
      data := ss.RawEmail[chunkPos:len(ss.RawEmail)]
      err = sendChunkOfData(conn, data)
      if err != nil {
        return dataArrays, err
      }
      break
    } else {
      data := ss.RawEmail[chunkPos:chunkPos + CHUNK_SIZE]
      err = sendChunkOfData(conn, data)
      if err != nil {
        return dataArrays, err
      }
      chunkPos = chunkPos + CHUNK_SIZE + 1
    }
  }
  // write end
  _, err = conn.Write([]byte{0, 0, 0, 0})
  if err != nil {
    return dataArrays, err
  }
  // read data
  reader := bufio.NewReader(conn)
  // reading
  for {
    line, err := reader.ReadString('\n')
    if err == io.EOF {
      conn.Close()
      break
    }
    if err != nil {
      conn.Close()
      return dataArrays, err
    }
    line = strings.TrimRight(line, " \t\r\n")
    dataArrays = append(dataArrays, line)
  }

  return dataArrays, nil
}

func sendChunkOfData(conn net.Conn, data []byte) error {
  lenData := len(data)
  var buf [4]byte
  buf[0] = byte(lenData >> 24)
  buf[1] = byte(lenData >> 16)
  buf[2] = byte(lenData >> 8)
  buf[3] = byte(lenData >> 0)

  dataWrite := []byte{}
  dataWrite = append(dataWrite, buf[0], buf[1], buf[2], buf[3])
  dataWrite = append(dataWrite, data...)

  _, err := conn.Write(dataWrite)
  return err
}

// parse spamassassin output

func (ss *Clamav) parseOutput(output []string) string {
  reg := regexp.MustCompile(`(?i)^stream:([\s+]?)(.*)`)
  result := strings.Join(output, ", ")
  result = strings.Trim(result, " \t\r\n")
  if reg.MatchString(result) {
    res := reg.FindStringSubmatch(result)
    if len(res) >= 2 {
      if strings.ToLower(res[2]) == "ok" {
        return ""
      } else {
        return res[2]
      }
    } else {
      return ""
    }
  }
  return ""
}


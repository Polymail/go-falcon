// http://www.clamav.net/doc/latest/clamdoc.pdf
// TODO: NOT FINISHED
package clamav

import (
  "net"
  "errors"
  "bufio"
  "strconv"
  "io"
  "strings"
  "encoding/json"
  "github.com/le0pard/go-falcon/config"
)

const CHUNK_SIZE = 2048

type Clamav struct {
  config *config.Config
  RawEmail  []byte
}

type ClamavResponse struct {
  ResponseMessage       string
}

func CheckEmailForViruses(config *config.Config, email []byte) (string, error) {
  clamav := &Clamav{
    config: config,
    RawEmail: email,
  }
  output, err := clamav.CheckEmail()
  if err != nil {
    return "", err
  }
  response := clamav.parseOutput(output)
  jsonResult, err := json.Marshal(response)
  if err != nil {
    return "", err
  }
  return string(jsonResult), nil
}

// check email by spamassassin

func (ss *Clamav) CheckEmail() ([]string, error) {
  var dataArrays []string
  ip := net.ParseIP(ss.config.Clamav.Ip)
  if ip == nil {
    return dataArrays, errors.New("Invalid ip address")
  }
  addr := &net.TCPAddr{
    IP: ip,
    Port: ss.config.Clamav.Port,
  }
  conn, err := net.DialTCP("tcp", nil, addr)
  if err != nil {
    return dataArrays, err
  }
  // write headers
  _, err = conn.Write([]byte("zINSTREAM\0"))
  if err != nil {
    return dataArrays, err
  }
  _, err = conn.Write([]byte("Content-length: " + strconv.Itoa(len(ss.RawEmail)) + "\r\n\r\n"))
  if err != nil {
    return dataArrays, err
  }
  // write email
  _, err = conn.Write(ss.RawEmail)
  if err != nil {
    return dataArrays, err
  }
  // force close writer
  conn.CloseWrite()
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

// parse spamassassin output

func (ss *Clamav) parseOutput(output []string) *ClamavResponse {
  response := &ClamavResponse{}
  return response
}


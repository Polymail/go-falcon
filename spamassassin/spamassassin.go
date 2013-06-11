// http://svn.apache.org/repos/asf/spamassassin/trunk/spamd/PROTOCOL
package spamassassin

import (
  "net"
  "errors"
  "bufio"
  "strconv"
  "io"
  "strings"
  "regexp"
  "encoding/json"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
)

type Spamassassin struct {
  config *config.Config
  RawEmail  []byte
}

type SpamassassinHeader struct {
  Pts                   string
  RuleName              string
  Description           string
}

type SpamassassinResponse struct {
  ResponseCode          int
  ResponseMessage       string
  Score                 float64
  Spam                  bool
  Threshold             float64
  Details               []SpamassassinHeader
}

func CheckSpamEmail(config *config.Config, email []byte) (string, error) {
  spamassassin := &Spamassassin{
    config: config,
    RawEmail: email,
  }
  output, err := spamassassin.CheckEmail()
  if err != nil {
    return "", err
  }
  response := spamassassin.parseOutput(output)
  jsonResult, err := json.Marshal(response)
  if err != nil {
    return "", err
  }
  log.Debugf("Spam: %s", string(jsonResult))
  return string(jsonResult), nil
}

// check email by spamassassin

func (ss *Spamassassin) CheckEmail() ([]string, error) {
  var dataArrays []string
  ip := net.ParseIP(ss.config.Spamassassin.Ip)
  if ip == nil {
    return dataArrays, errors.New("Invalid ip address")
  }
  addr := &net.TCPAddr{
    IP: ip,
    Port: ss.config.Spamassassin.Port,
  }
  conn, err := net.DialTCP("tcp", nil, addr)
  if err != nil {
    return dataArrays, err
  }
  // write headers
  _, err = conn.Write([]byte("REPORT SPAMC/1.2\r\n"))
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

func (ss *Spamassassin) parseOutput(output []string) *SpamassassinResponse {
  response := &SpamassassinResponse{}
  regInfo, regSpam := regexp.MustCompile(`(.+)\/(.+) (\d+) (.+)`), regexp.MustCompile(`^Spam: (.+) ; (.+) . (.+)$`)
  regDetails := regexp.MustCompile(`^(-?[0-9]*[.][0-9])\s([a-zA-Z0-9_]*)\s(\w*)`)
  for _, row := range output {
    if regInfo.MatchString(row) {
      res := regInfo.FindStringSubmatch(row)
      resCode, err := strconv.Atoi(res[3])
      if err == nil {
        response.ResponseCode = resCode
      }
      response.ResponseMessage = res[4]
    }
    if regSpam.MatchString(row) {
      res := regSpam.FindStringSubmatch(row)
      if strings.ToLower(res[1]) == "true" || strings.ToLower(res[1]) == "yes" {
        response.Spam = true
      } else {
        response.Spam = false
      }
      resFloat, err := strconv.ParseFloat(res[2], 64)
      if err == nil {
        response.Score = resFloat
      }
      resFloat, err = strconv.ParseFloat(res[3], 64)
      if err == nil {
        response.Threshold = resFloat
      }
    }
    // details
    row = strings.TrimLeft(row, " \t\r\n")
    row = strings.TrimRight(row, " \t\r\n")
    if regDetails.MatchString(row) {
      res := regDetails.FindStringSubmatch(row)
      log.Debugf("%v", res)
      header := SpamassassinHeader{ Pts: res[1], RuleName: res[2], Description: res[3] }
      response.Details = append(response.Details, header)
    }
  }
  return response
}


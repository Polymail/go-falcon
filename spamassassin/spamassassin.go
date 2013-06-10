// http://svn.apache.org/repos/asf/spamassassin/trunk/spamd/PROTOCOL
package spamassassin

import (
  "net"
  "errors"
  "bufio"
  "strconv"
  "io"
  "strings"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
)

type Spamassassin struct {
  config *config.Config
  RawEmail  []byte
}

func CheckSpamEmail(config *config.Config, email []byte) {
  spamassassin := &Spamassassin{
    config: config,
    RawEmail: email,
  }
  output, err := spamassassin.CheckEmail()
  log.Debugf("Spam info: %v", output)
  log.Debugf("Spam err: %v", err)
}

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
  // mail
  mail := strings.Replace(string(ss.RawEmail), "\r\n", "\n", -1)
  // write headers
  _, err = conn.Write([]byte("REPORT SPAMC/1.2\r\n"))
  if err != nil {
    return dataArrays, err
  }
  _, err = conn.Write([]byte("Content-length: " + strconv.Itoa(len(mail)) + "\r\n\r\n"))
  if err != nil {
    return dataArrays, err
  }
  // write email
  _, err = conn.Write([]byte(mail))
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

/*

package main

import (
        "bufio"
        "fmt"
        "net"
        "os"
        "strconv"
        "strings"
)

func main() {
        host := os.Args[1]
        ip, _, _ := net.ParseCIDR(host)
        addr := &net.TCPAddr{
                IP: ip,
                Port: 783,
        }
        conn, err := net.DialTCP("tcp", nil, addr)
        checkError(err)
        str := `From: Private Person <me@fromdomain.com>
To: A Test User <test@todomain.com>
CC: <test2@todomain.com>
CC: <test3@todomain.com>
Subject: SMTP e-mail test

This is a test e-mail message.`
        str = strings.Replace(str, "\n", "\r\n", -1)
        _, err = conn.Write([]byte("CHECK SPAMC/1.1\r\n"))
        _, err = conn.Write([]byte("Content-length: " + strconv.Itoa(len(str)) + "\r\n\r\n"))
        _, err = conn.Write([]byte(str))
        conn.CloseWrite()
        fmt.Println("Read")
        reader := bufio.NewReader(conn)
        for {
                line, err := reader.ReadString('\n')
                fmt.Println(err)
                line = strings.TrimRight(line, " \t\r\n")
                fmt.Println(line)
                //data := make([]byte, 1024)
                //_, err := conn.Read(data)
                //fmt.Println(err)
                if err != nil {
                        conn.Close()
                        break

                }
        }
}
func checkError(err error) {
        if err != nil {
                fmt.Println("Fatal error ", err.Error())
        }
}

<nil>
SPAMD/1.1 0 EX_OK
<nil>
Spam: False ; 1.5 / 5.0
<nil>

*/
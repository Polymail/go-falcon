package spamassassin

import (
  "net"
  "errors"
  "bufio"
  "strconv"
  "io"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/storage"
)

type Spamassassin struct {
  config *config.Config
  RawEmail  []byte
}

func CheckSpamEmail(config *config.Config, email []byte, db *storage.DBConn, messageId int) {
  spamassassin := &Spamassassin{
    config: config,
    RawEmail: email,
  }
  spamassassin.CheckEmail()
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
  // write headers
  _, err = conn.Write([]byte("CHECK SPAMC/1.2\r\n"))
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
      return dataArrays, err
    }
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
        _, err = conn.Write([]byte("CHECK SPAMC/1.2\r\n"))
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

*/
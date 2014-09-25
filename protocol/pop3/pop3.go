// Package pop3 implements an pop3 server.
package pop3

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/le0pard/go-falcon/config"
	"github.com/le0pard/go-falcon/log"
	"github.com/le0pard/go-falcon/utils"
	"io"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Server is an SMTP server.
type Server struct {
	Addr         string        // TCP address to listen on, ":2525" if empty
	Hostname     string        // optional Hostname to announce; "" to use system hostname
	ReadTimeout  time.Duration // optional read timeout
	WriteTimeout time.Duration // optional write timeout

	TLSconfig *tls.Config // tls config

	ServerConfig *config.Config

	// OnNewConnection, if non-nil, is called on new connections.
	// If it returns non-nil, the connection is closed.
	OnNewConnection func(c Connection) error
}

// Connection is implemented by the SMTP library and provided to callers
// customizing their own Servers.
type Connection interface {
	Addr() net.Addr
}

// SERVER

func (srv *Server) hostname() string {
	if srv.Hostname != "" {
		return srv.Hostname
	}
	out, err := exec.Command("hostname").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// ListenAndServe listens on the TCP network address srv.Addr and then
// calls Serve to handle requests on incoming connections.  If
// srv.Addr is blank, ":25" is used.
func (srv *Server) ListenAndServe() error {
	addr := srv.Addr
	if addr == "" {
		addr = ":110"
	}
	ln, e := net.Listen("tcp", addr)
	if e != nil {
		return e
	}
	return srv.Serve(ln)
}

func (srv *Server) Serve(ln net.Listener) error {
	defer ln.Close()
	for {
		rw, e := ln.Accept()
		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				log.Errorf("pop3: Accept error: %v", e)
				continue
			}
			return e
		}
		sess, err := srv.newSession(rw)
		if err != nil {
			continue
		}
		go sess.serve()
	}
	panic("not reached")
}

// SESSION

type session struct {
	srv *Server
	rwc net.Conn
	br  *bufio.Reader
	bw  *bufio.Writer

	authPlain        bool   // bool for 2 step plain auth
	authLogin        bool   // bool for 2 step login auth
	authApopLogin    string // bytes for apop login
	authCramMd5Login string // bytes for cram-md5 login

	mailboxId    int    // id of mailbox
	authUsername string // auth login
	authPassword string // auth password

	cachedList    [][2]int
	isListFetched bool
}

func (srv *Server) newSession(rwc net.Conn) (s *session, err error) {
	s = &session{
		srv:              srv,
		rwc:              rwc,
		br:               bufio.NewReader(rwc),
		bw:               bufio.NewWriter(rwc),
		authPlain:        false,
		authLogin:        false,
		authCramMd5Login: "",
		mailboxId:        0,
		isListFetched:    false,
	}
	return
}

func (s *session) errorf(format string, args ...interface{}) {
	log.Errorf("Client error: "+format, args...)
}

func (s *session) sendf(format string, args ...interface{}) {
	if s.srv.WriteTimeout != 0 {
		s.rwc.SetWriteDeadline(time.Now().Add(s.srv.WriteTimeout))
	}
	fmt.Fprintf(s.bw, format, args...)
	s.bw.Flush()
}

func (s *session) sendlinef(format string, args ...interface{}) {
	s.sendf(format+"\r\n", args...)
}

func (s *session) sendPOP3ErrorOrLinef(err error, format string, args ...interface{}) {
	if se, ok := err.(POP3Error); ok {
		s.sendlinef("%s", se.Error())
		return
	}
	s.sendlinef(format, args...)
}

func (s *session) Addr() net.Addr {
	return s.rwc.RemoteAddr()
}

// parse commands to server

func (s *session) serve() {
	defer s.rwc.Close()
	if onc := s.srv.OnNewConnection; onc != nil {
		if err := onc(s); err != nil {
			s.sendPOP3ErrorOrLinef(err, "-ERR connection rejected")
			return
		}
	}
	s.clearAuthData()
	s.authApopLogin = utils.GenerateProtocolCramMd5(s.srv.hostname())
	s.sendf("+OK POP3 server ready %s\r\n", s.authApopLogin)
	for {
		if s.srv.ReadTimeout != 0 {
			s.rwc.SetReadDeadline(time.Now().Add(s.srv.ReadTimeout))
		}
		sl, err := s.br.ReadString('\n')
		if err != nil {
			// client close connection
			if io.EOF != err {
				s.errorf("read error: %v", err)
			}
			return
		}
		line := cmdLine(string(sl))
		if err := line.checkValid(); err != nil {
			s.sendlinef("-ERR %v", err)
			continue
		}

		log.Debugf("Command from client %s", line)

		switch line.Verb() {
		case "USER":
			s.handleLoginUser(line.Arg())
		case "PASS":
			s.handleLoginPass(line.Arg())
		case "CAPA":
			s.handleCapa()
		case "STAT":
			s.handleStat()
		case "RSET":
			s.handleRset()
		case "LIST":
			s.handleList(line.Arg())
		case "RETR", "UIDL":
			s.handleRetr(line.Arg())
		case "TOP":
			s.handleTop(line.Arg())
		case "DELE":
			s.handleDel(line.Arg())
		case "QUIT":
			s.sendlinef("+OK Bye")
			return
		case "AUTH":
			s.handleAuth(line.Arg())
		case "APOP":
			s.handleApop(line.Arg())
		case "NOOP":
			s.sendlinef("+OK")
		case "STLS":
			s.handleStartTLS()
		case "XTND":
			s.sendlinef("-ERR command not supported")
		default:
			if s.checkSeveralSteps(line) {
				log.Debugf("Client: %q, verhb: %q", line, line.Verb())
				s.sendlinef("-ERR command not recognized")
			}
		}

	}
}

// handle CAPA

func (s *session) handleCapa() {
	s.sendlinef("+OK Capability list follows")
	s.sendlinef("TOP")
	s.sendlinef("UIDL")
	s.sendlinef("SASL LOGIN PLAIN CRAM-MD5")
	if s.srv.ServerConfig.Pop3.Tls {
		s.sendlinef("STLS")
	}
	s.sendlinef(".")
}

// handle STAT

func (s *session) handleStat() {
	if !s.checkNeedAuth() {
		count, sum, err := s.srv.ServerConfig.DbPool.Pop3MessagesCountAndSum(s.mailboxId)
		if err != nil {
			s.sendlinef("-ERR unable to lock maildrop")
		} else {
			s.sendlinef("+OK %d %d", count, sum)
		}
	}
}

// handle RSET

func (s *session) handleRset() {
	if !s.checkNeedAuth() {
		count, sum, err := s.srv.ServerConfig.DbPool.Pop3MessagesCountAndSum(s.mailboxId)
		if err != nil {
			s.sendlinef("-ERR unable to lock maildrop")
		} else {
			s.sendlinef("+OK maildrop has %d messages (%d octets)", count, sum)
		}
	}
}

// handle LIST

func (s *session) handleList(line string) {
	if !s.checkNeedAuth() {
		count, sum, err := s.srv.ServerConfig.DbPool.Pop3MessagesCountAndSum(s.mailboxId)
		if err != nil {
			s.sendlinef("-ERR unable to lock maildrop")
		} else {
			s.sendlinef("+OK %d messages (%d octets)", count, sum)
			if count > 0 {
				s.cacheMessagesList()
				messageId := s.parseMessageId(line)
				if messageId > 0 && len(s.cachedList) >= messageId {
					s.sendlinef("%d %d", messageId, s.cachedList[messageId-1][1])
				} else {
					for i, msg := range s.cachedList {
						s.sendlinef("%d %d", i+1, msg[1])
					}
				}
			}
			s.sendlinef(".")
		}
	}
}

// handle RETR

func (s *session) handleRetr(line string) {
	if !s.checkNeedAuth() {
		messageId := s.getMessageId(s.parseMessageId(line))
		if messageId > 0 {
			msgSize, msgBody, err := s.srv.ServerConfig.DbPool.Pop3Message(s.mailboxId, messageId)
			if err != nil {
				s.sendlinef("-ERR no such message")
			} else {
				s.sendlinef("+OK %d octets", msgSize)
				s.sendlinef("%s", msgBody)
				s.sendlinef(".")
			}
		} else {
			s.sendlinef("-ERR no such message")
		}
	}
}

// handle DELE

func (s *session) handleDel(line string) {
	if !s.checkNeedAuth() {
		messageId := s.getMessageId(s.parseMessageId(line))
		if messageId > 0 {
			err := s.srv.ServerConfig.DbPool.Pop3DeleteMessage(s.mailboxId, messageId)
			if err != nil {
				s.sendlinef("-ERR no such message")
			} else {
				s.sendlinef("+OK message 1 deleted")
			}
		} else {
			s.sendlinef("-ERR no such message")
		}
	}
}

// handle TOP

func (s *session) handleTop(line string) {
	if !s.checkNeedAuth() {
		var msgId string
		if idx := strings.Index(line, " "); idx != -1 {
			msgId = strings.TrimSpace(line[:idx])
		} else {
			msgId = strings.TrimSpace(line)
		}

		messageId := s.getMessageId(s.parseMessageId(msgId))
		if messageId > 0 {
			msgSize, msgBody, err := s.srv.ServerConfig.DbPool.Pop3Message(s.mailboxId, messageId)
			if err != nil {
				s.sendlinef("-ERR no such message")
			} else {
				s.sendlinef("+OK %d octets", msgSize)
				s.sendlinef("%s", msgBody)
				s.sendlinef(".")
			}
		} else {
			s.sendlinef("-ERR no such message")
		}
	}
}

// handle AUTH

func (s *session) handleAuth(auth string) {
	var command, authToken string
	if idx := strings.Index(auth, " "); idx != -1 {
		command = strings.ToUpper(auth[:idx])
		authToken = strings.TrimRightFunc(auth[idx+1:len(auth)], unicode.IsSpace)
	} else {
		command = strings.ToUpper(auth)
		authToken = ""
	}
	switch command {
	case "PLAIN":
		s.tryPlainAuth(authToken)
	case "LOGIN":
		s.tryLoginAuth()
	case "CRAM-MD5":
		s.tryCramMd5Auth()
	default:
		s.sendlinef("-ERR Unrecognized authentication type")
	}
}

// check several step command

func (s *session) checkSeveralSteps(line cmdLine) bool {
	if s.authPlain {
		s.plainAuth(string(line))
		return false
	}
	if s.authLogin {
		s.loginAuth(string(line))
		return false
	}
	if s.authCramMd5Login != "" {
		s.cramMd5Auth(string(line))
		return false
	}
	return true
}

// plain auth

func (s *session) plainAuth(line string) {
	_, s.authUsername, s.authPassword = utils.DecodeProtocolAuthPlain(line)
	if s.authUsername != "" && s.authPassword != "" {
		s.authByDB(utils.AUTH_PLAIN)
	} else {
		s.sendlinef("-ERR authentication failed")
	}
	s.clearAuthData()
}

// check if plain auth 2 step or one

func (s *session) tryPlainAuth(authToken string) {
	if strings.Trim(authToken, " ") != "" {
		s.plainAuth(authToken)
	} else {
		s.clearAuthData()
		s.authPlain = true
		s.sendlinef("+ ")
	}
}

// login auth

func (s *session) loginAuth(line string) {
	if s.authUsername == "" {
		s.authUsername = utils.DecodeBase64(line)
		if s.authUsername != "" {
			s.sendlinef("+ ")
		} else {
			s.clearAuthData()
			s.sendlinef("-ERR authentication failed")
		}
		return
	}
	if s.authPassword == "" {
		s.authPassword = utils.DecodeBase64(line)
		if s.authPassword != "" {
			s.authByDB(utils.AUTH_PLAIN)
		} else {
			s.sendlinef("-ERR authentication failed")
		}
		s.clearAuthData()
	}
}

// check if login auth 2 step

func (s *session) tryLoginAuth() {
	s.clearAuthData()
	s.authLogin = true
	s.sendlinef("+ ")
}

// check cram md5 login

func (s *session) cramMd5Auth(line string) {
	s.authUsername, s.authPassword = utils.DecodeProtocolCramMd5(line)
	if s.authUsername != "" && s.authPassword != "" {
		s.authByDB(utils.AUTH_CRAM_MD5)
	} else {
		s.sendlinef("-ERR authentication failed")
	}
	s.clearAuthData()
}

// check try cram-md5 login

func (s *session) tryCramMd5Auth() {
	s.clearAuthData()
	s.authCramMd5Login = utils.GenerateProtocolCramMd5(s.srv.hostname())
	s.sendlinef("+ " + utils.EncodeBase64(s.authCramMd5Login))
}

// handle login user

func (s *session) handleLoginUser(line string) {
	s.authUsername = line
	if s.srv.ServerConfig.DbPool.IfUserExist(s.authUsername) {
		s.sendlinef("+OK %s is a valid mailbox", s.authUsername)
	} else {
		s.sendlinef("-ERR never heard of mailbox name")
	}
}

// handle pass user

func (s *session) handleLoginPass(line string) {
	s.authPassword = line
	if s.authUsername != "" && s.authPassword != "" {
		s.authByDB(utils.AUTH_PLAIN)
	} else {
		s.sendlinef("-ERR invalid username or password")
	}
	s.clearAuthData()
}

// auth by DB

func (s *session) authByDB(authMethod string) {
	var err error
	s.mailboxId, err = s.srv.ServerConfig.DbPool.CheckUser(authMethod, s.authUsername, s.authPassword, s.authCramMd5Login)
	if err != nil {
		s.sendlinef("-ERR invalid username or password")
		return
	}
	s.sendlinef("+OK maildrop locked and ready")
}

// clear auth

func (s *session) clearAuthData() {
	s.authPlain = false
	s.authLogin = false
	s.authCramMd5Login = ""
	s.authUsername = ""
	s.authPassword = ""
}

// handle apop

func (s *session) handleApop(line string) {
	s.clearAuthData()
	if idx := strings.Index(line, " "); idx != -1 {
		s.authUsername = line[:idx]
		s.authPassword = strings.TrimRightFunc(line[idx+1:len(line)], unicode.IsSpace)
	} else {
		s.authUsername = line
		s.authPassword = ""
	}
	if s.authUsername != "" && s.authPassword != "" {
		s.authCramMd5Login = s.authApopLogin
		s.authByDB(utils.AUTH_APOP)
	} else {
		s.sendlinef("-ERR invalid username or password")
	}
	s.clearAuthData()
}

// handle StartTLS

func (s *session) handleStartTLS() {
	if s.srv.ServerConfig.Pop3.Tls {
		s.sendlinef("+OK Begin TLS negotiation")
		var tlsConn *tls.Conn
		tlsConn = tls.Server(s.rwc, s.srv.TLSconfig)
		err := tlsConn.Handshake()
		if err != nil {
			log.Errorf("Could not TLS handshake:%v", err)
		} else {
			s.rwc = net.Conn(tlsConn)
			s.br = bufio.NewReader(s.rwc)
			s.bw = bufio.NewWriter(s.rwc)
		}
		s.sendlinef("")
	} else {
		s.sendlinef("-ERR Tsl not supported")
	}
}

// check auth if need

func (s *session) checkNeedAuth() bool {
	if s.mailboxId == 0 {
		s.sendlinef("-ERR permission denied")
		return true
	}
	return false
}

// Handle error

func (s *session) handleError(err error) {
	if se, ok := err.(POP3Error); ok {
		s.sendlinef("%s", se)
		return
	}
	log.Errorf("Error: %s", err)
}

// Utils

func (s *session) parseMessageId(line string) int {
	messageId := strings.TrimSpace(line)
	if messageId != "" {
		messageId, err := strconv.Atoi(messageId)
		if err == nil {
			return messageId
		}
	}
	return 0
}

func (s *session) getMessageId(messageId int) int {
	s.cacheMessagesList()
	if len(s.cachedList) > 0 && messageId > 0 && len(s.cachedList) >= messageId {
		return s.cachedList[messageId-1][0]
	}
	return 0
}

func (s *session) cacheMessagesList() {
	if s.isListFetched == false && len(s.cachedList) == 0 {
		cachedList, errList := s.srv.ServerConfig.DbPool.Pop3MessagesList(s.mailboxId)
		if errList == nil {
			s.cachedList = cachedList
		}
		s.isListFetched = true
	}
}

// COMMAND LINE

type cmdLine string

func (cl cmdLine) checkValid() error {
	if !strings.HasSuffix(string(cl), "\r\n") {
		return errors.New(`line doesn't end in \r\n`)
	}
	return nil
}

func (cl cmdLine) Verb() string {
	s := string(cl)
	if idx := strings.Index(s, " "); idx != -1 {
		return strings.ToUpper(s[:idx])
	}
	return strings.ToUpper(s[:len(s)-2])
}

func (cl cmdLine) Arg() string {
	s := string(cl)
	if idx := strings.Index(s, " "); idx != -1 {
		return strings.TrimRightFunc(s[idx+1:len(s)-2], unicode.IsSpace)
	}
	return ""
}

func (cl cmdLine) String() string {
	return string(cl)
}

// ERRORS

type POP3Error string

func (e POP3Error) Error() string {
	return string(e)
}

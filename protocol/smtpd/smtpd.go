// Package smtpd implements an SMTP server. Hooks are provided to customize
// its behavior.
package smtpd

// TODO:
//  -- send 421 to connected clients on graceful server shutdown (s3.8)
//

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/le0pard/go-falcon/config"
	"github.com/le0pard/go-falcon/log"
	"github.com/le0pard/go-falcon/utils"
	"io"
	"io/ioutil"
	"net"
	"net/textproto"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	rcptToRE   = regexp.MustCompile(`[Tt][Oo]:[\s*]?<(.+)>`)
	mailFromRE = regexp.MustCompile(`[Ff][Rr][Oo][Mm]:[\s*]?<(.*)>`)
	nginxRE    = regexp.MustCompile(`(?i)(.*) LOGIN=(\d+) (.*)`)
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

	// OnNewMail must be defined and is called when a new message beings.
	// (when a MAIL FROM line arrives)
	OnNewMail func(c Connection, from MailAddress) (Envelope, error)
}

// MailAddress is defined by
type MailAddress interface {
	Email() string    // email address, as provided
	Hostname() string // canonical hostname, lowercase
	Username() string // clear username, without "+" part, lowercase
}

// Connection is implemented by the SMTP library and provided to callers
// customizing their own Servers.
type Connection interface {
	Addr() net.Addr
}

// EMAIL

type Envelope interface {
	AddMailboxId(mailboxId int) error
	AddSender(from MailAddress) error
	AddRecipient(rcpt MailAddress) error
	BeginData() error
	Write(line []byte) error
	Close() error
}

type BasicEnvelope struct {
	MailboxID int
	From      MailAddress
	Rcpts     []MailAddress
	MailBody  []byte
}

func (e *BasicEnvelope) AddMailboxId(mailboxId int) error {
	e.MailboxID = mailboxId
	return nil
}

func (e *BasicEnvelope) AddSender(from MailAddress) error {
	e.From = from
	return nil
}

func (e *BasicEnvelope) AddRecipient(rcpt MailAddress) error {
	e.Rcpts = append(e.Rcpts, rcpt)
	return nil
}

func (e *BasicEnvelope) BeginData() error {
	if len(e.Rcpts) == 0 {
		return SMTPError("554 5.5.1 Error: no valid recipients")
	}
	if e.MailboxID <= 0 {
		return SMTPError("554 5.5.1 Error: no inbox for this email")
	}
	return nil
}

func (e *BasicEnvelope) Write(line []byte) error {
	e.MailBody = line
	return nil
}

func (e *BasicEnvelope) Close() error {
	return nil
}

// SERVER

func (srv *Server) hostname() string {
	if srv.Hostname != "" {
		return srv.Hostname
	}
	if srv.ServerConfig.Adapter.Ssl_Hostname != "" {
		return srv.ServerConfig.Adapter.Ssl_Hostname
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
		addr = ":2525"
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
				log.Errorf("smtpd: Accept error: %v", e)
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

	env Envelope // current envelope, or nil

	helloType string
	helloHost string

	authPlain        bool   // bool for 2 step plain auth
	authLogin        bool   // bool for 2 step login auth
	authCramMd5Login string // bytes for cram-md5 login

	mailboxId    int    // id of mailbox
	maxMessages  int    // max messages
	authUsername string // auth login
	authPassword string // auth password

	rateLimit int  // rate limit from db
	isBlocked bool // is session blocked
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
		rateLimit:        srv.ServerConfig.Adapter.Rate_Limit,
		isBlocked:        false,
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

func (s *session) sendSMTPErrorOrLinef(err error, format string, args ...interface{}) {
	if se, ok := err.(SMTPError); ok {
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
			s.sendSMTPErrorOrLinef(err, "554 connection rejected")
			return
		}
	}
	s.sendf("220 %s %s\r\n", s.srv.ServerConfig.Adapter.Welcome_Msg, s.srv.hostname())
	for {
		if s.srv.ReadTimeout != 0 {
			s.rwc.SetReadDeadline(time.Now().Add(s.srv.ReadTimeout))
		}
		sl, err := s.br.ReadString('\n')
		if err != nil {
			// client close connection
			if io.EOF != err {
				s.errorf("read error: %v", err)
				s.resetEnvelope()
			}
			return
		}
		line := cmdLine(sl)
		if err := line.checkValid(); err != nil {
			s.sendlinef("500 %v", err)
			continue
		}

		log.Debugf("Command from client %s", line)

		switch line.Verb() {
		case "HELO", "EHLO", "LHLO":
			s.handleHello(line.Verb(), line.Arg())
		case "QUIT":
			s.sendlinef("221 2.0.0 Bye")
			return
		case "RSET":
			s.resetEnvelope()
			s.sendlinef("250 2.0.0 OK")
		case "NOOP":
			s.sendlinef("250 2.0.0 OK")
		case "MAIL":
			arg := line.Arg() // "From:<foo@bar.com>"
			m := mailFromRE.FindStringSubmatch(arg)
			if m == nil {
				log.Errorf("invalid MAIL arg: %q", arg)
				s.sendlinef("501 5.1.7 Bad sender address syntax")
				continue
			}
			s.handleMailFrom(m[1])
		case "RCPT":
			s.handleRcpt(line)
		case "DATA":
			s.handleData()
		case "VRFY", "EXPN":
			s.sendlinef("252 send some mail, i'll try my best")
		case "HELP":
			s.sendlinef("214-This server supports the following commands:")
			s.sendlinef("214 HELO EHLO STARTTLS RCPT DATA RSET MAIL QUIT HELP AUTH VRFY NOOP")
		case "XCLIENT":
			// Nginx sends this
			s.handleNginx(line.Arg())
		case "AUTH":
			s.handleAuth(line.Arg())
		case "STARTTLS":
			s.handleStartTLS()
		default:
			if s.checkSeveralSteps(line) {
				log.Debugf("Client: %q, verhb: %q", line, line.Verb())
				s.sendlinef("502 5.5.2 Error: command not recognized")
			}
		}

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

// Handle HELO, EHLO msg

func (s *session) handleHello(greeting, host string) {
	s.helloType = greeting
	s.helloHost = host
	fmt.Fprintf(s.bw, "250-%s\r\n", s.srv.hostname())
	extensions := []string{}
	if s.srv.ServerConfig.Adapter.Auth {
		extensions = append(extensions, "250-AUTH LOGIN PLAIN CRAM-MD5")
	}
	if s.srv.ServerConfig.Adapter.Tls {
		extensions = append(extensions, "250-STARTTLS")
	}
	// size end
	extensions = append(extensions,
		"250-DSN",
		"250-PIPELINING",
		fmt.Sprintf("250-SIZE %d", s.srv.ServerConfig.Adapter.Max_Mail_Size),
		"250-ENHANCEDSTATUSCODES",
		"250-8BITMIME",
		"250 HELP",
	)
	for _, ext := range extensions {
		fmt.Fprintf(s.bw, "%s\r\n", ext)
	}
	s.bw.Flush()
}

// Handle mail from

func (s *session) handleMailFrom(email string) {
	// TODO: 4.1.1.11.  If the server SMTP does not recognize or
	// cannot implement one or more of the parameters associated
	// qwith a particular MAIL FROM or RCPT TO command, it will return
	// code 555.

	if s.env != nil {
		s.sendlinef("503 5.5.1 Error: nested MAIL command")
		return
	}
	log.Debugf("mail from: %q", email)
	cb := s.srv.OnNewMail
	if cb == nil {
		log.Errorf("smtp: Server.OnNewMail is nil; rejecting MAIL FROM")
		s.sendf("451 Server.OnNewMail not configured\r\n")
		return
	}
	s.resetEnvelope()
	fromEmail := addrString(email)
	env, err := cb(s, fromEmail)
	if err != nil {
		log.Errorf("rejecting MAIL FROM %q: %v", email, err)
		// TODO: send it back to client if warranted, like above
		return
	}
	s.env = env
	s.env.AddSender(fromEmail)
	s.sendlinef("250 2.1.0 Ok")
}

// Handle to in mail

func (s *session) handleRcpt(line cmdLine) {
	// TODO: 4.1.1.11.  If the server SMTP does not recognize or
	// cannot implement one or more of the parameters associated
	// qwith a particular MAIL FROM or RCPT TO command, it will return
	// code 555.

	if s.env == nil {
		s.sendlinef("503 5.5.1 Error: need MAIL command")
		return
	}
	if s.checkNeedAuthOrBlocked() {
		return
	}

	arg := line.Arg() // "To:<foo@bar.com>"
	m := rcptToRE.FindStringSubmatch(arg)
	if m == nil {
		log.Errorf("bad RCPT address: %q", arg)
		s.sendlinef("501 5.1.7 Bad sender address syntax")
		return
	}

	rcptEmail := addrString(m[1])
	if s.srv.ServerConfig.Email_Address_Mode.Enabled {
		s.handleToAddressMode(rcptEmail)
	}
	err := s.env.AddRecipient(rcptEmail)
	if err != nil {
		s.sendSMTPErrorOrLinef(err, "550 bad recipient")
		return
	}
	s.sendlinef("250 2.1.0 Ok")
}

// Handle nginx

func (s *session) handleNginx(line string) {
	if s.srv.ServerConfig.Proxy.Enabled || s.srv.ServerConfig.Proxy.Proxy_Mode {
		if s.srv.ServerConfig.Adapter.Auth {
			if nginxRE.MatchString(line) {
				res := nginxRE.FindStringSubmatch(line)
				if len(res) == 4 {
					mailboxId, err := strconv.Atoi(res[2])
					if err == nil {
						s.setMailboxIdHook(mailboxId)
						s.sendlinef("250 2.0.0 OK")
						return
					}
				}
			}
		} else {
			s.sendlinef("250 2.0.0 OK")
			return
		}
	}
	s.sendlinef("535 5.7.1 authentication failed")
}

// Handle data

func (s *session) handleData() {
	if s.env == nil {
		s.sendlinef("503 5.5.1 Error: need RCPT command")
		return
	}
	// rate limit
	s.isBlocked = s.redisIsSessionBlocked()
	// is need to block?
	if s.checkNeedAuthOrBlocked() {
		return
	} else {
		// store mailbox id in envelop
		if s.mailboxId > 0 {
			s.env.AddMailboxId(s.mailboxId)
		}
	}

	if err := s.env.BeginData(); err != nil {
		s.handleError(err)
		return
	}

	s.sendlinef("354 Go ahead")

	data := &bytes.Buffer{}
	reader := textproto.NewReader(s.br).DotReader()
	_, err := io.CopyN(data, reader, int64(s.srv.ServerConfig.Adapter.Max_Mail_Size))

	if err == io.EOF {
		s.env.Write(data.Bytes())
		s.env.Close()
		s.resetEnvelope()
		s.sendlinef("250 2.0.0 Ok: queued")
	}

	if err != nil {
		// Network error, ignore (or just exit)
		return
	}

	log.Errorf("smtpd: Too big message for: %v", s.mailboxId)

	// Discard the rest and report an error.
	_, err = io.Copy(ioutil.Discard, reader)

	if err != nil {
		// reset envelope
		s.resetEnvelope()
		// Network error, ignore
		return
	}

	s.sendlinef(fmt.Sprintf("552 5.7.0 Message exceeded max message size of %d bytes", s.srv.ServerConfig.Adapter.Max_Mail_Size))
	s.resetEnvelope()
}

func (s *session) resetEnvelope() {
	s.env = nil
}

// check auth if need and not blocked

func (s *session) checkNeedAuthOrBlocked() bool {
	if s.srv.ServerConfig.Adapter.Auth && 0 == s.mailboxId {
		s.sendlinef("530 5.7.0 Authentication required")
		return true
	}
	if s.isBlocked {
		s.sendlinef("550 5.7.0 Requested action not taken: too many emails per second")
		return true
	}
	return false
}

// auth by DB

func (s *session) authByDB(authMethod string) {
	if s.srv.ServerConfig.Adapter.Auth {
		mailboxId, err := s.srv.ServerConfig.DbPool.CheckUser(authMethod, s.authUsername, s.authPassword, s.authCramMd5Login)
		if err != nil {
			s.sendlinef("535 5.7.1 authentication failed")
			return
		}
		s.setMailboxIdHook(mailboxId)
	}
	s.sendlinef("235 2.0.0 OK, go ahead")
}

// sucess set mailbox id

func (s *session) setMailboxIdHook(mailboxId int) {
	s.mailboxId = mailboxId
	// get rate limit
	if rateLimit, err := s.getInboxRateLimit(s.mailboxId); err == nil {
		s.rateLimit = rateLimit
	}
}

// plain auth

func (s *session) plainAuth(line string) {
	_, s.authUsername, s.authPassword = utils.DecodeProtocolAuthPlain(line)
	if s.authUsername != "" && s.authPassword != "" {
		s.authByDB(utils.AUTH_PLAIN)
	} else {
		s.sendlinef("535 5.7.1 authentication failed")
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
		s.sendlinef("334 2.0.0 OK")
	}
}

// login auth

func (s *session) loginAuth(line string) {
	if s.authUsername == "" {
		s.authUsername = utils.DecodeBase64(line)
		if s.authUsername != "" {
			s.sendlinef("334 UGFzc3dvcmQ6")
		} else {
			s.clearAuthData()
			s.sendlinef("535 5.7.1 authentication failed")
		}
		return
	}
	if s.authPassword == "" {
		s.authPassword = utils.DecodeBase64(line)
		if s.authPassword != "" {
			s.authByDB(utils.AUTH_PLAIN)
		} else {
			s.sendlinef("535 5.7.1 authentication failed")
		}
		s.clearAuthData()
	}
}

// check if login auth 2 step

func (s *session) tryLoginAuth() {
	s.clearAuthData()
	s.authLogin = true
	s.sendlinef("334 VXNlcm5hbWU6")
}

// check cram md5 login

func (s *session) cramMd5Auth(line string) {
	s.authUsername, s.authPassword = utils.DecodeProtocolCramMd5(line)
	if s.authUsername != "" && s.authPassword != "" {
		s.authByDB(utils.AUTH_CRAM_MD5)
	} else {
		s.sendlinef("535 5.7.1 authentication failed")
	}
	s.clearAuthData()
}

// check try cram-md5 login

func (s *session) tryCramMd5Auth() {
	s.clearAuthData()
	s.authCramMd5Login = utils.GenerateProtocolCramMd5(s.srv.hostname())
	s.sendlinef(fmt.Sprintf("334 %s", utils.EncodeBase64(s.authCramMd5Login)))
}

// clear auth

func (s *session) clearAuthData() {
	s.authPlain = false
	s.authLogin = false
	s.authCramMd5Login = ""
	s.authUsername = ""
	s.authPassword = ""
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
		s.sendlinef("504 5.5.2 Unrecognized authentication type")
	}
}

// handle StartTLS

func (s *session) handleStartTLS() {
	if s.srv.ServerConfig.Adapter.Tls {
		s.sendlinef("220 2.0.0 Ready to start TLS")
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
		s.sendlinef("503 5.5.1 Error: Tsl not supported")
	}
}

// Handle TO address for auth by address

func (s *session) handleToAddressMode(rcptEmail MailAddress) {
	username := rcptEmail.Username()
	hostname := rcptEmail.Hostname()
	if len(hostname) > 0 && posInSlice(s.srv.ServerConfig.Email_Address_Mode.Domains, hostname) > -1 && len(username) > 0 {
		mailboxId, err := s.srv.ServerConfig.DbPool.CheckAddressMode(username)
		if err == nil && mailboxId > 0 {
			s.setMailboxIdHook(mailboxId)
		}
	}
}

func posInSlice(slice []string, value string) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}

// Handle error

func (s *session) handleError(err error) {
	if se, ok := err.(SMTPError); ok {
		s.sendlinef("%s", se)
		return
	}
	log.Errorf("Error: %s", err)
	s.resetEnvelope()
}

// ADDRESS

type addrString string

func (a addrString) Email() string {
	return string(a)
}

func (a addrString) Hostname() string {
	e := string(a)
	if idx := strings.Index(e, "@"); idx != -1 {
		return strings.ToLower(e[idx+1:])
	}
	return ""
}

func (a addrString) Username() string {
	e := string(a)
	if idx := strings.Index(e, "@"); idx != -1 {
		username := strings.ToLower(e[0:idx])
		if sidx := strings.Index(username, "+"); sidx != -1 {
			return strings.ToLower(username[0:sidx])
		} else {
			return strings.ToLower(username)
		}
	}
	return ""
}

// COMMAND LINE

type cmdLine string

func (cl cmdLine) checkValid() error {
	if !strings.HasSuffix(string(cl), "\r\n") {
		return errors.New(`line doesn't end in \r\n`)
	}
	// Check for verbs defined not to have an argument
	// (RFC 5321 s4.1.1)
	switch cl.Verb() {
	case "RSET", "DATA", "QUIT":
		if cl.Arg() != "" {
			return errors.New("unexpected argument")
		}
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

type SMTPError string

func (e SMTPError) Error() string {
	return string(e)
}

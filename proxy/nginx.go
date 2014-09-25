package proxy

import (
	"fmt"
	"github.com/le0pard/go-falcon/config"
	"github.com/le0pard/go-falcon/log"
	"github.com/le0pard/go-falcon/utils"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

const (
	MAX_AUTH_RETRY         = 10
	INVALID_AUTH_WAIT_TIME = "3"
	PROTOCOL_SMTP          = "smtp"
	PROTOCOL_POP3          = "pop3"
)

var (
	roundRobinIterator = 0
	roundRobinMutex    = &sync.Mutex{}
)

// If running Nginx as a proxy, give Nginx the IP address and port for the SMTP server
// Primary use of Nginx is to terminate TLS so that Go doesn't need to deal with it.
// This could perform auth and load balancing too
// See http://wiki.nginx.org/MailCoreModule
func StartNginxHTTPProxy(config *config.Config) {
	if config.Proxy.Enabled {
		go nginxHTTPAuth(config)
	}
}

// nginx auth server

func nginxHTTPAuth(config *config.Config) {
	// handle
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		nginxHTTPAuthHandler(w, r, config)
	})
	// server ip:port
	serverBind := fmt.Sprintf("%s:%d", config.Proxy.Host, config.Proxy.Port)
	//
	log.Debugf("Nginx proxy working on %s", serverBind)
	//
	errServer := http.ListenAndServe(serverBind, nil)
	if errServer != nil {
		log.Errorf("Nginx proxy: %v", errServer)
	}
}

// nginx auth by nginx headers

func nginxHTTPAuthHandler(w http.ResponseWriter, r *http.Request, config *config.Config) {
	log.Debugf("Nginx proxy get request: %v", r)

	protocol := strings.ToLower(r.Header.Get("Auth-Protocol"))

	if protocol == PROTOCOL_SMTP || protocol == PROTOCOL_POP3 {
		authMethod := strings.ToLower(r.Header.Get("Auth-Method"))
		username := r.Header.Get("Auth-User")
		password := r.Header.Get("Auth-Pass")
		secret := ""
		if authMethod == utils.AUTH_CRAM_MD5 || authMethod == utils.AUTH_APOP {
			secret = r.Header.Get("Auth-Salt")
		}
		id, pass, err := config.DbPool.CheckUserWithPass(authMethod, username, password, secret)
		if err != nil {
			nginxResponseFail(w, r)
			return
		}
		nginxResponseSuccess(config, w, protocol, strconv.Itoa(id), pass)
	} else {
		nginxResponseFail(w, r)
	}
}

// success auth response

func nginxResponseSuccess(config *config.Config, w http.ResponseWriter, protocol, userId, password string) {
	serverHost, serverPort := config.Adapter.Host, strconv.Itoa(getRoundRobinFromArray(config.SmtpPortRanges))

	if protocol == PROTOCOL_SMTP {
		w.Header().Add("Auth-User", userId) // return mailbox id instead username
	} else if protocol == PROTOCOL_POP3 {
		serverHost, serverPort = config.Pop3.Host, strconv.Itoa(utils.GetRandFromArray(config.Pop3PortRanges)) // revrite server options
		w.Header().Add("Auth-Pass", password)                                                                  // return password for pop3
	}
	w.Header().Add("Auth-Status", "OK")
	w.Header().Add("Auth-Server", serverHost)
	w.Header().Add("Auth-Port", serverPort)
	// empty body
	fmt.Fprint(w, "")
}

// fail auth response

func nginxResponseFail(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Auth-Status", "Invalid login or password")
	// login attempt
	loginAttempt := r.Header.Get("Auth-Login-Attempt")
	if loginAttempt != "" {
		loginAttemptInt, err := strconv.Atoi(loginAttempt)
		if err == nil {
			if loginAttemptInt < MAX_AUTH_RETRY {
				w.Header().Add("Auth-Wait", INVALID_AUTH_WAIT_TIME)
			}
		}
	}
	// empty body
	fmt.Fprint(w, "")
}

// round robin

func getRoundRobinFromArray(arr []int) int {
	roundRobinMutex.Lock()
	defer roundRobinMutex.Unlock()

	roundRobinIterator = roundRobinIterator + 1
	if roundRobinIterator >= len(arr) {
		roundRobinIterator = 0
	}

	return arr[roundRobinIterator]
}

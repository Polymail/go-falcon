package config

import (
  "runtime"
  "io/ioutil"
  "time"
  "fmt"
  "launchpad.net/goyaml"
  "github.com/garyburd/redigo/redis"
  "github.com/le0pard/go-falcon/log"
)

// Config represents the supported configuration options for a falcon,
// as declared in its config.yml file.
type protocolType string

const (
    protocolSmtp protocolType = "smtp"
    protocolLmtp protocolType = "lmtp"
)

type Config struct {
  Adapter struct {
    Protocol        protocolType
    Host            string
    Port            int
    Hostname        string
    Auth            bool
    Tls             bool
    Ssl_Hostname    string
    Ssl_Pub_Key     string
    Ssl_Prv_Key     string
    Welcome_Msg     string
    Max_Mail_Size   int
  }
  Storage struct {
    Adapter                   string
    Host                      string
    Port                      int
    Username                  string
    Password                  string
    Database                  string
    Pool                      int

    Auth_Sql                  string

    Settings_Sql              string

    Messages_Sql              string
    Attachments_Sql           string

    Max_Messages_Enabled      bool
    Max_Messages_Sql          string
    Max_Messages_Cleanup_Sql  string
    Max_Attachments_Cleanup_Sql string

    Spamassassin_Sql          string

    Clamav_Sql                string

    Pop3_Count_And_Size_Messages  string
    Pop3_Messages_List            string
    Pop3_Messages_List_One        string
    Pop3_Message_One              string
    Pop3_Message_Delete           string
  }
  Pop3 struct {
    Enabled         bool
    Host            string
    Port            int
    Hostname        string
    Tls             bool
    Ssl_Hostname    string
    Ssl_Pub_Key     string
    Ssl_Prv_Key     string
  }
  Spamassassin struct {
    Enabled       bool
    Ip            string
    Port          int
    Timeout       int
  }
  Clamav struct {
    Enabled       bool
    Host          string
    Port          int
    Timeout       int
  }
  Proxy struct {
    Enabled       bool
    Host          string
    Port          int
  }
  Redis struct {
    Enabled       bool
    Host          string
    Port          int
    Pool          int
    Timeout       int
    Namespace     string
    Hook_Username string
    Hook_Password string
    Sidekiq_Queue string
    Sidekiq_Class string
  }
  Web_Hooks struct {
    Enabled       bool
    Username      string
    Password      string
    Urls          []string
  }
  Daemon struct {
    Max_Procs     int
  }
  Log struct {
    Debug         bool
  }
  RedisPool       *redis.Pool
}

// NewConfig returns a new Config without any options.
func NewConfig() *Config {
  return &Config{}
}

// ReadEnvirons reads the juju config.yml file
// and returns the result of running Config
// on the file's contents.
func ReadConfig(filename string) (*Config, error) {
  data, err := ioutil.ReadFile(filename)
  if err != nil {
    log.Errorf("cannot read file %q: %v", filename, err)
    return nil, err
  }
  e, err := readConfigBytes(data)
  if err != nil {
    log.Errorf("cannot parse file %q: %v", filename, err)
    return nil, err
  }
  e.setDefaultValues()
  if e.Redis.Enabled {
    e.initRedisPool()
  }
  return e, nil
}

// setDefaultValues for yaml config
func (config *Config) setDefaultValues() {
  // default for Adapter
  if config.Adapter.Protocol != protocolSmtp && config.Adapter.Protocol != protocolLmtp {
    config.Adapter.Protocol = protocolSmtp
  }
  if config.Adapter.Host == "" {
    config.Adapter.Host = "localhost"
  }
  if config.Adapter.Port <= 0 {
    config.Adapter.Port = 25
  }
  if config.Adapter.Welcome_Msg == "" {
    config.Adapter.Welcome_Msg = "Falcon Mail Server"
  }
  if config.Adapter.Max_Mail_Size <= 0 || config.Adapter.Max_Mail_Size > 99999999 {
    config.Adapter.Max_Mail_Size = 10240000
  }
  // default for Storage
  if config.Storage.Host == "" {
    config.Storage.Host = "localhost"
  }
  if config.Storage.Port <= 0 {
    config.Storage.Port = 5432
  }
  if config.Storage.Pool < 1 {
    config.Storage.Pool = 1
  }
  // default for Proxy
  if config.Proxy.Host == "" {
    config.Proxy.Host = "localhost"
  }
  if config.Proxy.Port <= 0 {
    config.Proxy.Port = 2525
  }
  if config.Daemon.Max_Procs <= 0 {
    config.Daemon.Max_Procs = runtime.NumCPU()
  }
}

func (config *Config) initRedisPool() {
  // pool
  config.RedisPool = &redis.Pool{
    MaxIdle: config.Redis.Pool,
    IdleTimeout: time.Duration(config.Redis.Timeout) * time.Second,
    Dial: func () (redis.Conn, error) {
      c, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port))
      if err != nil {
        return nil, err
      }
      /*
      if _, err := c.Do("AUTH", password); err != nil {
        c.Close()
        return nil, err
      }
      */
      return c, err
    },
    TestOnBorrow: func(c redis.Conn, t time.Time) error {
      _, err := c.Do("PING")
      return err
    },
  }
}

// readConfigBytes parses the contents of an config.yml file
// and returns its representation.
func readConfigBytes(data []byte) (*Config, error) {
  config := NewConfig()
  err := goyaml.Unmarshal(data, &config)
  if err != nil {
    return nil, err
  }
  return config, nil
}

package config

import (
  "fmt"
  "github.com/garyburd/redigo/redis"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/storage"
  "io/ioutil"
  "launchpad.net/goyaml"
  "time"
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
    Protocol      protocolType
    Host          string
    Port          int
    Hostname      string
    Auth          bool
    Tls           bool
    Ssl_Hostname  string
    Ssl_Pub_Key   string
    Ssl_Prv_Key   string
    Welcome_Msg   string
    Max_Mail_Size int
    Rate_Limit    int
    Workers_Size  int
  }
  Storage *storage.StorageConfig
  Pop3    struct {
    Enabled      bool
    Host         string
    Port         int
    Hostname     string
    Tls          bool
    Ssl_Hostname string
    Ssl_Pub_Key  string
    Ssl_Prv_Key  string
  }
  Spamassassin struct {
    Enabled bool
    Ip      string
    Port    int
    Timeout int
  }
  Clamav struct {
    Enabled bool
    Host    string
    Port    int
    Timeout int
  }
  Proxy struct {
    Enabled      bool
    Proxy_Mode   bool
    Host         string
    Port         int
    Client_Ports struct {
      Smtp []int
      Pop3 []int
    }
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
  Log struct {
    Debug bool
  }
  DbPool         *storage.DBConn
  RedisPool      *redis.Pool
  SmtpPortRanges []int
  Pop3PortRanges []int
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
  err = e.initDbPool()
  if err != nil {
    return nil, err
  }
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
  if config.Adapter.Rate_Limit <= 0 {
    config.Adapter.Rate_Limit = 2
  }
  if config.Adapter.Workers_Size <= 0 {
    config.Adapter.Workers_Size = 5
  }
  // default for Storage
  if config.Storage.Host == "" {
    config.Storage.Host = "localhost"
  }
  if config.Storage.Port <= 0 {
    config.Storage.Port = 5432
  }
  if config.Storage.Pool < 1 {
    config.Storage.Pool = 5
  }
  if config.Storage.Pool_Idle < 1 {
    config.Storage.Pool_Idle = 2
  }
  // default for Proxy
  if config.Proxy.Host == "" {
    config.Proxy.Host = "localhost"
  }
  if config.Proxy.Port <= 0 {
    config.Proxy.Port = 2525
  }
  // ports
  config.SmtpPortRanges = []int{config.Adapter.Port}
  if len(config.Proxy.Client_Ports.Smtp) > 0 {
    config.SmtpPortRanges = append(config.SmtpPortRanges, config.Proxy.Client_Ports.Smtp...)
  }
  config.Pop3PortRanges = []int{config.Pop3.Port}
  if len(config.Proxy.Client_Ports.Pop3) > 0 {
    config.Pop3PortRanges = append(config.Pop3PortRanges, config.Proxy.Client_Ports.Pop3...)
  }
}

func (config *Config) initDbPool() error {
  var (
    err error
  )
  config.DbPool, err = storage.InitDatabase(config.Storage)
  if err != nil {
    log.Errorf("Problem with connection to storage: %s", err)
    return err
  }
  return nil
}

func (config *Config) initRedisPool() {
  // pool
  config.RedisPool = &redis.Pool{
    MaxIdle:     config.Redis.Pool,
    IdleTimeout: time.Duration(config.Redis.Timeout) * time.Second,
    Dial: func() (redis.Conn, error) {
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

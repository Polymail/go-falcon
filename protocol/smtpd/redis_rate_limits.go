// Package smtpd implements an SMTP server. Hooks are provided to customize
// its behavior. Redis function add rate limit functionality.

package smtpd

import (
  "fmt"
  "github.com/garyburd/redigo/redis"
  "github.com/le0pard/go-falcon/log"
)


func (s *session) redisIsSessionBlocked() bool {
  if s.srv.ServerConfig.Redis.Enabled == false {
    return false
  }

  redisCon := s.srv.ServerConfig.RedisPool.Get()
  defer redisCon.Close()

  emailsCount, err := redis.Int(redisCon.Do("GET", s.redisRateLimitKey()))

  if err != nil || emailsCount <= s.srv.ServerConfig.Adapter.Rate_Limit {
    return false
  }

  return true
}

func (s *session) redisRateLimits() {
  if s.srv.ServerConfig.Redis.Enabled == false {
    return
  }

  redisCon := s.srv.ServerConfig.RedisPool.Get()
  defer redisCon.Close()

  redisKey := s.redisRateLimitKey()

  emailsKeyCount, err := redis.Int(redisCon.Do("INCR", redisKey))
  if err == nil {
    if emailsKeyCount == 1 {
      _, err = redisCon.Do("EXPIRE", redisKey, 1) // expire 1 sec
      if err != nil {
        log.Errorf("redisRateLimits EXPIRE error: %v", err)
      }
    }
  } else {
    log.Errorf("redisRateLimits INCR error: %v", err)
  }
  return
}


func (s *session) redisRateLimitKey() string {
  return fmt.Sprintf("%s:rate-inbox-limits-%d", s.srv.ServerConfig.Redis.Namespace, s.mailboxId);
}
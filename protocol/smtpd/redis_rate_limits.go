// Package smtpd implements an SMTP server. Hooks are provided to customize
// its behavior. Redis function add rate limit functionality.

package smtpd

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/le0pard/go-falcon/log"
	"github.com/le0pard/go-falcon/redisworker"
)

func (s *session) getInboxRateLimit(mailboxId int) (int, error) {
	inboxSettings, err := redisworker.GetCachedInboxSettings(s.srv.ServerConfig, mailboxId)
	if err != nil || 0 == inboxSettings.MaxMessages || 0 == inboxSettings.RateLimit {
		// inbox setting from database
		inboxSettings, err = s.srv.ServerConfig.DbPool.GeInboxSettings(mailboxId)
		// check settings
		if err == nil {
			// cache setting in redis
			redisworker.StoreCachedInboxSettings(s.srv.ServerConfig, mailboxId, inboxSettings)
		}
	}
	return inboxSettings.RateLimit, err
}

func (s *session) redisIsSessionBlocked() bool {
	if !s.srv.ServerConfig.Redis.Enabled {
		return false
	}

	redisCon := s.srv.ServerConfig.RedisPool.Get()
	defer redisCon.Close()

	redisKey := s.redisRateLimitKey()
	emailsKeyCount, err := redis.Int(redisCon.Do("INCR", redisKey))
	if err == nil {
		if 1 == emailsKeyCount {
			// ttl of key begin
			redisKeyTTL := 1
			if s.rateLimit > 10 {
				redisKeyTTL = 10
			}
			// ttl of key end
			_, err = redisCon.Do("EXPIRE", redisKey, redisKeyTTL)
			if err != nil {
				log.Errorf("redisRateLimits EXPIRE error: %v", err)
			}
		}
	} else {
		log.Errorf("redisRateLimits INCR error: %v", err)
	}

	if err != nil || emailsKeyCount <= s.rateLimit {
		return false
	}

	return true
}

func (s *session) redisRateLimitKey() string {
	return fmt.Sprintf("rate-inbox-limits-%d", s.mailboxId)
}

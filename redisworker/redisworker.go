package redisworker

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/le0pard/go-falcon/config"
	"github.com/le0pard/go-falcon/log"
	"github.com/le0pard/go-falcon/storage"
	"github.com/le0pard/go-falcon/utils"
	"strconv"
	"time"
)

const (
	REDIS_KEY_TTL        = 60
	REDIS_KEY_MAX_COUNT  = 15
	NOTIFICATION_TIMEOUT = 30
	MAX_TTL_FOR_SPAM     = 20
	MAX_EMAILS_FOR_SPAM  = 10
	INBOX_SETTINGS_TTL   = 14400 // 4 hours
)

// get cached inbox setting

func getRedisCacheInboxKey(config *config.Config, mailboxID int) string {
	return fmt.Sprintf("inboxes-settings-cache_%d", mailboxID)
}

func GetCachedInboxSettings(config *config.Config, mailboxID int) (storage.InboxSettings, error) {
	var (
		inboxSettings storage.InboxSettings
		err           error
	)

	redisCacheKey := getRedisCacheInboxKey(config, mailboxID)

	redisCon := config.RedisPool.Get()
	defer redisCon.Close()

	cacheData, err := redis.Values(redisCon.Do("HGETALL", redisCacheKey))
	if err == nil {
		redis.ScanStruct(cacheData, &inboxSettings)
	}
	return inboxSettings, err
}

// store cache inbox settings

func StoreCachedInboxSettings(config *config.Config, mailboxID int, inboxSettings storage.InboxSettings) {
	redisCacheKey := getRedisCacheInboxKey(config, mailboxID)

	redisCon := config.RedisPool.Get()
	defer redisCon.Close()

	redisCon.Do("HMSET", redis.Args{}.Add(redisCacheKey).AddFlat(&inboxSettings)...)
	redisCon.Do("EXPIRE", redisCacheKey, INBOX_SETTINGS_TTL)
}

// send redis notifications

func SendNotifications(config *config.Config, mailboxID, messageID int, subject string) (bool, error) {
	redisCon := config.RedisPool.Get()
	defer redisCon.Close()

	mailboxStr := strconv.Itoa(mailboxID)
	messageStr := strconv.Itoa(messageID)

	data := "{\"channel\": \"/inboxes/" + mailboxStr + "\", \"ext\": {\"username\": \"" + config.Redis.Hook_Username + "\", \"password\": \"" + config.Redis.Hook_Password + "\"}, \"data\": {\"mailbox_id\": \"" + mailboxStr + "\", \"message_id\": \"" + messageStr + "\"}}"

	if config.Redis.Hook_Username != "" && config.Redis.Hook_Password != "" {
		// Faye begin
		clients, err := redis.Strings(redisCon.Do("SUNION", fmt.Sprintf("channels/inboxes/%s", mailboxStr)))
		if err != nil {
			log.Errorf("redis SUNION command error: %v", err)
			return false, err
		}

		if clients != nil {
			for _, clientId := range clients {
				queue := fmt.Sprintf("clients/%s/messages", string(clientId))

				_, err := redisCon.Do("RPUSH", queue, data)
				if err != nil {
					log.Errorf("redis RPUSH command error: %v", err)
					continue
				}
				_, err = redisCon.Do("PUBLISH", "notifications/messages", string(clientId))
				if err != nil {
					log.Errorf("redis PUBLISH command error: %v", err)
					continue
				}
				//cleanup
				cutoff := time.Now().UTC().UnixNano() - (1600 * NOTIFICATION_TIMEOUT)
				score, err := redis.Int64(redisCon.Do("ZSCORE", "clients", string(clientId)))
				if err != nil {
					log.Errorf("redis ZSCORE command error: %v", err)
					continue
				}

				if score > cutoff {
					_, err := redisCon.Do("DEL", queue)
					if err != nil {
						log.Errorf("redis DEL command error: %v", err)
						continue
					}
				}

			}
		}
		// Faye end
	}

	if config.Redis.Sidekiq_Queue != "" && config.Redis.Sidekiq_Class != "" {
		// Sidekiq begin
		data = fmt.Sprintf("{\"retry\":true,\"queue\":\"%s\",\"class\":\"ActiveJob::QueueAdapters::SidekiqAdapter::JobWrapper\",\"args\":[{\"job_class\":\"%s\",\"job_id\":\"%s\",\"queue_name\":\"%s\",\"arguments\":[%s,%s]}],\"jid\":\"%s\",\"enqueued_at\":%s}", config.Redis.Sidekiq_Queue, config.Redis.Sidekiq_Class, utils.GenerateRandString(20), config.Redis.Sidekiq_Queue, mailboxStr, messageStr, utils.GenerateRandString(20), strconv.FormatInt(time.Now().UTC().Unix(), 10))

		redisCon.Send("MULTI")
		redisCon.Send("SADD", "queues", config.Redis.Sidekiq_Queue)
		redisCon.Send("LPUSH", fmt.Sprintf("queue:%s", config.Redis.Sidekiq_Queue), data)
		_, err := redisCon.Do("EXEC")
		if err != nil {
			log.Errorf("redis sidekiq command error: %v", err)
			return false, err
		}
		// Sidekiq end
	}

	return true, nil
}

func IsNotSpamAttackCampaign(config *config.Config, mailboxID int) bool {
	redisCon := config.RedisPool.Get()
	defer redisCon.Close()

	redisKey := fmt.Sprintf("mailbox_msg_count_%d", mailboxID)
	emailsKeyCount, err := redis.Int(redisCon.Do("INCR", redisKey))
	if err == nil {
		if emailsKeyCount > 1 {
			_, err = redisCon.Do("EXPIRE", redisKey, MAX_TTL_FOR_SPAM) // expire 5 min
			if err != nil {
				log.Errorf("CheckIfSendingCampaign EXPIRE error: %v", err)
			}
		}
	} else {
		log.Errorf("CheckIfSendingCampaign INCR error: %v", err)
	}

	return (emailsKeyCount < MAX_EMAILS_FOR_SPAM)
}

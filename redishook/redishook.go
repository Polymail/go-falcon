package redishook

import (
  "strconv"
  "time"
  "fmt"
  "github.com/garyburd/redigo/redis"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/utils"
)

const (
  REDIS_KEY_TTL = 60
  REDIS_KEY_MAX_COUNT = 15
  NOTIFICATION_TIMEOUT = 30
  MAX_TTL_FOR_SPAM = 600
  MAX_EMAILS_FOR_SPAM = 200
)


func SendNotifications(config *config.Config, mailboxID, messageID int, subject string) (bool, error) {
  redisCon := config.RedisPool.Get()
  defer redisCon.Close()

  mailboxStr := strconv.Itoa(mailboxID)
  messageStr := strconv.Itoa(messageID)

  data := "{\"channel\": \"/inboxes/" + mailboxStr + "\", \"ext\": {\"username\": \"" + config.Redis.Hook_Username + "\", \"password\": \"" + config.Redis.Hook_Password + "\"}, \"data\": {\"mailbox_id\": \"" + mailboxStr + "\", \"message_id\": \"" + messageStr + "\"}}"


  if config.Redis.Hook_Username != "" && config.Redis.Hook_Password != "" {
    // Faye begin
    clients, err := redis.Strings(redisCon.Do("SUNION", fmt.Sprintf("%s/channels/inboxes/%s", config.Redis.Namespace, mailboxStr)))
    if err != nil {
      log.Errorf("redis SUNION command error: %v", err)
      return false, err
    }

    if clients != nil {
      for _, clientId := range clients {
        queue := fmt.Sprintf("%s/clients/%s/messages", config.Redis.Namespace, string(clientId))

        _, err := redisCon.Do("RPUSH", queue, data)
        if err != nil {
          log.Errorf("redis RPUSH command error: %v", err)
          continue
        }
        _, err = redisCon.Do("PUBLISH", fmt.Sprintf("%s/notifications/messages", config.Redis.Namespace), string(clientId))
        if err != nil {
          log.Errorf("redis PUBLISH command error: %v", err)
          continue
        }
        //cleanup
        cutoff := time.Now().UTC().UnixNano() - (1600 * NOTIFICATION_TIMEOUT)
        score, err := redis.Int64(redisCon.Do("ZSCORE", fmt.Sprintf("%s/clients", config.Redis.Namespace), string(clientId)))
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
    data = "{\"retry\":true,\"queue\":\"" + config.Redis.Sidekiq_Queue + "\",\"class\":\"" + config.Redis.Sidekiq_Class + "\",\"args\":[" + mailboxStr + ", " + messageStr + "],\"jid\":\"" + utils.GenerateRandString(20) + "\",\"enqueued_at\":" + strconv.FormatInt(time.Now().UTC().Unix(), 10) + "}"

    redisCon.Send("MULTI")
    redisCon.Send("SADD", fmt.Sprintf("%s:queues", config.Redis.Namespace), config.Redis.Sidekiq_Queue)
    redisCon.Send("LPUSH", fmt.Sprintf("%s:queue:%s", config.Redis.Namespace, config.Redis.Sidekiq_Queue), data)
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

  redisKey := fmt.Sprintf("%s:mailbox_msg_count_%d", config.Redis.Namespace, mailboxID)
  emailsKeyCount, err := redis.Int(redisCon.Do("INCR", redisKey))
  if err == nil {
    if 1 == emailsKeyCount {
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
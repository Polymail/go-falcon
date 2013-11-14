package redishook

import (
  "strconv"
  "crypto/md5"
  "encoding/hex"
  "io"
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
)


func SendNotifications(config *config.Config, mailboxID, messageID int, subject string) (bool, error) {
  redisCon := config.RedisPool.Get()
  defer redisCon.Close()

  mailboxStr := strconv.Itoa(mailboxID)
  messageStr := strconv.Itoa(messageID)

  if checkIsSpamAtack(redisCon, mailboxStr, subject) {

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
          _, err = redisCon.Do("PUBLISH", fmt.Sprintf("%s/notifications", config.Redis.Namespace))
          if err != nil {
            log.Errorf("redis PUBLISH command error: %v", err)
            continue
          }
          //cleanup
          cutoff := time.Now().UTC().UnixNano() - 16000
          score, err := redis.Int64(redisCon.Do("ZSCORE", fmt.Sprintf("%s/clients", config.Redis.Namespace)))
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
      data = "{\"retry\":true,\"queue\":\"" + config.Redis.Sidekiq_Queue + "\",\"class\":\"" + config.Redis.Sidekiq_Class + "\",\"args\":[" + messageStr + "],\"jid\":\"" + utils.GenerateRandString(20) + "\",\"enqueued_at\":" + strconv.FormatInt(time.Now().UTC().Unix(), 10) + "}"

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

  }

  return true, nil
}

// check if it is spam attack

func checkIsSpamAtack(redisCon redis.Conn, mailboxId, subject string) bool {
  h := md5.New()
  io.WriteString(h, subject)
  redisKey := fmt.Sprintf("mailbox_last_msg_%s_%s", mailboxId, hex.EncodeToString(h.Sum(nil)))
  res, err := redis.Int(redisCon.Do("EXISTS", redisKey))
  if err != nil {
    log.Errorf("redis checkIsSpamAtack EXISTS key %s: %v", redisKey, err)
    return true
  }
  // exist
  if res != 0 {
    resCount, err := redis.Int(redisCon.Do("GET", redisKey))
    if err != nil {
      log.Errorf("redis checkIsSpamAtack GET key %s: %v", redisKey, err)
      return true
    }
    if resCount > 0 && resCount < REDIS_KEY_MAX_COUNT {
      setRedisAttackKey(redisCon, redisKey)
    // too many dublicates
    } else {
      _, err = redisCon.Do("EXPIRE", redisKey, REDIS_KEY_TTL)
      if err != nil {
        log.Errorf("redis checkIsSpamAtack EXPIRE key %s: %v", redisKey, err)
      }
      return false
    }
  // not exist
  } else {
    setRedisAttackKey(redisCon, redisKey)
  }
  return true
}

// set key

func setRedisAttackKey(redisCon redis.Conn, redisKey string) {
  redisCon.Send("MULTI")
  redisCon.Send("INCR", redisKey)
  redisCon.Send("EXPIRE", redisKey, REDIS_KEY_TTL)
  _, err := redisCon.Do("EXEC")
  if err != nil {
    log.Errorf("redis setRedisAttackKey INCR key %s: %v", redisKey, err)
  }
}
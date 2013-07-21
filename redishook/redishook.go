package redishook

import (
  "bytes"
  "strconv"
  "crypto/md5"
  "encoding/hex"
  "io"
  "time"
  "github.com/garyburd/redigo/redis"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/utils"
)

const RedisKeyTTL = 60
const RedisKeyMaxCount = 15


func SendNotifications(config *config.Config, mailboxID, messageID int, subject string) (bool, error) {
  var buffer bytes.Buffer
  buffer.WriteString(config.Redis.Host)
  buffer.WriteString(":")
  buffer.WriteString(strconv.Itoa(config.Redis.Port))

  redisCon, err := redis.Dial("tcp", buffer.String())
  if err != nil {
    log.Errorf("Error connect to redis: %v", err)
    return false, err
  }
  defer redisCon.Close()

  mailboxStr := strconv.Itoa(mailboxID)
  messageStr := strconv.Itoa(messageID)


  if checkIsSpamAtack(redisCon, mailboxStr, subject) {

    data := "{\"channel\": \"/inboxes/" + mailboxStr + "\", \"ext\": {\"username\": \"" + config.Redis.Hook_Username + "\", \"password\": \"" + config.Redis.Hook_Password + "\"}, \"data\": {\"mailbox_id\": \"" + mailboxStr + "\", \"message_id\": \"" + messageStr + "\"}}"


    if config.Redis.Hook_Username != "" && config.Redis.Hook_Password != "" {
      // Faye begin
      clients, err := redis.Strings(redisCon.Do("SUNION", config.Redis.Namespace + "/channels/inboxes/" + mailboxStr))
      if err != nil {
        log.Errorf("redis SUNION command error: %v", err)
        return false, err
      }

      if clients != nil {
        for _, clientId := range clients {
          queue := config.Redis.Namespace + "/clients/" + string(clientId) + "/messages"

          _, err := redisCon.Do("RPUSH", queue, data)
          if err != nil {
            log.Errorf("redis RPUSH command error: %v", err)
            continue
          }
          _, err = redisCon.Do("PUBLISH", config.Redis.Namespace + "/notifications", clientId)
          if err != nil {
            log.Errorf("redis PUBLISH command error: %v", err)
            continue
          }
          //cleanup
          cutoff := time.Now().UTC().UnixNano() - 16000
          score, err := redis.Int64(redisCon.Do("ZSCORE", config.Redis.Namespace + "/clients", clientId))
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
      redisCon.Send("SADD", config.Redis.Namespace + ":queues", config.Redis.Sidekiq_Queue)
      redisCon.Send("LPUSH", config.Redis.Namespace + ":queue:" + config.Redis.Sidekiq_Queue, data)
      _, err = redisCon.Do("EXEC")
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
  redisKey := "mailbox_last_msg_" + mailboxId + "_" + hex.EncodeToString(h.Sum(nil))
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
    if resCount > 0 && resCount < RedisKeyMaxCount {
      setRedisAttackKey(redisCon, redisKey)
    // too many dublicates
    } else {
      _, err = redisCon.Do("EXPIRE", redisKey, RedisKeyTTL)
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
  redisCon.Send("EXPIRE", redisKey, RedisKeyTTL)
  _, err := redisCon.Do("EXEC")
  if err != nil {
    log.Errorf("redis setRedisAttackKey INCR key %s: %v", redisKey, err)
  }
}
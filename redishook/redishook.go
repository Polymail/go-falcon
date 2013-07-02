package redishook

import (
  "bytes"
  "strconv"
  "time"
  "github.com/garyburd/redigo/redis"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
  "github.com/le0pard/go-falcon/utils"
)


func SendNotifications(config *config.Config, mailboxID, messageID int) (bool, error) {
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
    _, err = redisCon.Do("SADD", config.Redis.Namespace + ":queues", config.Redis.Sidekiq_Queue)
    if err != nil {
      log.Errorf("redis SADD command error: %v", err)
      return false, err
    }
    data = "{\"retry\":true,\"queue\":\"" + config.Redis.Sidekiq_Queue + "\",\"class\":\"" + config.Redis.Sidekiq_Class + "\",\"args\":[" + messageStr + "],\"jid\":\"" + utils.GenerateRandString(20) + "\",\"enqueued_at\":" + strconv.FormatInt(time.Now().UTC().Unix(), 10) + "}"
    _, err = redisCon.Do("LPUSH", config.Redis.Namespace + ":queue:" + config.Redis.Sidekiq_Queue, data)
    if err != nil {
      log.Errorf("redis LPUSH command error: %v", err)
      return false, err
    }
    // Sidekiq end
  }

  return true, nil
}
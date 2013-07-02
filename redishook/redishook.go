package redishook

import (
  "bytes"
  "strconv"
  "github.com/garyburd/redigo/redis"
  "github.com/le0pard/go-falcon/log"
  "github.com/le0pard/go-falcon/config"
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

  reply, err := redisCon.Do("SUNION", config.Redis.Namespace + "/channels/inboxes/" + mailboxStr)
  if err != nil {
    log.Errorf("redis SUNION command error: %v", err)
    return false, err
  }

  clients, err := redis.Strings(reply, nil)
  if err != nil {
    log.Errorf("redis Strings error: %v", err)
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
    }
  }

  return true, nil
}
package postgresql

import (
  "fmt"
  _ "github.com/lib/pq"
  "database/sql"
  "github.com/le0pard/go-falcon/config"
)


func InitDatabase(config *config.Config) (*sql.DB, error) {
  return sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", config.Storage.Host, config.Storage.Port, config.Storage.Username, config.Storage.Password, config.Storage.Database))
}
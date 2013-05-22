package postgresql

import (
  "strconv"
  _ "github.com/bmizerany/pq"
  "database/sql"
  "github.com/le0pard/go-falcon/config"
)


func InitDatabase(config *config.Config) (*sql.DB, error) {
  return sql.Open("postgres", "host="+config.Storage.Host+" port="+strconv.Itoa(config.Storage.Port)+" user="+config.Storage.Username+" password="+config.Storage.Password+" dbname="+config.Storage.Database+" sslmode=disable")
}
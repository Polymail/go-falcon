package postgresql

import (
  "strconv"
  _ "github.com/bmizerany/pq"
  "database/sql"
)


func InitDatabase(config *config.Config) (*sql.DB, error) {
  sql.Open("postgres", "host="+config.Storage.Host+" port="+strconv.Itoa(config.Storage.Port)+" user="+config.Storage.Username+" password="+config.Storage.Password+" dbname="+config.Storage.Database+" sslmode=require")
}
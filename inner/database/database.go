package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"idm/inner/common"
	"time"
)

var DB *sqlx.DB

func ConnectDb() *sqlx.DB {
	cfg := common.GetConfig(".env")
	return ConnectDbWithCfg(cfg)
}

// ConnectDbWithCfg подключиться к базе данных с переданным конфигом
func ConnectDbWithCfg(cfg common.Config) *sqlx.DB {
	var db = sqlx.MustConnect(cfg.DbDriverName, cfg.Dsn)
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(1 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)
	return db
}

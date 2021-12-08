package query

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
)

func MigrateDatabase(cfg mysql.Config) {
	dbname := cfg.DBName
	cfg.DBName = ""
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return
	}
	db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbname))
}

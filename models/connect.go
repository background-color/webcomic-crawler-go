package models

import (
	"database/sql"
	"time"

	"github.com/go-sql-driver/mysql"
)

func DBConnect(dbName string, dbUser string, dbPass string, dbAddr string) (*sql.DB, error) {
	// 接続設定
	locale, _ := time.LoadLocation("Asia/Tokyo")
	c := mysql.Config{
		DBName:    dbName,
		User:      dbUser,
		Passwd:    dbPass,
		Addr:      dbAddr,
		Net:       "tcp",
		Collation: "utf8mb4_general_ci",
		ParseTime: true,
		Loc:       locale,
	}

	db, err := sql.Open("mysql", c.FormatDSN())
	if err != nil {
		return nil, err
	}

	// 確認
	pingErr := db.Ping()
	if pingErr != nil {
		return nil, pingErr
	}

	return db, nil
}

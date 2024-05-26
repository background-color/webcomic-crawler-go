package models

import (
	"database/sql"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func DBConnect() (*sql.DB, error) {
	var err error

	// .envファイルを読み込む
	err = godotenv.Load()
	if err != nil {
		//logger.Error("Error open .env file")
		return nil, err
	}

	// 環境変数を変数に格納する
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbAddr := os.Getenv("DB_ADDRESS")

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

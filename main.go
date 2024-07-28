package main

import (
	"database/sql"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/background-color/webcomic-crawler-go/models"
	"github.com/background-color/webcomic-crawler-go/rss"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/joho/godotenv"
)

func main() {
	// .envファイルを読み込む
	godotenv.Load()

	logger, err := GetLogger(os.Getenv("LOG_FILE_PATH"))
	if err != nil {
		return
	}
	logger.Info("---------- start")

	db, err := models.DBConnect(
		os.Getenv("DB_NAME"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_ADDRESS"),
	)
	if err != nil {
		logger.Error("error", err)
		return
	}
	defer db.Close()

	// SQLクエリの作成
	query := `
select
    c.id, c.name, c.url, c.chk_url,
    s.url_type, s.check_field,
    r.check_text
from comic as c
inner join site as s on c.site_id = s.id
left join (select comic_id, max(id) as id from rss group by comic_id) as r_max on c.id = r_max.comic_id
left join rss as r on r_max.id = r.id
where c.is_disabled = 0
`
	// CHROMIUM_PATHが指定されていればそれを使う
	var browser *rod.Browser
	chromiumPath := os.Getenv("CHROMIUM_PATH")
	if chromiumPath != "" {
		u := launcher.New().Bin(chromiumPath).MustLaunch()
		browser = rod.New().ControlURL(u).MustConnect()
	} else {
		browser = rod.New().MustConnect()
	}

	defer browser.MustClose()

	rows, err := db.Query(query)
	if err != nil {
		logger.Error("Failed to execute query", slog.Any("error", err))
		return
	}
	defer rows.Close()

	// 登録用
	stmtIns, err := db.Prepare("INSERT INTO rss (`comic_id`, `check_text`) VALUES( ?, ? )")
	if err != nil {
		logger.Error("Failed query", slog.Any("error", err))
		return
	}
	defer stmtIns.Close()

	// スクレイピング、変更があれば登録
	for rows.Next() {
		var (
			id, name, url, checkField  string
			urlType, chkUrl, checkText sql.NullString
		)
		err := rows.Scan(&id, &name, &url, &chkUrl, &urlType, &checkField, &checkText)
		if err != nil {
			logger.Error("Failed to scan row", slog.Any("error", err))
			return
		}

		logger.Info("site", "id", id, "name", name, "url", url)

		page := browser.MustPage(url)
		// defer page.MustClose()

		elText := page.Timeout(5 * time.Second).MustElement(checkField).MustText()
		page.MustClose()

		logger.Info("get element", "タイトル", elText)

		if elText != checkText.String {
			logger.Info("update: id", id, elText)
			stmtIns.Exec(id, elText)

		}
	}

	err = rss.GenerateRSSFeed(db, os.Getenv("RSS_FILE_PATH"))
	if err != nil {
		logger.Error("Failed to generate RSS feed", slog.Any("error", err))
	}
	logger.Info("---------- end")
}

func GetLogger(logFilePath string) (*slog.Logger, error) {
	// ログ出力先
	if logFilePath == "" {
		logFilePath = "log/sample.log"
	}

	// ログファイルを開く（存在しない場合は作成）
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	logfile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return slog.New(slog.NewJSONHandler(logfile, nil)), nil
}

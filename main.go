package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/background-color/webcomic-crawler-go/models"
	"github.com/go-rod/rod"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Debug("---------- start startCrawl()")

	db, err := models.DBConnect()
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

	browser := rod.New().MustConnect()
	defer browser.MustClose()

	rows, err := db.Query(query)
	defer rows.Close()
	if err != nil {
		logger.Error("Failed to execute query", slog.Any("error", err))
		return
	}
	defer rows.Close()

	// 登録用
	stmtIns, err := db.Prepare("INSERT INTO rss (`comic_id`, `check_text`) VALUES( ?, ? )")
	if err != nil {
		logger.Error("Failed to execute query", slog.Any("error", err))
		return
	}
	defer stmtIns.Close()

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

		fmt.Printf("ID: %v, Name: %s, URL: %s\n", id, name, url)

		page := browser.MustPage(url)
		elText := page.Timeout(10 * time.Second).MustElement(checkField).MustText()
		fmt.Printf("タイトル: %s\n", elText)

		if elText != checkText.String {
			fmt.Printf("更新")
			logger.Info("update: id", id, elText)
			stmtIns.Exec(id, elText)

		}
	}
}

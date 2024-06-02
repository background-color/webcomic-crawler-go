package rss

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/gorilla/feeds"
)

type RSSItem struct {
	Id        int
	CheckText string
	Name      string
	Url       string
	Ins       time.Time
}

func GenerateRSSFeed(db *sql.DB, filePath string) error {
	rssRows, err := fetchRSSRows(db)
	if err != nil {
		return fmt.Errorf("failed fetch rss rows: %w", err)
	}
	defer rssRows.Close()

	feed, err := createFeed(rssRows)
	if err != nil {
		return fmt.Errorf("failed create feed: %w", err)
	}

	return outputRssFile(feed, filePath)
}

// RSS出力内容取得
func fetchRSSRows(db *sql.DB) (*sql.Rows, error) {
	const rssQuery = `
		SELECT t1.id, t1.check_text, t2.name, t2.url, t1.ins
		FROM rss AS t1 
		INNER JOIN comic AS t2 ON t1.comic_id = t2.id
		ORDER BY t1.id DESC
		LIMIT 30
	`
	return db.Query(rssQuery)
}

// Feed作成
func createFeed(rows *sql.Rows) (*feeds.Feed, error) {
	now := time.Now()
	feed := &feeds.Feed{
		Title:       "ALL RSS",
		Link:        &feeds.Link{Href: "rss.background-color.jp"},
		Description: "WEB COMIC RSS",
		Created:     now,
	}

	var feedItems []*feeds.Item

	for rows.Next() {
		item, err := createFeedItem(rows)
		if err != nil {
			return nil, err
		}
		feedItems = append(feedItems, item)
	}

	feed.Items = feedItems
	return feed, nil
}

// Feedの items作成
func createFeedItem(rows *sql.Rows) (*feeds.Item, error) {
	var rssItem RSSItem

	if err := rows.Scan(&rssItem.Id, &rssItem.CheckText, &rssItem.Name, &rssItem.Url, &rssItem.Ins); err != nil {
		return nil, err
	}

	return &feeds.Item{
		Title:       rssItem.Name,
		Link:        &feeds.Link{Href: rssItem.Url},
		Description: rssItem.CheckText,
		Created:     rssItem.Ins,
	}, nil
}

// RSSファイル出力
func outputRssFile(feed *feeds.Feed, filePath string) error {
	rss, err := feed.ToRss()
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(rss); err != nil {
		return err
	}

	return nil
}

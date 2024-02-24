package tinysearch

import (
	"context"
	"database/sql"
	"log"
	"net/url"
	"time"

	"github.com/gocolly/colly"
	normalizeurl "github.com/sekimura/go-normalize-url"
)

func CreateCollector(db *sql.DB) *colly.Collector {
	c := colly.NewCollector(
		colly.MaxDepth(1),
		colly.CacheDir("./cache"),
	)

	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	// Find and visit all links
	c.OnHTML("body", func(element *colly.HTMLElement) {
		url := NormalizeUrl(element.Request.URL)
		content := element.Text

		err := InsertPage(db, url, content)
		if err != nil {
			log.Printf("Failed to insert page %s: %s\n", url, err)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Failed to visite page %s: %s\n", r.Request.URL, err)
	})

	return c
}

func CreateDbConnection(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}

	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(3)
	db.SetMaxOpenConns(3)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}

	return db
}

func NormalizeUrl(url *url.URL) string {
	urlStr := url.String()
	normalized, err := normalizeurl.Normalize(urlStr)
	if err != nil {
		return urlStr
	}
	return normalized
}

func InsertPage(db *sql.DB, url string, content string) error {
	// Open transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	//  Prepare statement
	stmt, err := tx.Prepare("INSERT INTO page (url, content) values(?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Exacute statement
	_, err = stmt.Exec(url, content)
	if err != nil {
		return err
	}

	return nil
}

func Index(seeds []string) {
	db := CreateDbConnection("dev.db")
	collector := CreateCollector(db)

	for _, seed := range seeds {
		collector.Visit(seed)
	}
}

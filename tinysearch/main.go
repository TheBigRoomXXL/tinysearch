package tinysearch

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gocolly/colly"
	_ "github.com/mattn/go-sqlite3"
	normalizeurl "github.com/sekimura/go-normalize-url"
)

func CreateCollector(db *sql.DB) *colly.Collector {
	c := colly.NewCollector(
		colly.MaxDepth(3),
		colly.CacheDir(".cache/"),
		colly.DisallowedDomains("*.microsoft.com", "*.facebook.com", "*.google.com"),
	)

	// Find and visit all links
	c.OnHTML("body", func(element *colly.HTMLElement) {
		url := element.Request.URL.String()
		fmt.Println("visiting ", url)
		content := element.Text
		err := InsertPage(db, url, content)
		if err != nil {
			log.Printf("Failed to insert page %s: %s\n", url, err)
		}
	})

	// Find and visit all links
	c.OnHTML("a[href]", func(element *colly.HTMLElement) {
		link := element.Attr("href")
		c.Visit(element.Request.AbsoluteURL(link))
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

	db.SetConnMaxLifetime(time.Duration(10))
	db.SetMaxOpenConns(1)

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
	fmt.Println(" - start inserting...", url)
	// Open transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	//  Prepare statement
	stmt, err := tx.Prepare("INSERT INTO pages (url, content) values(?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Exacute statement
	_, err = stmt.Exec(url, content)
	if err != nil {
		return err
	}

	// Finish transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	fmt.Println(" - finish inserting...", url)
	return nil
}

func Index(seeds []string) {
	db := CreateDbConnection("data/dev.db")
	collector := CreateCollector(db)

	for _, seed := range seeds {
		collector.Visit(seed)
	}
}

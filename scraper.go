package main

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gauravpatil2468/rssagg/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func startScrapping(database *db.Queries, concurrency int, timeBetweenRequest time.Duration) {

	log.Printf("Scrapping on %v goroutines every %s seconds", concurrency, timeBetweenRequest)
	ticker := time.NewTicker(timeBetweenRequest)
	for ; ; <-ticker.C {
		feeds, err := database.GetNextFeedsToFetch(context.Background(), int32(concurrency))
		if err != nil {
			log.Println("Error fetching feeds:", err)
			continue
		}
		wg := &sync.WaitGroup{}
		for _, feed := range feeds {
			wg.Add(1)

			go scrapeFeed(database, wg, feed)
		}
		wg.Wait()
	}
}

func scrapeFeed(database *db.Queries, wg *sync.WaitGroup, feed db.Feed) {
	defer wg.Done()

	_, err := database.MarkFeedAsFetched(context.Background(), feed.ID)
	if err != nil {
		log.Println("Error marking feed as fetched:", err)
		return
	}

	rssFeed, err := urlToFeed(feed.Url)
	if err != nil {
		log.Println("Error fetching feed:", err)
		return
	}

	for _, item := range rssFeed.Channel.Item {
		t, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			log.Printf("Error parsing date time:%v", err)
		}
		_, err = database.CreatePost(context.Background(), db.CreatePostParams{
			ID: pgtype.UUID{
				Bytes: uuid.New(),
				Valid: true,
			},
			CreatedAt: pgtype.Timestamp{
				Time:  time.Now(),
				Valid: true,
			},
			UpdatedAt: pgtype.Timestamp{
				Time:  time.Now(),
				Valid: true,
			},
			Title: item.Title,
			Description: pgtype.Text{
				String: item.Description,
				Valid:  true,
			},
			PublishedAt: pgtype.Timestamp{
				Time:  t,
				Valid: true,
			},
			Url:    item.Link,
			FeedID: feed.ID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				continue
			}
			log.Printf("Failed to create post:%v", err)
		}
	}
	log.Printf("Feed %s collected, %v posts found", feed.Name, len(rssFeed.Channel.Item))

}

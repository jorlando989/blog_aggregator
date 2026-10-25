package main

import (
	"context"
	"blogaggregator/internal/database"
	"net/http"
	"time"
	"html"
	"encoding/xml"
	"fmt"
	"strings"
	"os"
	"github.com/google/uuid"
	"database/sql"
	"io"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")

	client := http.Client{Timeout: time.Duration(5) * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var v RSSFeed
	err = xml.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}

	v.Channel.Title = html.UnescapeString(v.Channel.Title)
	v.Channel.Title = html.UnescapeString(v.Channel.Description)

	for _, x := range v.Channel.Item {
		x.Title = html.UnescapeString(x.Title)
		x.Description = html.UnescapeString(x.Description)
	}

	return &v, nil
}

func scrapeFeeds(s *state) error {
	//get next feed to fetch from db
	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Println(err)
		return err
	}

	//mark as fetched
	markFeedParams := database.MarkFeedFetchedParams{
		LastFetchedAt: timeToNullTime(time.Now()),
		ID: nextFeed.ID,
	}
	err = s.db.MarkFeedFetched(context.Background(), markFeedParams)
	if err != nil {
		fmt.Println(err)
		return err
	}

	//fetch feed
	feed, err := fetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		fmt.Println(err)
		return err
	}

	//save posts
	for _, x := range feed.Channel.Item {
		createPostParams := database.CreatePostParams{
			ID: uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Title: stringToNullString(x.Title),
			Url: stringToNullString(x.Link),
			Description: stringToNullString(x.Description),
			PublishedAt: stringToNullTime(x.PubDate),
			FeedID: nextFeed.ID,
		}
		
		_, err := s.db.CreatePost(context.Background(), createPostParams)
		 if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				continue
			}
			fmt.Println(err)
			continue
		}
	}
	return nil
}

func stringToNullTime(s string) sql.NullTime {
	t, err := time.Parse(time.RFC1123, s)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return sql.NullTime{Time: t, Valid: true}
}

func timeToNullTime(t time.Time) sql.NullTime {
    return sql.NullTime{Time: t, Valid: true}
}

func stringToNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}
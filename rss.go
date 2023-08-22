package sn

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"time"
)

func (c *RssDate) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	dateFormat := "Mon, 02 Jan 2006 15:04:05 GMT"
	d.DecodeElement(&v, &start)
	parse, err := time.Parse(dateFormat, v)
	if err != nil {
		return err
	}
	*c = RssDate{parse}
	return nil
}

var (
	StackerNewsRssFeedUrl = "https://stacker.news/rss"
)

// Fetch RSS feed
func RssFeed() (*Rss, error) {
	resp, err := http.Get(StackerNewsRssFeedUrl)
	if err != nil {
		err = fmt.Errorf("error fetching RSS feed: %w", err)
		log.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	var rss Rss
	err = xml.NewDecoder(resp.Body).Decode(&rss)
	if err != nil {
		err = fmt.Errorf("error decoding RSS feed XML: %w", err)
		return nil, err
	}

	return &rss, nil
}

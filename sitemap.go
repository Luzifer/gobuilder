package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type xmlSitemap struct {
	XMLName xml.Name        `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
	URLs    []xmlSitemapURL `xml:"url"`
}

type xmlSitemapURL struct {
	Location     string         `xml:"loc"`
	LastModified xmlSitemapTime `xml:"lastmod"`
	Priority     float64        `xml:"priority,omitempty"`
}

type xmlSitemapTime struct {
	time.Time
}

func (x xmlSitemapTime) MarshalText() ([]byte, error) {
	t := x.UTC().Format("2006-01-02T15:04:05-07:00")
	return []byte(t), nil
}

func handleXMLSitemap(res http.ResponseWriter, r *http.Request) {
	out := xmlSitemap{}
	out.URLs = append(out.URLs, xmlSitemapURL{
		Location:     "https://gobuilder.me/",
		LastModified: xmlSitemapTime{Time: time.Now()},
		Priority:     0,
	})

	l, err := redisClient.ZRangeByScore("last-builds", "-inf", "+inf", true, false, 0, 0)
	if err != nil {
		http.Error(res, "An error ocurred", http.StatusInternalServerError)
		return
	}

	i := 0
	for i < len(l) {
		repo := l[i]
		i++
		t, err := strconv.ParseInt(l[i], 10, 64)
		if err != nil {
			http.Error(res, "An error ocurred", http.StatusInternalServerError)
			return
		}

		out.URLs = append(out.URLs, xmlSitemapURL{
			Location:     fmt.Sprintf("https://gobuilder.me/%s", repo),
			LastModified: xmlSitemapTime{Time: time.Unix(t, 0)},
			Priority:     0.5,
		})

		i++
	}

	res.Header().Set("Content-Type", "application/xml")
	xml.NewEncoder(res).Encode(out)
}

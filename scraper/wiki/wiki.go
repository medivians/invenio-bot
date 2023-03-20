package wiki

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Wiki struct {
	httpCli *http.Client
}

type Locations []string

// New Return new client medivia
func New() *Wiki {
	return &Wiki{httpCli: &http.Client{Timeout: time.Second * 5}}
}

func (w *Wiki) WhereToSell(n string) Locations {
	cleanedName := cleanName(n)
	log.Printf("searching for %v", cleanedName)
	resp, err := w.fetchItem(cleanedName)
	if err != nil {
		log.Printf("Error fetching item %v", err)
		return nil
	}
	tkr := html.NewTokenizer(resp.Body)
	var locs Locations
	for {
		tt := tkr.Next()
		token := tkr.Token()
		if tt == html.TextToken && token.Data == "Sell to:" {
			tt = tkr.Next()
			tt = tkr.Next()
			tt = tkr.Next()
			tt = tkr.Next()
			tt = tkr.Next()
			for {
				tt := tkr.Next()
				token := tkr.Token()
				if tt == html.StartTagToken && token.Data == "th" {
					tt = tkr.Next()
					tt = tkr.Next()
					tt = tkr.Next()
					token = tkr.Token()
					locs = append(locs, token.Data)
				}
				if tt == html.StartTagToken && token.Data == "td" {
					var info strings.Builder
					for {
						tt = tkr.Next()
						token = tkr.Token()
						if tt == html.TextToken {
							info.WriteString(" " + token.Data)
						}
						if tt == html.EndTagToken && token.Data == "td" {
							break
						}
					}
					locs = append(locs, info.String())
				}
				if (tt == html.EndTagToken && token.Data == "tbody") || tt == html.ErrorToken {
					break
				}
			}

		}
		if (tt == html.EndTagToken && token.Data == "table") || tt == html.ErrorToken {
			break
		}
	}
	return locs
}

func (w *Wiki) fetchItem(n string) (*http.Response, error) {
	if n == "" {
		return nil, fmt.Errorf("empty name")
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://wiki.mediviastats.info/%v", n), nil)
	if err != nil {
		return nil, err
	}

	// req.Header.Add("authority", "wiki.mediviastats.info")
	// req.Header.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	// req.Header.Add("accept-language", "en-US,en;q=0.9,pt-BR;q=0.8,pt;q=0.7")
	// req.Header.Add("cache-control", "max-age=0")
	// req.Header.Add("dnt", "1")
	// req.Header.Add("if-modified-since", "Sun, 18 Dec 2022 01:54:44 GMT")
	// req.Header.Add("referer", "https://www.google.com/")
	// req.Header.Add("sec-ch-ua", "Google Chrome;v=111, Not(A:Brand;v=8, Chromium;v=111")
	// req.Header.Add("sec-ch-ua-mobile", "?0")
	// req.Header.Add("sec-ch-ua-platform", "Windows")
	// req.Header.Add("sec-fetch-dest", "document")
	// req.Header.Add("sec-fetch-mode", "navigate")
	// req.Header.Add("sec-fetch-site", "cross-site")
	// req.Header.Add("sec-fetch-user", "?1")
	// req.Header.Add("upgrade-insecure-requests", "1")
	// req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36")
	return w.httpCli.Do(req)
}

func cleanName(n string) string {
	return strings.ReplaceAll(strings.Title(n), " ", "_")
}

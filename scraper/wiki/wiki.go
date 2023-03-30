package wiki

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
			tkr.Next()
			tkr.Next()
			tkr.Next()
			tkr.Next()
			tt = tkr.Next()
			for {
				tt := tkr.Next()
				token := tkr.Token()
				if tt == html.StartTagToken && token.Data == "th" {
					tkr.Next()
					tkr.Next()
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
		if tt == html.ErrorToken {
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
	return w.httpCli.Do(req)
}

func cleanName(n string) string {
	var articlesAndPrepositions = map[string]struct{}{"the": {}, "of": {}}

	partials := strings.Split(n, " ")
	var cleanedName string
	for i, partialName := range partials {
		if i == 0 {
			cleanedName += cases.Title(language.English).String(partialName)
			continue
		}
		if _, ok := articlesAndPrepositions[partialName]; ok {
			cleanedName += "_" + cases.Lower(language.English).String(partialName)
			continue
		}
		cleanedName += "_" + cases.Title(language.English).String(partialName)
	}
	return cleanedName
}

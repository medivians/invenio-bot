package mediviastats

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Information []string
type KillList []string

type MediviaStats struct {
	httpCli *http.Client
}

// New Return new client MediviaStats
func New() *MediviaStats {
	return &MediviaStats{httpCli: &http.Client{Timeout: time.Second * 5}}
}

func (m *MediviaStats) WhoIs(n string) Information {
	resp, err := m.fetchCharacter(n)
	if err != nil {
		log.Printf("error %v", err)
		return nil
	}
	tkr := html.NewTokenizer(resp.Body)
	var info Information
	for {
		tt := tkr.Next()
		token := tkr.Token()
		if tt == html.StartTagToken && token.Data == "table" && checkClass(token.Attr, "table-condensed table-bordered") {
			for {
				tt := tkr.Next()
				token := tkr.Token()
				if tt == html.StartTagToken && token.Data == "tr" {
					var information string
					for {
						tt = tkr.Next()
						token := tkr.Token()
						if tt == html.TextToken {
							information = information + " " + token.Data
							continue
						}
						if tt == html.EndTagToken && token.Data == "tr" {
							info = append(info, strings.TrimSpace(information))
							break
						}
					}
				}
				if tt == html.StartTagToken && token.Data == "div" {
					return info
				}

				if (tt == html.EndTagToken && token.Data == "tbody") || tt == html.ErrorToken {
					break
				}
			}

			if (tt == html.EndTagToken && token.Data == "tbody") || tt == html.ErrorToken {
				break
			}
		}
	}
	return nil
}

func (m *MediviaStats) KillList(n string) KillList {
	resp, err := m.fetchCharacter(n)
	if err != nil {
		log.Printf("error %v", err)
		return nil
	}
	tkr := html.NewTokenizer(resp.Body)
	var kl KillList

	for {
		tt := tkr.Next()
		token := tkr.Token()
		if tt == html.StartTagToken && token.Data == "h2" {
			tt := tkr.Next()
			token := tkr.Token()
			if tt == html.TextToken && token.Data == "Frags" {
				for {
					tt := tkr.Next()
					token = tkr.Token()
					if tt == html.StartTagToken && token.Data == "th" {
						tkr.Next()
						token := tkr.Token()
						kl = append(kl, token.Data)
						continue
					}
					if tt == html.StartTagToken && token.Data == "td" {
						var killItem string
						for {
							tt = tkr.Next()
							token = tkr.Token()
							if tt == html.TextToken {
								killItem = killItem + " " + token.Data
								continue
							}
							if tt == html.EndTagToken && token.Data == "tr" {
								kl = append(kl, killItem)
								break
							}
						}
					}
					if tt == html.EndTagToken && token.Data == "table" {
						return kl
					}

					if (tt == html.EndTagToken && token.Data == "tbody") || tt == html.ErrorToken {
						break
					}
				}
				if (tt == html.EndTagToken && token.Data == "tbody") || tt == html.ErrorToken {
					break
				}
			}
		}
		if (tt == html.EndTagToken && token.Data == "tbody") || tt == html.ErrorToken {
			break
		}
	}
	return nil
}

func (m *MediviaStats) fetchCharacter(n string) (*http.Response, error) {
	if n == "" {
		return nil, fmt.Errorf("empty name")
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://mediviastats.info/characters.php?name=%v", url.QueryEscape(n)), nil)
	if err != nil {
		return nil, err
	}
	return m.httpCli.Do(req)
}

func checkClass(attrs []html.Attribute, class string) bool {
	for _, attr := range attrs {
		if attr.Val == class {
			return true
		}
	}
	return false
}

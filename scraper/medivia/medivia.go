package medivia

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Character struct {
	Informations []string
}

type Medivia struct {
	httpCli *http.Client
}

// New Return new client medivia
func New() *Medivia {
	return &Medivia{httpCli: &http.Client{Timeout: time.Second * 5}}
}

// WhoIs return all information knowledge about some character in medivia server
func (m *Medivia) WhoIs(n string) (*Character, error) {
	resp, err := m.fetchCharacter(n)
	if err != nil {
		return nil, fmt.Errorf("fetch info unavailable")
	}

	c := Character{}
	tkr := html.NewTokenizer(resp.Body)
	for {
		tt := tkr.Next()
		if tt == html.StartTagToken {
			token := tkr.Token()

			if token.Data == "div" && checkClass(token.Attr, "med-width-50") {
				tt = tkr.Next()
				token := tkr.Token()
				if tt == html.TextToken {
					c.Informations = append(c.Informations, token.Data)
				}
				if tt == html.StartTagToken && token.Data == "span" {
					tkr.Next()
					tt = tkr.Next()
					token = tkr.Token()
					if tt == html.TextToken {
						c.Informations = append(c.Informations, token.Data)
					}
				}
				if tt == html.StartTagToken && token.Data == "strong" {
					tt = tkr.Next()
					token = tkr.Token()
					if tt == html.TextToken {
						c.Informations = append(c.Informations, token.Data)
					}
				}
			}

			if token.Data == "div" && checkClass(token.Attr, "med-footer-ivy") {
				break
			}
			if tt == html.ErrorToken {
				fmt.Printf("problem %v", token.Data)
				return nil, fmt.Errorf("reading tokens: %w", tkr.Err())
			}
		}
	}

	if len(c.Informations) == 0 {
		return nil, fmt.Errorf("player not found")
	}
	return &c, nil
}

type KillList []string

func (m *Medivia) KillList(n string) (KillList, error) {
	resp, err := m.fetchCharacter(n)
	if err != nil {
		return nil, fmt.Errorf("fetch unavailable %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch wrong status code %v", resp.StatusCode)
	}

	var killList KillList
	tkr := html.NewTokenizer(resp.Body)

	for {
		tt := tkr.Next()
		if tt == html.StartTagToken {
			token := tkr.Token()

			if token.Data == "div" && checkClass(token.Attr, "title") {
				tt = tkr.Next()
				token = tkr.Token()
				if tt == html.TextToken && token.Data == "Kill list" {
					for {
						if tt == html.TextToken && token.Data == "Task list" {
							return killList, nil
						}
						tt = tkr.Next()
						token = tkr.Token()
						if checkClass(token.Attr, "med-width-75") {
							tkr.Next()
							token = tkr.Token()
							prefix := strings.TrimSpace(token.Data)
							tkr.Next()
							tkr.Next()
							token = tkr.Token()
							player := strings.TrimSpace(token.Data)
							tkr.Next()
							tt = tkr.Next()
							token = tkr.Token()
							sufix := strings.TrimSpace(token.Data)
							killList = append(killList, fmt.Sprintf("%v %v %v", prefix, player, sufix))
						}
					}
				}
				if tt == html.TextToken && token.Data == "Task list" {
					break
				}
			}

			if tt == html.ErrorToken {
				fmt.Printf("problem %v", token.Data)
				return nil, fmt.Errorf("reading tokens: %w", tkr.Err())
			}
		}
	}
	return nil, fmt.Errorf("player not found")
}

func (m *Medivia) fetchCharacter(n string) (*http.Response, error) {
	if n == "" {
		return nil, fmt.Errorf("empty name")
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://medivia.online/community/character/%v", n), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("authority", "medivia.online")

	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9,pt-BR;q=0.8,pt;q=0.7")
	req.Header.Add("Cache-Control", "max-age=0")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Cookie", "cf_chl_2=368403408f9f7ad; cf_clearance=uOlfkJApuEudL3x7Cn1ndadglq80YzdJdN7mqiwks28-1679360241-0-150; _ga=GA1.2.2143851611.1679360238; _gid=GA1.2.2036756327.1679360238; XSRF-TOKEN=eyJpdiI6IjRRSk1wVFZ3R1lnNnhPaVgwdk1ONEE9PSIsInZhbHVlIjoiWWpaZHNPMXdoNUpYVXdRTy9WL0lGQkVXNzZhME5hZFhYY25hbEY0ZTRzM24zelVQZUppckNXNlBJTUwzdndkQnFtQTRnblRIb3ZxZU94b015TmlrbUhCbnhlU0xkSlJkVm5zR3Fab29KQmhqR0Fja3pweFhGVHRKRm9DSG5KUGEiLCJtYWMiOiI4NTYxNzU4ZmY2NmI4MjkzNmNkNTlmMTc5NzQyNjhjNmI1NDU1YWUxZGUxYTA0MjZiYTg4ODg3Y2MwZmJmNzczIiwidGFnIjoiIn0%3D; medivia_session=eyJpdiI6IkFOQm5hNTNZRjBhQVFibGhlVUd4R3c9PSIsInZhbHVlIjoiallBNDFKbzRMeHFjRjF1cFJBSVljdnFlTkQyRGhpYkd6SnhHU0Q2TkdGTkVta3ExeUlSZlZUNlVwYzFrNTlnK08zR3p3b0RTc1BvN2U0N2p2Ums5K1FyaThrcWtWRHFKT3Q5c0ovSXdSMjYzVjE2MXIxcWxXaklnaHAyNDdYN20iLCJtYWMiOiIyODBlN2E2MDE2OWYwYmQzYWNhOGI3NjAzOTVlNDcyMDk0NzljMTZlMGYwYjYxNjE5MTRiZjkwZTA3ZjEzMzVkIiwidGFnIjoiIn0%3D")
	req.Header.Add("DNT", "1")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Referer", "https://medivia.online/community/character")
	req.Header.Add("Sec-Fetch-Dest", "document")
	req.Header.Add("Sec-Fetch-Mode", "navigate")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("Sec-Fetch-User", "?1")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36")
	req.Header.Add("sec-ch-ua", "Google Chrome;v=111, Not(A:Brand;v=8, Chromium;v=111")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", "Windows")
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

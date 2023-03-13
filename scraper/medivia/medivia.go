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
					tt = tkr.Next()
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
		return nil, fmt.Errorf("fetch info unavailable")
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
							tt = tkr.Next()
							token = tkr.Token()
							prefix := strings.TrimSpace(token.Data)
							tt = tkr.Next()
							tt = tkr.Next()
							token = tkr.Token()
							player := strings.TrimSpace(token.Data)
							tt = tkr.Next()
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
	req.Header.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Add("accept-language", "en-US,en;q=0.9,pt-BR;q=0.8,pt;q=0.7")
	req.Header.Add("cache-control", "max-age=0")
	req.Header.Add("cookie", "_ga=GA1.2.912253875.1677167453; medivia-cookie-info=read; _fbp=fb.1.1677358909786.355079940; _gid=GA1.2.859713168.1678200538; mmo=eyJpdiI6InQ2UVBJT3dSUHNtTW9ZV2x4VFJscmc9PSIsInZhbHVlIjoiT2FFdnFYcitCUFBzdmplUHNqU1RrcFFZNEdKb0ZsYTJhelI3K1M2aGJSNDJjVVF1UkVGWEtwVXd1TUJjcTBkcmF5a3dYQzgxeDZaajBrMy9rZmNxN2c9PSIsIm1hYyI6ImVmYTY0YjI4MTljMWYyYTExNTZkYjkwMzk1MTRhMjE4MmJhM2Q0MGNlNGViMWYxN2RjOTc1YTQzMDIyZmZmMDQiLCJ0YWciOiIifQ%3D%3D; cf_clearance=zOOrOz7OKbQO_0Xyv8HfOCcNt5oK5G5tit.vLeuuZyw-1678449709-0-150; XSRF-TOKEN=eyJpdiI6IkRVYklBdTM0bUtNRm1hcFdzSmsvd0E9PSIsInZhbHVlIjoiRW9Ec1dEZ3JOU0FYMTVTSGpaQkVSZW94aUJpSGNnUFZ3OHZZcmgyV3pLRlc3bmRrSXlHb0l1WFBHMWhMaG5sOUpUNTI4L2E4enJnQUtaMFU2ZFRnRjlqWC9MOFNGM0VDK3dOalV4NGhraWo4MEdKN3RJUlBlT1JqMjZKY3VUWkwiLCJtYWMiOiJiYWNkN2VjZGEwOTBjMDM3YzU4ZDA0YTc0Y2Y3NDk3MGIzZjk4ZWRjMjllZTI2NmNkYzg0M2NlOWFiY2M5YzAxIiwidGFnIjoiIn0%3D; medivia_session=eyJpdiI6IkZ0QkU3RmpERDVIcGhpSWdHd2JKYXc9PSIsInZhbHVlIjoiSVR6WjNpNjVpVC9ScU9uVkxuay8xV0xCU2R0VGhEdFdjMEtNM28wNjFNbFc4ZlU0c0M5NTJrS1pFWXFvYzFLSk9INDU3blJ5NzY2RkpxN1dBc1FJcUNHY013RUNQSDlmaTJIZ1hkdHp2YllHM2N3TE4xV1ZaMGZuUG5nR2twRnoiLCJtYWMiOiIyMDY4MTg5ZTA1OTRkYzFhNjA5Yzg1MmQyNGU1ODM1MDc1Njk4MWJhZjc4ODM5ZGUwNTQzYWJjNDk1YTdlOTllIiwidGFnIjoiIn0%3D")
	req.Header.Add("dnt", "1")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", "Windows")
	req.Header.Add("sec-fetch-dest", "document")
	req.Header.Add("sec-fetch-mode", "navigate")
	req.Header.Add("sec-fetch-site", "none")
	req.Header.Add("sec-fetch-user", "?1")
	req.Header.Add("upgrade-insecure-requests", "1")
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36")
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

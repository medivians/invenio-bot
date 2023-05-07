package medivia

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var ErrPlayerNotFound = errors.New("Player not found")

const (
	login = "AUTH_LOGIN"
	pass  = "AUTH_PASS"
)

type Task struct {
	Kills int    `json:"kills"`
	Name  string `json:"name"`
}

type Kill struct {
	Timestamp   int    `json:"timestamp"`
	Name        string `json:"name"`
	Level       int    `json:"level"`
	Unjustified bool   `json:"unjustified"`
}

func (k Kill) String() string {
	dt := time.Unix(int64(k.Timestamp), 0)
	return fmt.Sprintf("%v - %q(Level: %v) - Unjustified: %v", dt, k.Name, k.Level, strconv.FormatBool(k.Unjustified))
}

type Death struct {
	Active    int `json:"active"`
	Timestamp int `json:"timestamp"`
	Level     int `json:"level"`
	Killers   []*struct {
		Environmental bool   `json:"environmental"`
		Name          string `json:"name"`
	} `json:"killers"`
}

func (k Death) String() string {
	dt := time.Unix(int64(k.Timestamp), 0)
	killers := make([]string, 0, len(k.Killers))
	for _, k := range k.Killers {
		killers = append(killers, k.Name)
	}
	return fmt.Sprintf("%v - at Level: %v by %v", dt, k.Level, killers)
}

type Player struct {
	Name       string   `json:"name"`
	Level      int      `json:"level"`
	World      string   `json:"world"`
	Premium    bool     `json:"premium"`
	Comment    string   `json:"comment"`
	Experience int      `json:"experience"`
	Group      string   `json:"group"`
	Sex        string   `json:"sex"`
	Vocation   string   `json:"vocation"`
	Town       string   `json:"town"`
	Guild      any      `json:"guild"`
	House      any      `json:"house"`
	Bans       any      `json:"bans"`
	Deaths     []*Death `json:"deaths"`
	Kills      []*Kill  `json:"kills"`
	Tasks      []*Task  `json:"tasks"`
}

// Whois DTO used to summary information about player
type WhoIs struct {
	Name       string `json:"Name"`
	Level      int    `json:"Level"`
	World      string `json:"World"`
	Premium    bool   `json:"Premium"`
	Guild      any    `json:"Guild"`
	Comment    string `json:"Comment"`
	Experience int    `json:"Experience"`
	Vocation   string `json:"Vocation"`

	Group string `json:"Group"`
	Town  string `json:"Town"`
	House any    `json:"House"`
	Bans  any    `json:"Bans"`
}

func (w WhoIs) String() string {
	var whois strings.Builder
	whois.WriteString(w.Name)
	whois.WriteString("\n---")
	whois.WriteString(fmt.Sprintf("\nLevel: %v", w.Level))
	whois.WriteString(fmt.Sprintf("\nExp: %v", w.Experience))
	whois.WriteString(fmt.Sprintf("\nVocation: %s", w.Vocation))
	whois.WriteString("\n---")
	whois.WriteString(fmt.Sprintf("\nWorld: %s", w.World))
	whois.WriteString(fmt.Sprintf("\nPremium: %v", w.Premium))
	whois.WriteString(fmt.Sprintf("\nGuild: %v", w.Guild))
	whois.WriteString(fmt.Sprintf("\nGroup: %s", w.Group))
	whois.WriteString("\n---")
	whois.WriteString(fmt.Sprintf("\nTown: %s", w.Town))
	whois.WriteString(fmt.Sprintf("\nHouse: %v", w.House))
	whois.WriteString("\n---")
	whois.WriteString("\nComments")
	whois.WriteString(fmt.Sprintf("\n%v", w.Comment))
	whois.WriteString("\n---")
	whois.WriteString("\nBans")
	whois.WriteString(fmt.Sprintf("\n%v", w.Bans))
	return whois.String()
}

type Medivia struct {
	httpCli *http.Client
}

// New Return new client Medivia
func New() *Medivia {
	return &Medivia{httpCli: &http.Client{Timeout: time.Second * 5}}
}

func (m *Medivia) Player(p string) *Player {
	var payload response
	endpoint := fmt.Sprintf("https://medivia.online/api/public/player/%s", url.PathEscape(p))
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		log.Printf("ERROR building request %q", err)
		return nil
	}
	req.SetBasicAuth(os.Getenv(login), os.Getenv(pass))
	r, err := m.httpCli.Do(req)
	if err != nil {
		log.Printf("ERROR fetching player info %q", err)
		return nil
	}
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("ERROR decoding response %q", err)
		return nil
	}
	return &payload.Player
}

func (m *Medivia) WhoIs(p string) (*WhoIs, error) {
	player := m.Player(p)
	if player == nil {
		return nil, ErrPlayerNotFound
	}
	return &WhoIs{
		Name:       player.Name,
		Level:      player.Level,
		World:      player.World,
		Premium:    player.Premium,
		Comment:    player.Comment,
		Experience: player.Experience,
		Group:      player.Group,
		Vocation:   player.Vocation,
		Town:       player.Town,
		Guild:      player.Guild,
		House:      player.House,
		Bans:       player.Bans,
	}, nil
}

func (m *Medivia) Kills(p string) ([]*Kill, error) {
	player := m.Player(p)
	if player == nil {
		return nil, ErrPlayerNotFound
	}
	return player.Kills, nil
}

func (m *Medivia) Deaths(p string) ([]*Death, error) {
	player := m.Player(p)
	if player == nil {
		return nil, ErrPlayerNotFound
	}
	return player.Deaths, nil
}

type response struct {
	CachedAt int    `json:"cached_at"`
	Player   Player `json:"player"`
}

package discord

import (
	"fmt"
	"log"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"

	"github.com/medivians/invenio-bot/scraper/medivia"
	"github.com/medivians/invenio-bot/scraper/wiki"
)

const token = "TOKEN_BOT"

// Discord size limit to reveice messages
const discordCharactersLimit = 2000

type whoisCli interface {
	WhoIs(p string) (*medivia.WhoIs, error)
}

type pkService interface {
	Kills(p string) ([]*medivia.Kill, error)
	Deaths(p string) ([]*medivia.Death, error)
}

type wikiCli interface {
	WhereToSell(n string) wiki.Locations
}

func Start(whoisCli whoisCli, pk pkService, wikiCli wikiCli) (*discordgo.Session, error) {
	appCommands := []*discordgo.ApplicationCommand{
		{
			Name:        "who-is",
			Description: "returns informations about player name informed",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "player-name",
					Description: "Player name",
					Required:    true,
				},
			},
		},
		{
			Name:        "where-to-sell",
			Description: "return a list of npc where you can sell an item",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "item-name",
					Description: "Item name",
					Required:    true,
				},
			},
		},
		{
			Name:        "kill-list",
			Description: "returns list of players killed by informed player",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "player-name",
					Description: "Player name",
					Required:    true,
				},
			},
		},
		{
			Name:        "death-list",
			Description: "returns list of deaths of the given player",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "player-name",
					Description: "Player name",
					Required:    true,
				},
			},
		},
	}

	commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"who-is": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var playerOpt *discordgo.ApplicationCommandInteractionDataOption
			for _, opt := range i.ApplicationCommandData().Options {
				if opt.Name == "player-name" {
					playerOpt = opt
					break
				}
			}

			whois, err := whoisCli.WhoIs(playerOpt.StringValue())
			if err != nil || whois == nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Player not found!",
					},
				})
				return
			}

			var data strings.Builder
			data.Write([]byte("```"))
			data.WriteString(fmt.Sprintf("%s", whois))
			data.Write([]byte("```"))

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: data.String(),
				},
			})
		},
		"where-to-sell": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var itemOPT *discordgo.ApplicationCommandInteractionDataOption
			for _, opt := range i.ApplicationCommandData().Options {
				if opt.Name == "item-name" {
					itemOPT = opt
					break
				}
			}

			locs := wikiCli.WhereToSell(itemOPT.StringValue())

			var data strings.Builder
			data.Write([]byte("```"))
			if len(locs) == 0 {
				data.Write([]byte("Not found"))
			}
			for i, info := range locs {
				if i%2 != 0 {
					continue
				}
				data.Write([]byte(fmt.Sprintf("%v %v \n", info, locs[i+1])))
			}
			data.Write([]byte("```"))

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: data.String(),
				},
			})
		},
		"kill-list": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var playerOpt *discordgo.ApplicationCommandInteractionDataOption
			for _, opt := range i.ApplicationCommandData().Options {
				if opt.Name == "player-name" {
					playerOpt = opt
					break
				}
			}

			kl, err := pk.Kills(playerOpt.StringValue())
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Player not found!",
					},
				})
				return
			}
			if len(kl) == 0 {
				if err != nil {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Nothing was found!",
						},
					})
					return
				}
			}

			var data strings.Builder
			var rc int
			data.Write([]byte("```"))
			for _, k := range kl {
				ln := fmt.Sprintf("\n%s", k)
				rc += utf8.RuneCountInString(ln)
				if rc > discordCharactersLimit {
					break
				}
				data.WriteString(ln)
			}
			data.Write([]byte("```"))

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Title:   playerOpt.StringValue(),
					Content: data.String(),
				},
			})
		},
		"death-list": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var playerOpt *discordgo.ApplicationCommandInteractionDataOption
			for _, opt := range i.ApplicationCommandData().Options {
				if opt.Name == "player-name" {
					playerOpt = opt
					break
				}
			}

			deaths, err := pk.Deaths(playerOpt.StringValue())
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Player not found!",
					},
				})
				return
			}
			if len(deaths) == 0 {
				if err != nil {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Nothing was found!",
						},
					})
					return
				}
			}

			var data strings.Builder
			data.Write([]byte("```"))
			for i, d := range deaths {
				if i > 0 {
					data.WriteString("\n")
				}
				data.WriteString(fmt.Sprintf("%s", d))
			}
			data.Write([]byte("```"))

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: data.String(),
				},
			})

		},
	}

	cli, err := discordgo.New("Bot " + os.Getenv(token))
	if err != nil {
		return nil, fmt.Errorf("creating bot %w", err)
	}

	cli.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is ready")
	})

	cli.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	if err := cli.Open(); err != nil {
		return nil, fmt.Errorf("opening connection %w", err)
	}

	registeredCommands := make([]*discordgo.ApplicationCommand, len(appCommands))
	for i, v := range appCommands {
		cmd, err := cli.ApplicationCommandCreate(cli.State.User.ID, "", v)
		if err != nil {
			return nil, fmt.Errorf("creating command %v %w", v.Name, err)
		}
		registeredCommands[i] = cmd
	}
	return cli, nil
}

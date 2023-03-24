package discord

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/medivians/invenio-bot/scraper/mediviastats"
	"github.com/medivians/invenio-bot/scraper/wiki"
)

const token = "TOKEN_BOT"

type whoisCli interface {
	WhoIs(n string) mediviastats.Information
}

type killListCli interface {
	KillList(n string) mediviastats.KillList
}

type wikiCli interface {
	WhereToSell(n string) wiki.Locations
}

func Start(whoisCli whoisCli, killList killListCli, wikiCli wikiCli) (*discordgo.Session, error) {
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

			informations := whoisCli.WhoIs(playerOpt.StringValue())
			if len(informations) == 0 {
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
			for i, info := range informations {
				if i%2 != 0 || i == len(informations)-1 {
					continue
				}
				data.Write([]byte(fmt.Sprintf("%v %v \n", info, informations[i+1])))
			}
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

			c := killList.KillList(playerOpt.StringValue())
			if len(c) == 0 {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Nothing was found!",
					},
				})
				return
			}

			var data strings.Builder
			data.Write([]byte("```"))
			data.Write([]byte(strings.Join(c, "\n")))
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

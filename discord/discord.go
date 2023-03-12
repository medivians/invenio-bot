package discord

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/lramosduarte/god-sql/scraper/medivia"
)

const token = "TOKEN_BOT"

type mediviaCli interface {
	WhoIs(n string) (*medivia.Character, error)
}

func Start(medCli mediviaCli) (*discordgo.Session, error) {
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

			c, err := medCli.WhoIs(playerOpt.StringValue())
			if err != nil {
				log.Printf("Error: %v", err)
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
			for i, info := range c.Informations {
				if i%2 != 0 {
					continue
				}
				data.Write([]byte(fmt.Sprintf("%v %v \n", info, c.Informations[i+1])))
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
